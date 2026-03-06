package routes

import (
	"github.com/lemon3studio/blkhole/internal/controllers"
	"github.com/lemon3studio/blkhole/internal/middleware"

	"github.com/go-chi/chi/v5"
)

// Init initializes the DoH routes
func InitDoH(
	r *chi.Mux,
	dohController controllers.DoHController,
	domain string,
	devMode string,
) {
	// Route specifically for /dns-query, handling both GET and POST requests
	r.Route("/dns-query", func(r chi.Router) {
		r.Use(middleware.DeviceHashFromHost(domain))
		r.Get("/", dohController.DNSQuery)
		r.Post("/", dohController.DNSQuery)
	})
}
