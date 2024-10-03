package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"server/controllers"
	"server/dns"
	"server/repos"
	"server/services"

	_ "github.com/mattn/go-sqlite3"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

var (
	// Repos
	userRepo         *repos.UserRepoImpl
	deviceRepo       *repos.DeviceRepoImpl
	domainRepo       *repos.DomainRepoImpl
	categoryRepo     *repos.CategoryRepoImpl
	scheduleRepo     *repos.ScheduleRepoImpl
	domainRuleRepo   *repos.DomainRuleRepoImpl
	categoryRuleRepo *repos.CategoryRuleRepoImpl

	// Services
	contentBlocker *services.ContentBlocker

	// Controllers
	blockedController  *controllers.BlockedController
	dnsController      *controllers.DnsController
	scheduleController *controllers.ScheduleController
)

func initDependencies() {
	// Initialize db
	db, err := sql.Open("sqlite3", "./db/.leo.db")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize repos
	userRepo = repos.NewUserRepo(db)
	deviceRepo = repos.NewDeviceRepo(db)
	domainRepo = repos.NewDomainRepo(db)
	scheduleRepo = repos.NewScheduleRepo(db)
	domainRuleRepo = repos.NewDomainRuleRepo(db)
	categoryRuleRepo = repos.NewCategoryRuleRepo(db)

	// Initialize services
	contentBlocker = services.NewContentBlocker(scheduleRepo)

	// Inizialize controllers
	blockedController = controllers.NewBlockedController()
	dnsController = controllers.NewDnsController(contentBlocker)
	scheduleController = controllers.NewScheduleController(contentBlocker)
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize dependencies
	initDependencies()

	// Start DNS server
	go dns.ListenAndServe()

	// Create a new router using chi
	r := chi.NewRouter()

	// Add some middleware for better logging and recovery
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Route specifically for /dns-query, handling both GET and POST requests
	r.Route("/dns-query", func(r chi.Router) {
		r.Get("/", dnsController.DnsQuery)
		r.Post("/", dnsController.DnsQuery)
	})

	// Serve the blocked page on the root "/"
	r.Get("/", blockedController.BlockedPage)

	// Serve the is blocked endpoint on "/is-blocked"
	r.Get("/is-blocked", scheduleController.IsBlocked)

	// Read port from .env
	port := os.Getenv("PORT")

	// Start server on given port
	log.Printf("Starting Leo on :%s\n", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal(err)
	}
}
