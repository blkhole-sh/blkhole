// Package main provides the Leo  server application.
package main

import (
	"database/sql"
	"embed"
	"encoding/hex"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/lemon3studio/leo/internal/controllers"
	schema "github.com/lemon3studio/leo/internal/db"
	"github.com/lemon3studio/leo/internal/repos"
	"github.com/lemon3studio/leo/internal/services"
	"github.com/lemon3studio/leo/internal/test"

	_ "modernc.org/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
)

// Config defines the Leo configuration
type Config struct {
	Domain string
	Port   string
	Secret string
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

	flag.StringVar(&cfg.Port, "p", "", "Server port")
	flag.StringVar(&cfg.Domain, "d", "", "Server domain")
	flag.StringVar(&cfg.Secret, "s", "", "JWT secret (hex)")

	flag.Parse()

	// Validate required flags
	var missing []string
	if cfg.Port == "" {
		missing = append(missing, "-p (port)")
	}
	if cfg.Domain == "" {
		missing = append(missing, "-d (domain)")
	}
	if cfg.Secret == "" {
		missing = append(missing, "-s (secret)")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required flags: %v", missing)
	}

	return &cfg, nil
}

func initDependencies(cfg *Config) {
	// Get database path
	configDir, _ := os.UserConfigDir()
	dbPath := filepath.Join(configDir, "leo", "leo.db")
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	// Initialize repos
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize db schema
	err = schema.Init(db)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize secret
	secret, err := hex.DecodeString(cfg.Secret)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize repos
	devices = repos.NewDeviceRepo(db)
	rules = repos.NewRuleRepo(db)
	lists = repos.NewListRepo(db)
	schedules = repos.NewScheduleRepo(db)
	users = repos.NewUserRepo(db)

	// Initialize services
	contentBlocker = services.NewContentBlocker(schedules)
	cryptoService = services.NewCryptoService(secret)
	listService = services.NewListsService(lists, rules)

	// Initialize auth service with token auth
	tokenAuth = jwtauth.New("HS256", secret, nil)
	authService = services.NewAuthService(users, cryptoService, tokenAuth)

	// Initialize controllers
	deviceController = controllers.NewDeviceController(devices, cryptoService)
	userController = controllers.NewUserController(users, cryptoService)
	dnsController = controllers.NewDNSController(contentBlocker)
	listController = controllers.NewListController(lists)
	mobileConfigController = controllers.NewMobileConfigController(cfg.Domain)
	scheduleController = controllers.NewScheduleController(schedules, contentBlocker)
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
	t := test.NewTest(users, devices, rules, lists, listService, schedules, cryptoService)
	testController = controllers.NewTestController(t)
}

func initRouter() *chi.Mux {
	// Create a new router using chi
	r := chi.NewRouter()

	// Add some middleware for better logging and recovery
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Configure CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all origins for single binary
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Max cache time for preflight requests
	}))

	// Route specifically for /dns-query, handling both GET and POST requests
	r.Route("/{userHash}/{deviceHash}/dns-query", func(r chi.Router) {
		r.Get("/", dnsController.DNSQuery)
		r.Post("/", dnsController.DNSQuery)
	})

	// Serve mobile config controller (needs to be at root level)
	r.Get("/config/{userHash}/{deviceHash}", mobileConfigController.GenerateConfig)

	// API routes group
	r.Route("/api", func(r chi.Router) {
		// Public auth routes
		r.Post("/auth/login", authController.Login)
		r.Get("/api/test", testController.RunTest)

		// Protected routes
		r.Group(func(r chi.Router) {
			// JWT authentication middleware
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))

			// Auth routes
			r.Get("/auth/me", authController.GetCurrentUser)

			// User API routes
			r.Get("/users/{hash}", userController.FindByHash)
			r.Put("/users", userController.Create)
			r.Post("/users/{hash}", userController.Update)
			r.Delete("/users/{hash}", userController.Delete)

			// Device API routes
			r.Get("/devices/{hash}", deviceController.FindByHash)
			r.Get("/users/{userHash}/devices", deviceController.FindByUser)
			r.Put("/devices", deviceController.Create)
			r.Post("/devices/{hash}", deviceController.Update)
			r.Delete("/devices/{hash}", deviceController.Delete)

			// List API routes
			r.Get("/lists/{id}", listController.FindByID)
			r.Get("/users/{userHash}/lists", listController.FindByUser)
			r.Put("/lists", listController.Create)
			r.Post("/lists/{id}", listController.Update)
			r.Delete("/lists/{id}", listController.Delete)

			// Schedule API routes
			r.Get("/schedules/{id}", scheduleController.FindByID)
			r.Get("/users/{userHash}/schedules", scheduleController.FindByUser)
			r.Put("/schedules", scheduleController.Create)
			r.Post("/schedules/{id}", scheduleController.Update)
			r.Get("/is-blocked", scheduleController.IsBlocked)

			// Quote API route
			r.Get("/quote", quoteController.Random)
		})
	})

	// Serve static files (legacy)
	r.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))).ServeHTTP)

	// Serve frontend - MUST BE LAST to avoid catching API routes
	r.Get("/*", webController.Serve)

	return r
}

func main() {
	// Initialize config
	cfg, err := initConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize dependencies
	initDependencies(cfg)

	// Initialize router
	r := initRouter()

	// Start server on given port
	log.Printf("starting leo on :%s", cfg.Port)
	err = http.ListenAndServe(":"+cfg.Port, r)
	if err != nil {
		log.Fatal(err)
	}
}
