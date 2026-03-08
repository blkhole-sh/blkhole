// Package main provides the blkhole server application.
package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"embed"
	"encoding/hex"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/lemon3studio/blkhole/internal/cache"
	"github.com/lemon3studio/blkhole/internal/controllers"
	schema "github.com/lemon3studio/blkhole/internal/db"
	"github.com/lemon3studio/blkhole/internal/repos"
	"github.com/lemon3studio/blkhole/internal/routes"
	"github.com/lemon3studio/blkhole/internal/servers"
	"github.com/lemon3studio/blkhole/internal/services"
	"github.com/lemon3studio/blkhole/internal/test"
	"golang.org/x/sync/errgroup"

	_ "github.com/mattn/go-sqlite3"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
)

var devMode = "false" // Set via ldflags

// Config defines the blkhole configuration
type Config struct {
	Domain      string
	Port        string
	UpstreamDNS string
	Secret      string
	TLSConfig   *tls.Config
}

var (
	// Static file server
	//go:embed static
	webFS embed.FS

	// Repos
	users     repos.UserRepo
	devices   repos.DeviceRepo
	rules     repos.RuleRepo
	lists     repos.ListRepo
	schedules repos.ScheduleRepo
	domains   repos.DomainRepo

	// Services
	contentBlocker services.ContentBlocker
	resolver       services.Resolver
	cryptoService  services.CryptoService
	listService    services.ListsService
	authService    services.AuthService
	tokenAuth      *jwtauth.JWTAuth

	// Caches
	statsCache  cache.StatsCache
	deviceCache cache.DeviceCache

	// Controllers
	deviceController       controllers.DeviceController
	userController         controllers.UserController
	dohController          controllers.DoHController
	listController         controllers.ListController
	mobileConfigController controllers.MobileConfigController
	scheduleController     controllers.ScheduleController
	quoteController        controllers.QuoteController
	authController         controllers.AuthController
	statsController        controllers.StatsController
	webController          controllers.WebController

	// Test
	t              test.Test
	testController controllers.TestController
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func initConfig() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.Port, "p", envOrDefault("BLKHOLE_PORT", ""), "HTTP port (local mode only, mutually exclusive with -d)")
	flag.StringVar(&cfg.Domain, "d", envOrDefault("BLKHOLE_DOMAIN", "localhost"), "Domain for HTTPS with autocert (production mode, mutually exclusive with -p)")
	flag.StringVar(&cfg.UpstreamDNS, "u", envOrDefault("BLKHOLE_UPSTREAM_DNS", "1.1.1.1:53"), "Upstream DNS server")
	flag.StringVar(&cfg.Secret, "s", envOrDefault("BLKHOLE_SECRET", ""), "JWT secret (hex)")

	flag.Parse()

	// Validate mode flags
	if cfg.Domain == "" && cfg.Port == "" {
		return nil, fmt.Errorf("either -p (local mode) or -d (production mode) is required")
	}

	// Validate common required flags
	var missing []string
	if cfg.UpstreamDNS == "" {
		missing = append(missing, "-u / BLKHOLE_UPSTREAM_DNS")
	}
	if cfg.Secret == "" {
		missing = append(missing, "-s / BLKHOLE_SECRET")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required flags: %v", missing)
	}

	return &cfg, nil
}

func initUpstreamDNS(addr string) (string, error) {
	// Parse the address to validate format
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", fmt.Errorf("invalid upstream DNS format (expected host:port): %w", err)
	}

	// Validate IP address
	if ip := net.ParseIP(host); ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", host)
	}

	// Validate port
	if _, err := net.LookupPort("tcp", port); err != nil {
		return "", fmt.Errorf("invalid port: %s", port)
	}

	return addr, nil
}

func initDatabase() *sql.DB {
	configDir, _ := os.UserConfigDir()
	dbPath := filepath.Join(configDir, "blkhole", "blkhole.db")
	os.MkdirAll(filepath.Dir(dbPath), 0o755)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	err = schema.Init(db)
	if err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}

	return db
}

func initSecret(secretHex string) []byte {
	secret, err := hex.DecodeString(secretHex)
	if err != nil {
		log.Fatalf("failed to decode secret: %v", err)
	}
	return secret
}

func initRepos(db *sql.DB) {
	devices = repos.NewDeviceRepo(db)
	rules = repos.NewRuleRepo(db)
	lists = repos.NewListRepo(db)
	schedules = repos.NewScheduleRepo(db)
	users = repos.NewUserRepo(db)
	domains = repos.NewDomainRepo(db)
}

func initCaches() {
	deviceCache = cache.NewDeviceCache()
	statsCache = cache.NewStatsCache(deviceCache)
}

func initServices(secret []byte, upstreamDNS string) {
  contentBlocker = services.NewContentBlocker(devices, rules, schedules, domains, deviceCache)
	resolver = services.NewResolver(contentBlocker, statsCache, upstreamDNS)
	cryptoService = services.NewCryptoService(secret)
	listService = services.NewListsService(lists, rules, domains)

	tokenAuth = jwtauth.New("HS256", secret, nil)
	authService = services.NewAuthService(users, cryptoService, tokenAuth)
}

func initControllers(domain, upstreamDNS string) {
	deviceController = controllers.NewDeviceController(devices, schedules, cryptoService)
	userController = controllers.NewUserController(users, cryptoService)
	dohController = controllers.NewDoHController(resolver, upstreamDNS, domain, statsCache)
	listController = controllers.NewListController(lists)
	mobileConfigController = controllers.NewMobileConfigController(domain, devices, authService)
	scheduleController = controllers.NewScheduleController(schedules, devices, lists, contentBlocker)
	quoteController = controllers.NewQuoteController()
	authController = controllers.NewAuthController(authService)
	statsController = controllers.NewStatsController(statsCache, devices)
}

func initWebAndTest(db *sql.DB) {
	// Initialize frontend controller with embedded assets
	webSubFS, err := fs.Sub(webFS, "static")
	if err != nil {
		log.Fatal(err)
	}
	webController = controllers.NewWebController(webSubFS)

	// Initialize test
	t := test.NewTest(users, devices, rules, lists, listService, schedules, cryptoService, domains, statsCache, deviceCache)
	testController = controllers.NewTestController(t)
}

func initDependencies(cfg *Config) {
	db := initDatabase()
	secret := initSecret(cfg.Secret)
	upstreamDNS, err := initUpstreamDNS(cfg.UpstreamDNS)
	if err != nil {
		log.Fatalf("failed to initialize upstream dns server: %v", err)
	}

	initRepos(db)
	initCaches()
	initServices(secret, upstreamDNS)
	initControllers(cfg.Domain, upstreamDNS)
	initWebAndTest(db)
}

func initRouter(domain string) *chi.Mux {
	// Create a new router using chi
	r := chi.NewRouter()

	// Add some middleware for better logging and recovery
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	// Configure CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   map[bool][]string{true: {"http://localhost:5173"}, false: {"*"}}[devMode == "true"],
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Max cache time for preflight requests
	}))

	// Initialize routes
	routes.InitAPI(r, authController, userController, deviceController, listController, scheduleController, statsController, quoteController, mobileConfigController, testController, webController, tokenAuth)
	routes.InitWeb(r, webController, &webFS, devMode)
	routes.InitDoH(r, dohController, domain, devMode)

	return r
}

func main() {
	// Initialize config
	cfg, err := initConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	log.Println("inizializing blkhole ...")

	// Initialize dependencies
	initDependencies(cfg)

	// Initialize content blocker
	if err = contentBlocker.Init(); err != nil {
		log.Fatalf("failed to initialize content blocker: %v", err)
	}

	// Start background tasks
	statsCache.Start()

	// Initialize HTTP router
	r := initRouter(cfg.Domain)

	// Initialize context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize error group
	g, ctx := errgroup.WithContext(ctx)

	log.Println("supernova detected, collapsing star in progress ...")

	// Start servers
	g.Go(func() error {
		err := servers.StartHTTP(ctx, r, cfg.Domain, cfg.Port, cfg.TLSConfig)
		if err != nil {
			log.Printf("http server error: %v", err)
		}
		return err
	})

	if cfg.TLSConfig != nil {
		g.Go(func() error {
			err := servers.StartDoT(ctx, resolver, cfg.TLSConfig)
			if err != nil {
				log.Printf("dot server error: %v", err)
			}

			return err
		})
	}

	// Catch server errors
	if err := g.Wait(); err != nil && ctx.Err() == nil {
		log.Fatalf("server error: %v", err)
	}
}
