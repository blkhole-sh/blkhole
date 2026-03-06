package routes

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/lemon3studio/blkhole/internal/controllers"
	"github.com/lemon3studio/blkhole/internal/middleware"

	"github.com/go-chi/chi/v5"
)

// Init initializes the static file hosting (SolidJS bundle)
func InitWeb(r *chi.Mux, webController controllers.WebController, webFS *embed.FS, devMode string) {
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
}
