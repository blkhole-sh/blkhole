// Package main provides the Leo DNS content blocker server application.
package main

import (
	"database/sql"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"server/controllers"
	schema "server/db"
	"server/repos"
	"server/services"
	"server/test"

	_ "github.com/mattn/go-sqlite3"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

var (
	// Repos
	userRepo     repos.UserRepo
	deviceRepo   repos.DeviceRepo
	domainRepo   repos.DomainRepo
	listRepo     repos.ListRepo
	scheduleRepo repos.ScheduleRepo

	// Services
	contentBlocker services.ContentBlocker
	cryptoService  services.CryptoService

	// Controllers
	deviceController   controllers.DeviceController
	userController     controllers.UserController
	dnsController      controllers.DnsController
	listController     controllers.ListController
	scheduleController controllers.ScheduleController
	quoteController    controllers.QuoteController

	// Test
	t test.Test
)

func initDependencies() {
	// Initialize repos
	db, err := sql.Open("sqlite3", "./db/leo.db")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize db schema
	err = schema.Init(db)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize secret
	secret, err := hex.DecodeString(os.Getenv("SECRET"))
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Repos
	deviceRepo = repos.NewDeviceRepo(db)
	domainRepo = repos.NewDomainRepo(db)
	listRepo = repos.NewListRepo(db)
	scheduleRepo = repos.NewScheduleRepo(db)
	userRepo = repos.NewUserRepo(db)

	// Initialize services
	contentBlocker = services.NewContentBlocker(scheduleRepo)
	cryptoService = services.NewCryptoService(secret)

	// Inizialize controllers
	deviceController = controllers.NewDeviceController(deviceRepo, cryptoService)
	userController = controllers.NewUserController(userRepo, cryptoService)
	dnsController = controllers.NewDnsController(contentBlocker)
	listController = controllers.NewListController(listRepo)
	scheduleController = controllers.NewScheduleController(scheduleRepo, contentBlocker)
	quoteController = controllers.NewQuoteController()

	// Initilaize test

	// t = test.NewTest(userRepo, deviceRepo, domainRepo, listRepo, scheduleRepo, cryptoService)
	// if err := t.Test(); err != nil {
	// log.Fatal(err)
	//}
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize dependencies
	initDependencies()

	// Create a new router using chi
	r := chi.NewRouter()

	// Add some middleware for better logging and recovery
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Configure CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Change this to your frontend URL
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Max cache time for preflight requests
	}))

	// Serve static files
	r.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))).ServeHTTP)

	// Route specifically for /dns-query, handling both GET and POST requests
	r.Route("/{userHash}/{deviceHash}/dns-query", func(r chi.Router) {
		r.Get("/", dnsController.DnsQuery)
		r.Post("/", dnsController.DnsQuery)
	})

	// Serve the blocked page on "/blocked"
	// r.Get("/blocked", blockedController.BlockedPage)
	// Servce user api routes
	r.Get("/users/{hash}", userController.FindByHash)
	r.Put("/users", userController.Create)
	r.Post("/users/{hash}", userController.Update)
	r.Delete("/users/{hash}", userController.Delete)

	// Serve device api routes
	r.Get("/devices/{hash}", deviceController.FindByHash)
	r.Get("/users/{userHash}/devices", deviceController.FindByUser)
	r.Put("/devices", deviceController.Create)
	r.Post("/devices/{hash}", deviceController.Update)
	r.Delete("/devices/{hash}", deviceController.Delete)

	// Serve list api routes
	r.Get("/lists/{id}", listController.FindByID)
	r.Get("/users/{userHash}/lists", listController.FindByUser)
	r.Put("/lists", listController.Create)
	r.Post("/lists/{id}", listController.Update)
	r.Delete("/lists/{id}", listController.Delete)

	// Serve schedule api routes
	r.Get("/schedules/{id}", scheduleController.FindByID)
	r.Get("/users/{userHash}/schedules", scheduleController.FindByUser)
	r.Put("/schedules", scheduleController.Create)
	r.Post("/schedules/{id}", scheduleController.Update)
	r.Get("/is-blocked", scheduleController.IsBlocked)

	r.Get("/quote", quoteController.Random)

	// Read port from .env
	port := os.Getenv("PORT")

	// Start server on given port
	log.Printf("Starting Leo on :%s\n", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal(err)
	}
}
