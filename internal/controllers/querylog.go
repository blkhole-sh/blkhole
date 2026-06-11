package controllers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/blkhole-sh/blkhole/internal/repos"
	"github.com/blkhole-sh/blkhole/internal/services"
)

// QueryLogController defines the interface for query log operations
type QueryLogController interface {
	GetLogs(http.ResponseWriter, *http.Request)
	ExportLogs(http.ResponseWriter, *http.Request)
}

type queryLogController struct {
	queryLogs   repos.QueryLogRepo
	authService services.AuthService
}

// NewQueryLogController creates a new QueryLogController instance
func NewQueryLogController(queryLogs repos.QueryLogRepo, authService services.AuthService) QueryLogController {
	return &queryLogController{queryLogs: queryLogs, authService: authService}
}

// requireOwnUserID parses the userId path parameter and verifies it matches
// the authenticated user. On failure it writes an error response and returns false.
func (c *queryLogController) requireOwnUserID(w http.ResponseWriter, r *http.Request) (int, bool) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil {
		log.Printf("unable to parse userId from path parameter: %v", err)
		http.Error(w, "Unable to parse userId from path parameter", http.StatusBadRequest)
		return 0, false
	}

	user, ok := currentUser(w, r, c.authService)
	if !ok {
		return 0, false
	}
	if userID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return 0, false
	}

	return userID, true
}

func (c *queryLogController) GetLogs(w http.ResponseWriter, r *http.Request) {
	userID, ok := c.requireOwnUserID(w, r)
	if !ok {
		return
	}

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	logs, err := c.queryLogs.FindByUser(userID, limit)
	if err != nil {
		log.Printf("failed to fetch query logs for user %d: %v", userID, err)
		http.Error(w, "Unable to fetch query logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		log.Printf("failed to encode query logs: %v", err)
	}
}

func (c *queryLogController) ExportLogs(w http.ResponseWriter, r *http.Request) {
	userID, ok := c.requireOwnUserID(w, r)
	if !ok {
		return
	}

	logs, err := c.queryLogs.FindByUser(userID, 100000)
	if err != nil {
		log.Printf("failed to fetch query logs for user %d: %v", userID, err)
		http.Error(w, "Unable to fetch query logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="query_log_%d.csv"`, userID))

	cw := csv.NewWriter(w)
	cw.Write([]string{"timestamp", "domain", "device_hash", "blocked"})
	for _, entry := range logs {
		blockedStr := "false"
		if entry.Blocked {
			blockedStr = "true"
		}
		cw.Write([]string{
			time.Unix(entry.Timestamp, 0).UTC().Format(time.RFC3339),
			entry.Domain,
			entry.DeviceHash,
			blockedStr,
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("failed to write query log export: %v", err)
	}
}
