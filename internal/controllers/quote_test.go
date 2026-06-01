package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lemon3studio/blkhole/internal/model"
)

func TestQuoteController_Random_ReturnsValidQuote(t *testing.T) {
	controller := NewQuoteController()

	req := httptest.NewRequest(http.MethodGet, "/quote", nil)
	rr := httptest.NewRecorder()

	controller.Random(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var q model.Quote
	if err := json.NewDecoder(rr.Body).Decode(&q); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if q.Quote == "" {
		t.Error("expected non-empty quote text")
	}
	if q.Author == "" {
		t.Error("expected non-empty author")
	}
}

func TestQuoteController_Random_ReturnsKnownQuote(t *testing.T) {
	controller := NewQuoteController()
	seen := make(map[string]bool)

	for i := 0; i < 50; i++ {
		req := httptest.NewRequest(http.MethodGet, "/quote", nil)
		rr := httptest.NewRecorder()
		controller.Random(rr, req)

		var q model.Quote
		json.NewDecoder(rr.Body).Decode(&q)
		seen[q.Author] = true
	}

	if len(seen) == 0 {
		t.Error("no authors returned in 50 calls")
	}
	for author := range seen {
		validAuthors := map[string]bool{"Epictetus": true, "Marcus Aurelius": true, "Seneca": true}
		if !validAuthors[author] {
			t.Errorf("unexpected author: %s", author)
		}
	}
}
