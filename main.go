package main

import (
	"log"
	"net/http"
	"os"
	"server/views"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Create a new router using chi
	r := chi.NewRouter()

	// Add some middleware for better logging and recovery
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Route specifically for /dns-query, handling both GET and POST requests
	r.Route("/dns-query", func(r chi.Router) {
		r.Get("/", views.DnsQueryHandler)
		r.Post("/", views.DnsQueryHandler)
	})

	// Serve the blocked page on the root "/"
	r.Get("/", views.BlockedPageHandler)

	// Read port from .env
	port := os.Getenv("PORT")

	// Start server on given port
	log.Printf("Starting Leo on :%s\n", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal(err)
	}
}
