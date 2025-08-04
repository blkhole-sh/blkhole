// Package main provides the Leo DNS content blocker server application.
package main

import (
	"database/sql"
	"embed"
	"encoding/hex"
	"io/fs"
	"log"
	"net/http"
	"os"
	"server/internal/controllers"
	schema "server/internal/db"
	"server/internal/repos"
	"server/internal/services"
	"server/internal/test"

	_ "github.com/mattn/go-sqlite3"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

//go:embed assets
var webFS embed.FS

var (
	// Repos
	userRepo     repos.UserRepo
	deviceRepo   repos.DeviceRepo
	ruleRepo     repos.RuleRepo
	listRepo     repos.ListRepo
	scheduleRepo repos.ScheduleRepo

	// Services
	contentBlocker services.ContentBlocker
	cryptoService  services.CryptoService
	listService    services.ListsService

	// Controllers
	deviceController       controllers.DeviceController
	userController         controllers.UserController
	dnsController          controllers.DNSController
	listController         controllers.ListController
	mobileConfigController controllers.MobileConfigController
	scheduleController     controllers.ScheduleController
	quoteController        controllers.QuoteController
	frontendController     *controllers.FrontendController

	// Test
	t              test.Test
	testController controllers.TestController
)

func initDependencies() {
	// Initialize repos
	db, err := sql.Open("sqlite3", "./internal/db/leo.db")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize db schema
	err = schema.Init(db)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize secret
	secret, err := hex.DecodeString(os.Getenv("LEO_SECRET"))
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Repos
	deviceRepo = repos.NewDeviceRepo(db)
	ruleRepo = repos.NewRuleRepo(db)
	listRepo = repos.NewListRepo(db)
	scheduleRepo = repos.NewScheduleRepo(db)
	userRepo = repos.NewUserRepo(db)

	// Initialize services
	contentBlocker = services.NewContentBlocker(scheduleRepo)
	cryptoService = services.NewCryptoService(secret)
	listService = services.NewListsService(listRepo, ruleRepo)

	// Inizialize controllers
	deviceController = controllers.NewDeviceController(deviceRepo, cryptoService)
	userController = controllers.NewUserController(userRepo, cryptoService)
	dnsController = controllers.NewDNSController(contentBlocker)
	listController = controllers.NewListController(listRepo)
	mobileConfigController = controllers.NewMobileConfigController()
	scheduleController = controllers.NewScheduleController(scheduleRepo, contentBlocker)
	quoteController = controllers.NewQuoteController()
	testController = controllers.NewTestController(t)

	// Initialize frontend controller with embedded assets
	webSubFS, err := fs.Sub(webFS, "assets")
	if err != nil {
		log.Fatal(err)
	}
	frontendController = controllers.NewFrontendController(webSubFS)

	// Initialize test
	t := test.NewTest(userRepo, deviceRepo, ruleRepo, listRepo, listService, scheduleRepo, cryptoService)
	testController = controllers.NewTestController(t)
}

func main() {
	// Initialize dependencies
	initDependencies()

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

		// Test route
		r.Get("/test", testController.RunTest)
	})

	// Serve static files (legacy)
	r.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))).ServeHTTP)

	// Serve frontend - MUST BE LAST to avoid catching API routes
	r.Get("/*", frontendController.Serve)

	// Read port from .env
	port := os.Getenv("LEO_PORT")

	// Start server on given port
	log.Printf("Starting Leo on :%s\n", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal(err)
	}
}
