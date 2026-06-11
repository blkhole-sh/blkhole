package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lemon3studio/blkhole/internal/model"
)

// MockQueryLogRepo is a no-op QueryLogRepo for tests that don't need DB stats.
type MockQueryLogRepo struct{}

func (m *MockQueryLogRepo) Insert(_, _ string, _ bool) error { return nil }
func (m *MockQueryLogRepo) FindByUser(_ int, _ int) ([]*model.QueryLog, error) {
	return nil, nil
}
func (m *MockQueryLogRepo) DeleteOlderThan(_ int) error { return nil }
func (m *MockQueryLogRepo) GetAggregatedStats(_ []string, _, _ time.Time, _ int64) (map[time.Time]int, map[time.Time]int, error) {
	return nil, nil, nil
}

// fullStatsCache extends MockStatsCache with configurable GetUserCounts/GetUserBlockedCounts
type fullStatsCache struct {
	MockStatsCache
	GetUserCountsFn        func(deviceHashes []string, timeRange string) []model.StatCount
	GetUserBlockedCountsFn func(deviceHashes []string, timeRange string) []model.StatCount
}

func (m *fullStatsCache) GetUserCounts(hs []string, r string) []model.StatCount {
	if m.GetUserCountsFn != nil {
		return m.GetUserCountsFn(hs, r)
	}
	return []model.StatCount{}
}
func (m *fullStatsCache) GetUserBlockedCounts(hs []string, r string) []model.StatCount {
	if m.GetUserBlockedCountsFn != nil {
		return m.GetUserBlockedCountsFn(hs, r)
	}
	return []model.StatCount{}
}

func TestStatsController_GetQueryStats_Success(t *testing.T) {
	deviceRepo := &MockDeviceRepo{
		FindByUserFunc: func(userID int) ([]*model.Device, error) {
			return []*model.Device{{ID: 1, Hash: "device-hash"}}, nil
		},
	}
	statsCache := &fullStatsCache{
		GetUserCountsFn: func(hashes []string, _ string) []model.StatCount {
			return []model.StatCount{{Count: 10}}
		},
		GetUserBlockedCountsFn: func(hashes []string, _ string) []model.StatCount {
			return []model.StatCount{{Count: 3}}
		},
	}
	controller := NewStatsController(statsCache, deviceRepo, &MockQueryLogRepo{}, mockAuth(1))

	req := withParam(httptest.NewRequest(http.MethodGet, "/users/1/stats", nil), "userId", "1")
	rr := httptest.NewRecorder()
	controller.GetQueryStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result model.QueryStatsDTO
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result.Total) == 0 || result.Total[0].Count != 10 {
		t.Errorf("expected Total count=10, got %v", result.Total)
	}
}

func TestStatsController_GetQueryStats_BadUserID(t *testing.T) {
	controller := NewStatsController(&fullStatsCache{}, &MockDeviceRepo{}, &MockQueryLogRepo{}, mockAuth(1))

	req := withParam(httptest.NewRequest(http.MethodGet, "/users/abc/stats", nil), "userId", "abc")
	rr := httptest.NewRecorder()
	controller.GetQueryStats(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStatsController_GetQueryStats_InvalidRange(t *testing.T) {
	deviceRepo := &MockDeviceRepo{
		FindByUserFunc: func(userID int) ([]*model.Device, error) {
			return []*model.Device{}, nil
		},
	}
	controller := NewStatsController(&fullStatsCache{}, deviceRepo, &MockQueryLogRepo{}, mockAuth(1))

	req := withParam(httptest.NewRequest(http.MethodGet, "/users/1/stats?range=invalid", nil), "userId", "1")
	rr := httptest.NewRecorder()
	controller.GetQueryStats(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid range, got %d", rr.Code)
	}
}

func TestStatsController_GetQueryStats_DefaultRange(t *testing.T) {
	rangeUsed := ""
	deviceRepo := &MockDeviceRepo{
		FindByUserFunc: func(userID int) ([]*model.Device, error) {
			return []*model.Device{{Hash: "h1"}}, nil
		},
	}
	statsCache := &fullStatsCache{
		GetUserCountsFn: func(_ []string, r string) []model.StatCount {
			rangeUsed = r
			return []model.StatCount{}
		},
		GetUserBlockedCountsFn: func(_ []string, r string) []model.StatCount {
			return []model.StatCount{}
		},
	}
	controller := NewStatsController(statsCache, deviceRepo, &MockQueryLogRepo{}, mockAuth(1))

	req := withParam(httptest.NewRequest(http.MethodGet, "/users/1/stats", nil), "userId", "1")
	rr := httptest.NewRecorder()
	controller.GetQueryStats(rr, req)

	if !strings.Contains(rangeUsed, "24h") {
		t.Errorf("expected default range 24h, got %q", rangeUsed)
	}
}

func TestStatsController_GetQueryStats_DeviceRepoError(t *testing.T) {
	deviceRepo := &MockDeviceRepo{
		FindByUserFunc: func(userID int) ([]*model.Device, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	controller := NewStatsController(&fullStatsCache{}, deviceRepo, &MockQueryLogRepo{}, mockAuth(1))

	req := withParam(httptest.NewRequest(http.MethodGet, "/users/1/stats", nil), "userId", "1")
	rr := httptest.NewRecorder()
	controller.GetQueryStats(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on repo error, got %d", rr.Code)
	}
}
