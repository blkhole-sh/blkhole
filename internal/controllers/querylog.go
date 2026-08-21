package controllers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/blkhole-sh/blkhole/internal/model"
	"github.com/blkhole-sh/blkhole/internal/repos"
	"github.com/blkhole-sh/blkhole/internal/services"
	"github.com/go-chi/chi/v5"
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
	if limit > 100 {
		limit = 100
	}
	offset := 0
	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		parsed, err := strconv.Atoi(rawOffset)
		if err != nil || parsed < 0 {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		}
		offset = parsed
	}
	filter, ok := queryLogFilter(r, limit, offset)
	if !ok {
		http.Error(w, "Invalid query log filter", http.StatusBadRequest)
		return
	}

	logs, total, err := c.queryLogs.FindFilteredByUser(userID, filter)
	if err != nil {
		log.Printf("failed to fetch query logs for user %d: %v", userID, err)
		http.Error(w, "Unable to fetch query logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(struct {
		Items []*model.QueryLogDTO `json:"items"`
		Total int                  `json:"total"`
	}{Items: logs, Total: total}); err != nil {
		log.Printf("failed to encode query logs: %v", err)
	}
}

func (c *queryLogController) ExportLogs(w http.ResponseWriter, r *http.Request) {
	userID, ok := c.requireOwnUserID(w, r)
	if !ok {
		return
	}

	filter, ok := queryLogFilter(r, 100000, 0)
	if !ok {
		http.Error(w, "Invalid query log filter", http.StatusBadRequest)
		return
	}
	logs, _, err := c.queryLogs.FindFilteredByUser(userID, filter)
	if err != nil {
		log.Printf("failed to fetch query logs for user %d: %v", userID, err)
		http.Error(w, "Unable to fetch query logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="query_log_%d.csv"`, userID))

	cw := csv.NewWriter(w)
	cw.Write([]string{"timestamp", "device", "domain", "verdict"})
	for _, entry := range logs {
		verdict := "Allowed"
		if entry.Blocked {
			verdict = "Blocked"
		}
		cw.Write([]string{
			time.Unix(entry.Timestamp, 0).UTC().Format(time.RFC3339),
			entry.DeviceName,
			entry.Domain,
			verdict,
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("failed to write query log export: %v", err)
	}
}

func queryLogFilter(r *http.Request, limit, offset int) (repos.QueryLogFilter, bool) {
	deviceIDs := []int{}
	rawDeviceIDs := r.URL.Query().Get("deviceIds")
	if rawDeviceIDs == "" {
		rawDeviceIDs = r.URL.Query().Get("deviceId")
	}
	if rawDeviceIDs != "" {
		for _, rawID := range strings.Split(rawDeviceIDs, ",") {
			id, err := strconv.Atoi(strings.TrimSpace(rawID))
			if err != nil || id <= 0 {
				return repos.QueryLogFilter{}, false
			}
			deviceIDs = append(deviceIDs, id)
		}
	}

	rangeValue := r.URL.Query().Get("range")
	if rangeValue == "" {
		rangeValue = "24h"
	}
	var span time.Duration
	switch rangeValue {
	case "1h":
		span = time.Hour
	case "24h":
		span = 24 * time.Hour
	case "7d":
		span = 7 * 24 * time.Hour
	case "30d":
		span = 30 * 24 * time.Hour
	default:
		return repos.QueryLogFilter{}, false
	}
	to := time.Now()
	return repos.QueryLogFilter{
		DeviceIDs: deviceIDs,
		From:      to.Add(-span),
		To:        to,
		Limit:     limit,
		Offset:    offset,
	}, true
}
