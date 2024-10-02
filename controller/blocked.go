package controller

import (
	"net/http"
	"server/model"
	"server/view"

	"github.com/a-h/templ"
)

// Handler that serves the blocked message on root "/"
func BlockedPageController(w http.ResponseWriter, r *http.Request) {
	// Initialize a quote
	quote := model.Quote{
		Quote:  "You become what you give your attention to.",
		Author: "Epictetus",
	}

	// Create blocked page
	page := view.Blocked(&quote)

	// Serve the template component using templ.Handler
	templ.Handler(page).ServeHTTP(w, r)
}
