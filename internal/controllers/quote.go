package controllers

import (
	"encoding/json"
	"github.com/lemon3studio/blkhole/internal/model"
	"math/rand"
	"net/http"
)

// QuoteController defines the interface for quote operations
type QuoteController interface {
	Random(w http.ResponseWriter, r *http.Request)
}

// quoteController implements the QuoteController interface
type quoteController struct{}

// NewQuoteController creates a new QuoteController instance
func NewQuoteController() QuoteController {
	return &quoteController{}
}

// quotes contains a collection of stoic quotes
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
		Quote:  "He who conquers himself conquers all.",
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

func (qc *quoteController) Random(w http.ResponseWriter, r *http.Request) {
	// Pick a random stoic quote
	quote := quotes[rand.Intn(len(quotes))]

	json.NewEncoder(w).Encode(quote)
}
