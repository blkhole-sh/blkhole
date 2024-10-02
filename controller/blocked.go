package controller

import (
	"math/rand"
	"net/http"
	"server/model"
	"server/view"

	"github.com/a-h/templ"
)

var quotes = []model.Quote{
	{
		Quote:  "You become what you give your attention to.",
		Author: "Epictetus",
	},
	{
		Quote:  "Progress is not achieved by luck or accident, but by working on yourself daily.",
		Author: "Epictetus",
	},
	{
		Quote:  "No man is free who is not master of himself.",
		Author: "Epictetus",
	},
	{
		Quote:  "Stop allowing your mind to be a slave, to be jerked about.",
		Author: "Marcus Aurelius",
	},
	{
		Quote:  "Most powerful is he who has himself in his own power.",
		Author: "Seneca",
	},
	{
		Quote:  "He ho conquers himself conquers all.",
		Author: "Epictetus",
	},
	{
		Quote:  "Life is long if you know how to use it.",
		Author: "Seneca",
	},
	{
		Quote:  "I do what is mine to do, the rest does not disturb me.",
		Author: "Marcus Aurelius",
	},
	{
		Quote:  "The present is all we have to live in. Or to lose.",
		Author: "Marcus Aurelius",
	},
}

// Handler that serves the blocked message on root "/"
func BlockedPageController(w http.ResponseWriter, r *http.Request) {
	// Create blocked page
	page := view.Blocked(&quotes[rand.Intn(len(quotes))])

	// Serve the template component using templ.Handler
	templ.Handler(page).ServeHTTP(w, r)
}
