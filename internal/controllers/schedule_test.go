package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lemon3studio/blkhole/internal/model"
)

// MockContentBlocker satisfies services.ContentBlocker
type MockContentBlocker struct {
	IsBlockedFunc func(domain, deviceHash string) (bool, error)
}

func (m *MockContentBlocker) Init() error   { return nil }
func (m *MockContentBlocker) Reload() error { return nil }
func (m *MockContentBlocker) IsBlocked(domain, deviceHash string) (bool, error) {
	if m.IsBlockedFunc != nil {
		return m.IsBlockedFunc(domain, deviceHash)
	}
	return false, nil
}

// MockScheduleRepoFull extends MockScheduleRepo with configurable FindByID/FindByUser
type MockScheduleRepoFull struct {
	MockScheduleRepo
	FindByIDFn   func(id int) (*model.Schedule, error)
	FindByUserFn func(userID int) ([]*model.Schedule, error)
}

func (m *MockScheduleRepoFull) FindByID(id int) (*model.Schedule, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(id)
	}
	return nil, fmt.Errorf("not found")
}
func (m *MockScheduleRepoFull) FindByUser(userID int) ([]*model.Schedule, error) {
	if m.FindByUserFn != nil {
		return m.FindByUserFn(userID)
	}
	return []*model.Schedule{}, nil
}

func TestScheduleController_IsBlocked_Blocked(t *testing.T) {
	blocker := &MockContentBlocker{
		IsBlockedFunc: func(domain, _ string) (bool, error) { return true, nil },
	}
	controller := NewScheduleController(&MockScheduleRepo{}, &MockDeviceRepo{}, &MockListRepo{}, blocker)

	req := httptest.NewRequest(http.MethodGet, "/schedule/isblocked?domain=evil.com&deviceHash=abc", nil)
	rr := httptest.NewRecorder()
	controller.IsBlocked(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"blocked":true`) {
		t.Errorf("expected blocked:true in response, got %s", rr.Body.String())
	}
}

func TestScheduleController_IsBlocked_NotBlocked(t *testing.T) {
	blocker := &MockContentBlocker{
		IsBlockedFunc: func(_, _ string) (bool, error) { return false, nil },
	}
	controller := NewScheduleController(&MockScheduleRepo{}, &MockDeviceRepo{}, &MockListRepo{}, blocker)

	req := httptest.NewRequest(http.MethodGet, "/schedule/isblocked?domain=google.com&deviceHash=abc", nil)
	rr := httptest.NewRecorder()
	controller.IsBlocked(rr, req)

	if !strings.Contains(rr.Body.String(), `"blocked":false`) {
		t.Errorf("expected blocked:false in response, got %s", rr.Body.String())
	}
}

func TestScheduleController_IsBlocked_Error(t *testing.T) {
	blocker := &MockContentBlocker{
		IsBlockedFunc: func(_, _ string) (bool, error) { return false, fmt.Errorf("invalid") },
	}
	controller := NewScheduleController(&MockScheduleRepo{}, &MockDeviceRepo{}, &MockListRepo{}, blocker)

	req := httptest.NewRequest(http.MethodGet, "/schedule/isblocked?domain=bad..domain", nil)
	rr := httptest.NewRecorder()
	controller.IsBlocked(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestScheduleController_Create_Success(t *testing.T) {
	schedRepo := &MockScheduleRepoFull{}
	controller := NewScheduleController(schedRepo, &MockDeviceRepo{}, &MockListRepo{}, &MockContentBlocker{})

	body, _ := json.Marshal(map[string]interface{}{
		"name":      "Morning",
		"startTime": "08:00",
		"endTime":   "12:00",
		"userId":    1,
	})
	req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	controller.Create(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestScheduleController_Create_InvalidJSON(t *testing.T) {
	controller := NewScheduleController(&MockScheduleRepo{}, &MockDeviceRepo{}, &MockListRepo{}, &MockContentBlocker{})

	req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewBufferString("bad"))
	rr := httptest.NewRecorder()
	controller.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestScheduleController_FindByID_BadParam(t *testing.T) {
	controller := NewScheduleController(&MockScheduleRepo{}, &MockDeviceRepo{}, &MockListRepo{}, &MockContentBlocker{})

	req := withParam(httptest.NewRequest(http.MethodGet, "/schedules/abc", nil), "id", "abc")
	rr := httptest.NewRecorder()
	controller.FindByID(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestScheduleController_FindByID_NotFound(t *testing.T) {
	schedRepo := &MockScheduleRepoFull{
		FindByIDFn: func(id int) (*model.Schedule, error) { return nil, fmt.Errorf("not found") },
	}
	controller := NewScheduleController(schedRepo, &MockDeviceRepo{}, &MockListRepo{}, &MockContentBlocker{})

	req := withParam(httptest.NewRequest(http.MethodGet, "/schedules/99", nil), "id", "99")
	rr := httptest.NewRecorder()
	controller.FindByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestScheduleController_Delete_Success(t *testing.T) {
	repo := &MockScheduleRepoFull{
		FindByIDFn: func(id int) (*model.Schedule, error) {
			return &model.Schedule{ID: id, IsDefault: false}, nil
		},
	}
	controller := NewScheduleController(repo, &MockDeviceRepo{}, &MockListRepo{}, &MockContentBlocker{})

	req := withParam(httptest.NewRequest(http.MethodDelete, "/schedules/1", nil), "id", "1")
	rr := httptest.NewRecorder()
	controller.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestScheduleController_Delete_BadParam(t *testing.T) {
	controller := NewScheduleController(&MockScheduleRepo{}, &MockDeviceRepo{}, &MockListRepo{}, &MockContentBlocker{})

	req := withParam(httptest.NewRequest(http.MethodDelete, "/schedules/abc", nil), "id", "abc")
	rr := httptest.NewRecorder()
	controller.Delete(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
