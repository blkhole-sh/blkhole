// Package main provides the blkhole server application.
package main

import (
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
	"path/filepath"

	"github.com/lemon3studio/blkhole/internal/controllers"
	schema "github.com/lemon3studio/blkhole/internal/db"
	"github.com/lemon3studio/blkhole/internal/middleware"
	"github.com/lemon3studio/blkhole/internal/repos"
	"github.com/lemon3studio/blkhole/internal/services"
	"github.com/lemon3studio/blkhole/internal/test"

	_ "github.com/mattn/go-sqlite3"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/acme/autocert"
)

var devMode = "false" // Set via ldflags

// Config defines the blkhole configuration
type Config struct {
	Domain      string
	Port        string
	UpstreamDNS string
	Secret      string
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
	cryptoService  services.CryptoService
	listService    services.ListsService
	authService    services.AuthService
	tokenAuth      *jwtauth.JWTAuth

	// Controllers
	deviceController       controllers.DeviceController
	userController         controllers.UserController
	dnsController          controllers.DNSController
	listController         controllers.ListController
	mobileConfigController controllers.MobileConfigController
	scheduleController     controllers.ScheduleController
	quoteController        controllers.QuoteController
	authController         controllers.AuthController
	webController          controllers.WebController

	// Test
	t              test.Test
	testController controllers.TestController
)

func initConfig() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.Port, "p", "", "HTTP port (local mode only, mutually exclusive with -d)")
	flag.StringVar(&cfg.Domain, "d", "", "Domain for HTTPS with autocert (production mode, mutually exclusive with -p)")
	flag.StringVar(&cfg.UpstreamDNS, "u", "1.1.1.1:53", "Upstream DNS server")
	flag.StringVar(&cfg.Secret, "s", "", "JWT secret (hex)")

	flag.Parse()

	// Validate mode flags
	if cfg.Domain != "" && cfg.Port != "" {
		return nil, fmt.Errorf("-p has no effect when -d is set")
	}
	if cfg.Domain == "" && cfg.Port == "" {
		return nil, fmt.Errorf("either -p (local mode) or -d (production mode) is required")
	}

	// Validate common required flags
	var missing []string
	if cfg.UpstreamDNS == "" {
		missing = append(missing, "-u (upstream dns server)")
	}
	if cfg.Secret == "" {
		missing = append(missing, "-s (secret)")
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

func initDependencies(cfg *Config) {
	// Get database path
	configDir, _ := os.UserConfigDir()
	dbPath := filepath.Join(configDir, "blkhole", "blkhole.db")
	os.MkdirAll(filepath.Dir(dbPath), 0o755)

	// Initialize repos
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize db schema
	err = schema.Init(db)
	if err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}

	// Initialize secret
	secret, err := hex.DecodeString(cfg.Secret)
	if err != nil {
		log.Fatalf("failed to decode secret: %v", err)
	}

	// Initialize upstream dns server
	upstreamDNS, err := initUpstreamDNS(cfg.UpstreamDNS)
	if err != nil {
		log.Fatalf("failed to initialize upstream dns server: %v", err)
	}

	// Initialize repos
	devices = repos.NewDeviceRepo(db)
	rules = repos.NewRuleRepo(db)
	lists = repos.NewListRepo(db)
	schedules = repos.NewScheduleRepo(db)
	users = repos.NewUserRepo(db)
	domains = repos.NewDomainRepo(db)

	// Initialize services
	contentBlocker = services.NewContentBlocker(devices, rules, schedules, domains)
	cryptoService = services.NewCryptoService(secret)
	listService = services.NewListsService(lists, rules, domains)

	// Initialize auth service with token auth
	tokenAuth = jwtauth.New("HS256", secret, nil)
	authService = services.NewAuthService(users, cryptoService, tokenAuth)

	// Initialize controllers
	deviceController = controllers.NewDeviceController(devices, schedules, cryptoService)
	userController = controllers.NewUserController(users, cryptoService)
	dnsController = controllers.NewDNSController(contentBlocker, upstreamDNS, cfg.Domain)
	listController = controllers.NewListController(lists)
	mobileConfigController = controllers.NewMobileConfigController(cfg.Domain, devices)
	scheduleController = controllers.NewScheduleController(schedules, devices, lists, contentBlocker)
	quoteController = controllers.NewQuoteController()
	authController = controllers.NewAuthController(authService)
	testController = controllers.NewTestController(t)

	// Initialize frontend controller with embedded assets
	webSubFS, err := fs.Sub(webFS, "static")
	if err != nil {
		log.Fatal(err)
	}
	webController = controllers.NewWebController(webSubFS)

	// Initialize test
	t := test.NewTest(users, devices, rules, lists, listService, schedules, cryptoService, domains)
	testController = controllers.NewTestController(t)
}

func initRouter() *chi.Mux {
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

	// Route specifically for /dns-query, handling both GET and POST requests
	r.Route("/{deviceHash}/dns-query", func(r chi.Router) {
		r.Get("/", dnsController.DNSQuery)
		r.Post("/", dnsController.DNSQuery)
	})

	// API routes group
	r.Route("/api", func(r chi.Router) {
		// Apply JSON middleware to all API routes
		r.Use(middleware.JSONMiddleware)

		// Public routes
		r.Post("/auth/login", authController.Login)
		r.Post("/auth/refresh", authController.RefreshToken)
		r.Post("/auth/logout", authController.Logout)
		r.Get("/quote", quoteController.Random)

		// Protected routes
		r.Group(func(r chi.Router) {
			// Cookie-based JWT authentication middleware
			r.Use(middleware.CookieAuthenticator(tokenAuth))

			// Auth routes
			r.Get("/auth/me", authController.GetCurrentUser)

			// User API routes
			r.Get("/users/{id}", userController.FindByID)
			r.Put("/users", userController.Create)
			r.Patch("/users/{id}", userController.Update)
			r.Delete("/users/{id}", userController.Delete)

			// Device API routes
			r.Get("/devices/{id}", deviceController.FindByID)
			r.Get("/devices/{id}/config", mobileConfigController.GenerateConfig)
			r.Get("/users/{userId}/devices", deviceController.FindByUser)
			r.Put("/devices", deviceController.Create)
			r.Patch("/devices/{id}", deviceController.Update)
			r.Delete("/devices/{id}", deviceController.Delete)

			// List API routes
			r.Get("/lists/{id}", listController.FindByID)
			r.Get("/users/{userId}/lists", listController.FindByUser)
			r.Put("/lists", listController.Create)
			r.Patch("/lists/{id}", listController.Update)
			r.Delete("/lists/{id}", listController.Delete)

			// Schedule API routes
			r.Get("/schedules/{id}", scheduleController.FindByID)
			r.Get("/users/{userId}/schedules", scheduleController.FindByUser)
			r.Put("/schedules", scheduleController.Create)
			r.Patch("/schedules/{id}", scheduleController.Update)
			r.Delete("/schedules/{id}", scheduleController.Delete)
			r.Get("/is-blocked", scheduleController.IsBlocked)
		})
	})

	// Serve test route
	r.Get("/test", testController.RunTest)

	// Serve static files (legacy)
	r.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))).ServeHTTP)

	if devMode == "true" {
		log.Printf("dev mode enabled")
	} else {
		// Serve frontend - only in production mode
		// Serve frontend with compression middleware
		// Must be last to avoid catching API routes
		webSubFS, _ := fs.Sub(webFS, "static")
		r.With(middleware.CompressionMiddleware(webSubFS)).Get("/*", webController.Serve)
	}

	return r
}

func startServer(cfg *Config, handler http.Handler) error {
	// Local mode — plain HTTP on specified port
	if cfg.Domain == "" {
		log.Printf("starting blkhole on :%s", cfg.Port)
		return http.ListenAndServe(":"+cfg.Port, handler)
	}

	// Store certs alongside the database in the user config directory
	configDir, _ := os.UserConfigDir()
	m := &autocert.Manager{
		Cache:      autocert.DirCache(filepath.Join(configDir, "blkhole", "certs")),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.Domain),
	}

	// HTTP server on :80 for ACME challenge and redirect to HTTPS
	go func() {
		log.Printf("starting HTTP redirect on :80")
		if err := http.ListenAndServe(":80", m.HTTPHandler(nil)); err != nil {
			log.Fatalf("HTTP redirect server error: %v", err)
		}
	}()

	log.Printf("starting blkhole on :443 (domain: %s)", cfg.Domain)
	return http.Serve(m.Listener(), handler)
}

func main() {
	// Initialize config
	cfg, err := initConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Initialize dependencies
	initDependencies(cfg)

	// Initialize content blocker
	if err = contentBlocker.Init(); err != nil {
		log.Fatalf("failed to initialize content blocker: %v", err)
	}

	// Initialize router
	r := initRouter()

	// Start server
	if err = startServer(cfg, r); err != nil {
		log.Fatalf("failed to start server :%v", err)
	}
}
