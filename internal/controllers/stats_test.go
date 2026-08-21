package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/blkhole-sh/blkhole/internal/model"
	"github.com/blkhole-sh/blkhole/internal/repos"
)

// MockQueryLogRepo is a no-op QueryLogRepo for tests that don't need DB stats.
type MockQueryLogRepo struct {
	FindFilteredByUserFn func(int, repos.QueryLogFilter) ([]*model.QueryLogDTO, int, error)
	GetDomainStatsFn     func([]string, time.Time, time.Time, bool, int) ([]model.DomainStat, error)
	GetHourlyActivityFn  func([]string, time.Time, time.Time) (map[string][]int, error)
	GetLastQueriesFn     func([]string) (map[string]time.Time, error)
}

func (m *MockQueryLogRepo) Insert(_, _ string, _ bool) error { return nil }
func (m *MockQueryLogRepo) FindByUser(_ int, _ int) ([]*model.QueryLog, error) {
	return nil, nil
}
func (m *MockQueryLogRepo) FindFilteredByUser(userID int, filter repos.QueryLogFilter) ([]*model.QueryLogDTO, int, error) {
	if m.FindFilteredByUserFn != nil {
		return m.FindFilteredByUserFn(userID, filter)
	}
	return []*model.QueryLogDTO{}, 0, nil
}
func (m *MockQueryLogRepo) DeleteOlderThan(_ int) error { return nil }
func (m *MockQueryLogRepo) GetDomainStats(hashes []string, from, to time.Time, blockedOnly bool, limit int) ([]model.DomainStat, error) {
	if m.GetDomainStatsFn != nil {
		return m.GetDomainStatsFn(hashes, from, to, blockedOnly, limit)
	}
	return []model.DomainStat{}, nil
}
func (m *MockQueryLogRepo) GetHourlyActivity(hashes []string, from, to time.Time) (map[string][]int, error) {
	if m.GetHourlyActivityFn != nil {
		return m.GetHourlyActivityFn(hashes, from, to)
	}
	return map[string][]int{}, nil
}
func (m *MockQueryLogRepo) GetLastQueries(hashes []string) (map[string]time.Time, error) {
	if m.GetLastQueriesFn != nil {
		return m.GetLastQueriesFn(hashes)
	}
	return map[string]time.Time{}, nil
}
func (m *MockQueryLogRepo) GetAggregatedStats(_ []string, _, _ time.Time, _ int64) (map[time.Time]int, map[time.Time]int, error) {
	return nil, nil, nil
}

// fullStatsCache extends MockStatsCache with configurable GetUserCounts/GetUserBlockedCounts
type fullStatsCache struct {
	MockStatsCache
	GetUserCountsFn        func(deviceHashes []string, timeRange string) []model.StatCount
	GetUserBlockedCountsFn func(deviceHashes []string, timeRange string) []model.StatCount
	secondCounts           map[int64]int
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
func (m *fullStatsCache) GetUserSecondCounts(hs []string) map[int64]int {
	if m.secondCounts != nil {
		return m.secondCounts
	}
	return map[int64]int{}
}
func (m *fullStatsCache) GetUserBlockedSecondCounts(hs []string) map[int64]int {
	return map[int64]int{}
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

func TestStatsController_GetQueryStats_DeviceFilter(t *testing.T) {
	var statsHashes []string
	var domainHashes []string
	deviceRepo := &MockDeviceRepo{
		FindByUserFunc: func(userID int) ([]*model.Device, error) {
			return []*model.Device{
				{ID: 1, Hash: "mac", Name: "MacBook", UserID: userID},
				{ID: 2, Hash: "tv", Name: "TV", UserID: userID},
			}, nil
		},
	}
	statsCache := &fullStatsCache{
		GetUserCountsFn: func(hashes []string, _ string) []model.StatCount {
			statsHashes = append([]string{}, hashes...)
			return []model.StatCount{}
		},
	}
	queryLogs := &MockQueryLogRepo{
		GetDomainStatsFn: func(hashes []string, _, _ time.Time, blockedOnly bool, _ int) ([]model.DomainStat, error) {
			domainHashes = append([]string{}, hashes...)
			if blockedOnly {
				t.Fatal("device domain ranking should include all requests")
			}
			return []model.DomainStat{{Domain: "example.com", Count: 4, Blocked: 1}}, nil
		},
		GetHourlyActivityFn: func(_ []string, _, _ time.Time) (map[string][]int, error) {
			return map[string][]int{"mac": make([]int, 24)}, nil
		},
		GetLastQueriesFn: func(_ []string) (map[string]time.Time, error) {
			return map[string]time.Time{}, nil
		},
	}
	controller := NewStatsController(statsCache, deviceRepo, queryLogs, mockAuth(1))
	req := withParam(httptest.NewRequest(http.MethodGet, "/users/1/stats?range=24h&deviceId=1", nil), "userId", "1")
	rr := httptest.NewRecorder()

	controller.GetQueryStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if strings.Join(statsHashes, ",") != "mac" || strings.Join(domainHashes, ",") != "mac" {
		t.Fatalf("expected only selected device, got stats=%v domains=%v", statsHashes, domainHashes)
	}
	var result model.QueryStatsDTO
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result.Domains) != 1 || len(result.Activity) != 1 || result.Activity[0].DeviceName != "MacBook" {
		t.Fatalf("unexpected dashboard stats: %#v", result)
	}
}

func TestStatsController_GetQueryStats_RejectsUnknownDevice(t *testing.T) {
	controller := NewStatsController(&fullStatsCache{}, &MockDeviceRepo{
		FindByUserFunc: func(int) ([]*model.Device, error) {
			return []*model.Device{{ID: 1, Hash: "mac"}}, nil
		},
	}, &MockQueryLogRepo{}, mockAuth(1))
	req := withParam(httptest.NewRequest(http.MethodGet, "/users/1/stats?deviceId=2", nil), "userId", "1")
	rr := httptest.NewRecorder()

	controller.GetQueryStats(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
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

func TestStatsController_GetQueryStats_QPSSeries(t *testing.T) {
	deviceRepo := &MockDeviceRepo{
		FindByUserFunc: func(userID int) ([]*model.Device, error) {
			return []*model.Device{{ID: 1, Hash: "device-hash"}}, nil
		},
	}
	peakSec := time.Now().UTC().Truncate(time.Second).Add(-time.Minute)
	statsCache := &fullStatsCache{
		secondCounts: map[int64]int{peakSec.Unix(): 7, peakSec.Unix() - 1: 2},
	}
	controller := NewStatsController(statsCache, deviceRepo, &MockQueryLogRepo{}, mockAuth(1))

	req := withParam(httptest.NewRequest(http.MethodGet, "/users/1/stats?range=24h", nil), "userId", "1")
	rr := httptest.NewRecorder()
	controller.GetQueryStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result model.QueryStatsDTO
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// 24h of 5-minute windows
	if len(result.QPS) != 288 {
		t.Fatalf("expected 288 qps windows, got %d", len(result.QPS))
	}
	// The window containing the peak must report it at its original second
	found := false
	for _, p := range result.QPS {
		if p.Timestamp.Equal(peakSec) && p.Count == 7 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected peak qps 7 at %v in series", peakSec)
	}
	if len(result.BlockedQPS) != 288 {
		t.Errorf("expected 288 blocked qps windows, got %d", len(result.BlockedQPS))
	}

	// Longer ranges window the same samples with scaled windows (30 min for 7d)
	req = withParam(httptest.NewRequest(http.MethodGet, "/users/1/stats?range=7d", nil), "userId", "1")
	rr = httptest.NewRecorder()
	controller.GetQueryStats(rr, req)
	var weekly model.QueryStatsDTO
	if err := json.NewDecoder(rr.Body).Decode(&weekly); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(weekly.QPS) != 336 {
		t.Errorf("expected 336 qps windows for 7d range, got %d points", len(weekly.QPS))
	}
	found = false
	for _, p := range weekly.QPS {
		if p.Timestamp.Equal(peakSec) && p.Count == 7 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected peak qps 7 at %v in 7d series", peakSec)
	}
}
