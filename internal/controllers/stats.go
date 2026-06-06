package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lemon3studio/blkhole/internal/cache"
	"github.com/lemon3studio/blkhole/internal/model"
	"github.com/lemon3studio/blkhole/internal/repos"
)

// StatsController defines the interface for stats operations
type StatsController interface {
	GetQueryStats(http.ResponseWriter, *http.Request)
}

// statsController implements the StatsController interface
type statsController struct {
	statsCache cache.StatsCache
	devices    repos.DeviceRepo
	queryLogs  repos.QueryLogRepo
}

// NewStatsController creates a new StatsController instance
func NewStatsController(statsCache cache.StatsCache, devices repos.DeviceRepo, queryLogs repos.QueryLogRepo) StatsController {
	return &statsController{
		statsCache: statsCache,
		devices:    devices,
		queryLogs:  queryLogs,
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
	// For each bucket: use max(cache, db) — cache wins for live data,
	// DB fills in buckets the cache doesn't have yet.
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
		for i, s := range total {
			if dbCount := dbTotal[s.Timestamp]; dbCount > s.Count {
				total[i].Count = dbCount
			}
		}
		for i, s := range blocked {
			if dbCount := dbBlocked[s.Timestamp]; dbCount > s.Count {
				blocked[i].Count = dbCount
			}
		}
	}

	stats := model.QueryStatsDTO{Total: total, Blocked: blocked}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
