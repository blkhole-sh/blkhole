package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lemon3studio/blkhole/internal/cache"
	"github.com/lemon3studio/blkhole/internal/model"
	"github.com/lemon3studio/blkhole/internal/repos"
	"github.com/lemon3studio/blkhole/internal/services"
)

// StatsController defines the interface for stats operations
type StatsController interface {
	GetQueryStats(http.ResponseWriter, *http.Request)
}

// statsController implements the StatsController interface
type statsController struct {
	statsCache  cache.StatsCache
	devices     repos.DeviceRepo
	queryLogs   repos.QueryLogRepo
	authService services.AuthService
}

// NewStatsController creates a new StatsController instance
func NewStatsController(statsCache cache.StatsCache, devices repos.DeviceRepo, queryLogs repos.QueryLogRepo, authService services.AuthService) StatsController {
	return &statsController{
		statsCache:  statsCache,
		devices:     devices,
		queryLogs:   queryLogs,
		authService: authService,
	}
}

// GetQueryStats returns aggregated query statistics (total and blocked) for all user devices
func (sc *statsController) GetQueryStats(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL params
	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("failed to parse userId: %v", err)
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	// Users may only read their own stats
	user, ok := currentUser(w, r, sc.authService)
	if !ok {
		return
	}
	if userID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get time range from query params (default to 24h)
	timeRange := r.URL.Query().Get("range")
	if timeRange == "" {
		timeRange = cache.Range24h
	}

	// Validate time range
	if timeRange != cache.Range24h && timeRange != cache.Range7d && timeRange != cache.Range30d {
		http.Error(w, "Invalid time range. Must be 24h, 7d, or 30d", http.StatusBadRequest)
		return
	}

	// Get all devices for the user
	devices, err := sc.devices.FindByUser(userID)
	if err != nil {
		log.Printf("failed to find devices for user %d: %v", userID, err)
		http.Error(w, "Failed to find devices", http.StatusInternalServerError)
		return
	}

	// Extract device hashes
	deviceHashes := make([]string, 0, len(devices))
	for _, device := range devices {
		deviceHashes = append(deviceHashes, device.Hash)
	}

	// Get cache stats (full window, real-time)
	total := sc.statsCache.GetUserCounts(deviceHashes, timeRange)
	blocked := sc.statsCache.GetUserBlockedCounts(deviceHashes, timeRange)

	// Merge with DB stats to recover history lost across restarts.
	// The cache is authoritative for buckets it has (the DB may lag behind
	// unflushed query logs); DB-only buckets fill in pre-restart history.
	var stepSec int64
	var span time.Duration
	switch timeRange {
	case cache.Range24h:
		stepSec, span = 60, 24*time.Hour
	case cache.Range7d:
		stepSec, span = 300, 7*24*time.Hour
	case cache.Range30d:
		stepSec, span = 3600, 30*24*time.Hour
	}
	end := time.Now().UTC().Truncate(time.Duration(stepSec) * time.Second).Add(time.Duration(stepSec) * time.Second)
	start := end.Add(-span)

	dbTotal, dbBlocked, err := sc.queryLogs.GetAggregatedStats(deviceHashes, start, end, stepSec)
	if err != nil {
		log.Printf("failed to get db stats: %v", err)
	} else {
		total = mergeCounts(total, dbTotal)
		blocked = mergeCounts(blocked, dbBlocked)
	}

	stats := model.QueryStatsDTO{Total: total, Blocked: blocked}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

// mergeCounts merges cache and DB buckets, taking the larger count per bucket.
// Timestamps are normalized to UTC so cache (local) and DB (UTC) buckets for
// the same instant collapse into one.
func mergeCounts(cacheCounts []model.StatCount, dbCounts map[time.Time]int) []model.StatCount {
	merged := make(map[time.Time]int, len(cacheCounts)+len(dbCounts))
	for ts, count := range dbCounts {
		merged[ts.UTC()] = count
	}
	for _, s := range cacheCounts {
		ts := s.Timestamp.UTC()
		if s.Count > merged[ts] {
			merged[ts] = s.Count
		}
	}

	result := make([]model.StatCount, 0, len(merged))
	for ts, count := range merged {
		result = append(result, model.StatCount{Timestamp: ts, Count: count})
	}

	// Sort by timestamp for proper chart rendering
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	return result
}
