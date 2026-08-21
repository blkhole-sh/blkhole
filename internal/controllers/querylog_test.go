package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/blkhole-sh/blkhole/internal/model"
	"github.com/blkhole-sh/blkhole/internal/repos"
)

func TestQueryLogControllerGetLogsFiltersDeviceAndRange(t *testing.T) {
	var gotFilter repos.QueryLogFilter
	repo := &MockQueryLogRepo{
		FindFilteredByUserFn: func(userID int, filter repos.QueryLogFilter) ([]*model.QueryLogDTO, int, error) {
			gotFilter = filter
			return []*model.QueryLogDTO{{ID: 1, DeviceID: 2, DeviceName: "TV", Domain: "example.com"}}, 12, nil
		},
	}
	controller := NewQueryLogController(repo, mockAuth(1))
	req := withParam(httptest.NewRequest(http.MethodGet, "/users/1/logs?deviceId=2&range=7d&limit=8&offset=16", nil), "userId", "1")
	rr := httptest.NewRecorder()

	controller.GetLogs(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if len(gotFilter.DeviceIDs) != 1 || gotFilter.DeviceIDs[0] != 2 || gotFilter.Limit != 8 || gotFilter.Offset != 16 {
		t.Fatalf("unexpected filter: %#v", gotFilter)
	}
	if time.Since(gotFilter.From) < 6*24*time.Hour || time.Since(gotFilter.From) > 8*24*time.Hour {
		t.Fatalf("unexpected range start: %v", gotFilter.From)
	}
	var response struct {
		Items []*model.QueryLogDTO `json:"items"`
		Total int                  `json:"total"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Total != 12 || len(response.Items) != 1 {
		t.Fatalf("unexpected response: %#v", response)
	}
}
