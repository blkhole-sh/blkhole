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
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/blkhole-sh/blkhole/internal/cache"
	"github.com/blkhole-sh/blkhole/internal/controllers"
	schema "github.com/blkhole-sh/blkhole/internal/db"
	"github.com/blkhole-sh/blkhole/internal/repos"
	"github.com/blkhole-sh/blkhole/internal/routes"
	"github.com/blkhole-sh/blkhole/internal/servers"
	"github.com/blkhole-sh/blkhole/internal/services"
	"golang.org/x/crypto/acme/autocert"
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
	queryLogs repos.QueryLogRepo

	// Services
	contentBlocker services.ContentBlocker
	resolver       services.Resolver
	cryptoService  services.CryptoService
	listService    services.ListService
	authService    services.AuthService
	queryLogBuffer *services.QueryLogBuffer
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
	settingsController     controllers.SettingsController
	webController          controllers.WebController
	queryLogController     controllers.QueryLogController
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
	flag.StringVar(&cfg.UpstreamDNS, "u", envOrDefault("BLKHOLE_UPSTREAM_DNS", "9.9.9.9:53"), "Upstream DNS server")
	flag.StringVar(&cfg.Secret, "s", envOrDefault("BLKHOLE_SECRET", ""), "JWT secret (hex)")

	flag.Parse()

	// Validate mode flags
	if cfg.Domain == "" && cfg.Port == "" {
		return nil, fmt.Errorf("either -p (local mode) or -d (production mode) is required")
	}
	if cfg.Port != "" && cfg.Domain != "" && cfg.Domain != "localhost" {
		return nil, fmt.Errorf("-p (local mode) and -d (production mode) are mutually exclusive")
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
	queryLogs = repos.NewQueryLogRepo(db)
}

func initCaches() {
	deviceCache = cache.NewDeviceCache()
	statsCache = cache.NewStatsCache(deviceCache)
}

func initServices(secret []byte, upstreamDNS string) {
	contentBlocker = services.NewContentBlocker(devices, rules, schedules, domains, deviceCache)
	queryLogBuffer = services.NewQueryLogBuffer(queryLogs)
	resolver = services.NewResolver(contentBlocker, statsCache, deviceCache, upstreamDNS, queryLogBuffer)
	cryptoService = services.NewCryptoService(secret)
	listService = services.NewListService(lists, rules, domains, contentBlocker)

	tokenAuth = jwtauth.New("HS256", secret, nil)
	authService = services.NewAuthService(users, schedules, cryptoService, tokenAuth)
}

func initControllers(domain, upstreamDNS string) {
	// Production mode serves over HTTPS, so auth cookies must be Secure
	secureCookies := domain != "" && domain != "localhost"

	deviceController = controllers.NewDeviceController(devices, schedules, cryptoService, authService)
	userController = controllers.NewUserController(users, authService, cryptoService)
	dohController = controllers.NewDoHController(resolver, upstreamDNS, domain, statsCache)
	listController = controllers.NewListController(lists, listService, authService)
	mobileConfigController = controllers.NewMobileConfigController(domain, devices, authService)
	scheduleController = controllers.NewScheduleController(schedules, devices, lists, contentBlocker, authService)
	quoteController = controllers.NewQuoteController()
	authController = controllers.NewAuthController(authService, listService, secureCookies)
	statsController = controllers.NewStatsController(statsCache, devices, queryLogs, authService)
	settingsController = controllers.NewSettingsController(upstreamDNS)
	queryLogController = controllers.NewQueryLogController(queryLogs, authService)
}

func initWeb() {
	// Initialize frontend controller with embedded assets
	webSubFS, err := fs.Sub(webFS, "static")
	if err != nil {
		log.Fatal(err)
	}
	webController = controllers.NewWebController(webSubFS)
}

func initTLS(domain string) *tls.Config {
	// Only initialize TLS for production mode (when domain is set and not localhost)
	if domain == "" || domain == "localhost" {
		return nil
	}

	configDir, _ := os.UserConfigDir()
	certDir := filepath.Join(configDir, "blkhole", "certs")
	log.Printf("TLS: using certificate cache directory: %s", certDir)
	log.Printf("TLS: allowing root domain and single-level subdomains for: %s", domain)

	hostPolicy := func(ctx context.Context, host string) error {
		if host == domain {
			return nil
		}
		suffix := "." + domain
		if strings.HasSuffix(host, suffix) {
			prefix := host[:len(host)-len(suffix)]
			if !strings.Contains(prefix, ".") {
				return nil
			}
		}
		return fmt.Errorf("host %q not authorized for TLS certificate", host)
	}

	m := &autocert.Manager{
		Cache:      autocert.DirCache(certDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hostPolicy,
	}

	// Start ACME HTTP Challenge server on :80 for Let's Encrypt certificate verification
	go func() {
		log.Println("starting acme http server on :80")
		if err := http.ListenAndServe(":80", m.HTTPHandler(nil)); err != nil {
			log.Printf("acme http server error: %v", err)
		}
	}()

	// Wrap TLS config to log certificate errors
	tlsConfig := m.TLSConfig()
	originalGetCert := tlsConfig.GetCertificate
	tlsConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		log.Printf("TLS: GetCertificate called for domain: %s", hello.ServerName)
		cert, err := originalGetCert(hello)
		if err != nil {
			log.Printf("TLS: GetCertificate error for %s: %v", hello.ServerName, err)
			return nil, err
		}
		log.Printf("TLS: GetCertificate success for %s", hello.ServerName)
		return cert, nil
	}

	return tlsConfig
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
	initWeb()
}

func initRouter(cfg *Config) *chi.Mux {
	// Create a new router using chi
	r := chi.NewRouter()

	// Add some middleware for better logging and recovery
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	// Determine allowed origins based on mode
	var allowedOrigins []string
	if devMode == "true" {
		allowedOrigins = []string{"http://localhost:5173"}
	} else {
		if cfg.Domain != "" && cfg.Domain != "localhost" {
			allowedOrigins = []string{"https://" + cfg.Domain}
		} else if cfg.Port != "" {
			allowedOrigins = []string{"http://localhost:" + cfg.Port}
		}
	}

	// Configure CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Max cache time for preflight requests
	}))

	// Initialize routes
	routes.InitAPI(r, authController, userController, deviceController, listController, scheduleController, statsController, quoteController, mobileConfigController, settingsController, webController, queryLogController, tokenAuth)
	routes.InitWeb(r, webController, &webFS, devMode)
	routes.InitDoH(r, dohController, cfg.Domain, devMode)

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

	// Initialize TLS configuration (for production mode with domain)
	cfg.TLSConfig = initTLS(cfg.Domain)

	// Initialize content blocker
	if err = contentBlocker.Init(); err != nil {
		log.Fatalf("failed to initialize content blocker: %v", err)
	}

	// Seed default lists and schedules for existing users in the background
	go func() {
		ids, err := users.FindAllIDs()
		if err != nil {
			log.Printf("failed to load user IDs for default seeding: %v", err)
			return
		}
		for _, id := range ids {
			if err := listService.SeedDefaults(id); err != nil {
				log.Printf("failed to seed default lists for user %d: %v", id, err)
			}
			if err := schedules.SeedDefaults(id); err != nil {
				log.Printf("failed to seed default schedule for user %d: %v", id, err)
			}
		}
	}()

	// Start background tasks
	statsCache.Start()

	// Prune query logs older than 90 days, repeating daily
	go func() {
		for {
			if err := queryLogs.DeleteOlderThan(90); err != nil {
				log.Printf("failed to prune old query logs: %v", err)
			}
			time.Sleep(24 * time.Hour)
		}
	}()

	// Initialize HTTP router
	r := initRouter(cfg)

	// Initialize context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	queryLogBuffer.Start(ctx)

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
			err := servers.StartDoT(ctx, resolver, cfg.Domain, cfg.TLSConfig)
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
