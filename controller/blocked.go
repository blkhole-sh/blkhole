package controller

import (
	"net/http"
	"server/model"
	"server/view"

	"github.com/a-h/templ"
)

// Handler that serves the blocked message on root "/"
func BlockedPageController(w http.ResponseWriter, r *http.Request) {
	quote := model.Quote{
		Quote:  "You become what you give your attention to.",
		Author: "Epictetus",
	}

	component := view.Blocked(&quote)

	templ.Handler(component)
}
