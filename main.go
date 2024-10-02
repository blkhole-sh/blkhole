package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"server/repo"
	"server/view"

	_ "github.com/mattn/go-sqlite3"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

var (
	userRepo         *repo.UserRepoImpl
	deviceRepo       *repo.DeviceRepoImpl
	domainRepo       *repo.DomainRepoImpl
	categoryRepo     *repo.CategoryRepoImpl
	scheduleRepo     *repo.ScheduleRepoImpl
	domainRuleRepo   *repo.DomainRuleRepoImpl
	categoryRuleRepo *repo.CategoryRuleRepoImpl
)

func initDependencies() {
	// Initialize db
	db, err := sql.Open("sqlite3", "./db/.leo.db")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize repos
	userRepo = repo.NewUserRepo(db)
	deviceRepo = repo.NewDeviceRepo(db)
	domainRepo = repo.NewDomainRepo(db)
	scheduleRepo = repo.NewScheduleRepo(db)
	domainRuleRepo = repo.NewDomainRuleRepo(db)
	categoryRuleRepo = repo.NewCategoryRuleRepo(db)
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Start DNS server
	// go dns.ListenAndServe()

	// Create a new router using chi
	r := chi.NewRouter()

	// Add some middleware for better logging and recovery
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Route specifically for /dns-query, handling both GET and POST requests
	r.Route("/dns-query", func(r chi.Router) {
		r.Get("/", view.DnsQueryHandler)
		r.Post("/", view.DnsQueryHandler)
	})

	// Serve the blocked page on the root "/"
	r.Get("/", view.BlockedPageHandler)

	// Read port from .env
	port := os.Getenv("PORT")

	// Start server on given port
	log.Printf("Starting Leo on :%s\n", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal(err)
	}
}
