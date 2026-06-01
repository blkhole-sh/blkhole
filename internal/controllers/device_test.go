package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lemon3studio/blkhole/internal/model"
)

// MockDeviceRepo satisfies repos.DeviceRepo
type MockDeviceRepo struct {
	CreateFunc                func(d *model.Device) error
	UpdateFunc                func(id int, d *model.Device) error
	DeleteFunc                func(id int) error
	LoadScheduleIDsFunc       func(id int) ([]int, error)
	FindByIDFunc              func(id int) (*model.Device, error)
	FindByHashFunc            func(hash string) (*model.Device, error)
	FindByUserFunc            func(userID int) ([]*model.Device, error)
	FindNamesByScheduleIDFunc func(scheduleID int) ([]string, error)
}

func (m *MockDeviceRepo) Create(d *model.Device) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(d)
	}
	return nil
}
func (m *MockDeviceRepo) Update(id int, d *model.Device) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(id, d)
	}
	return nil
}
func (m *MockDeviceRepo) Delete(id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}
func (m *MockDeviceRepo) LinkSchedule(id int, scheduleID int) error { return nil }
func (m *MockDeviceRepo) LoadScheduleIDs(id int) ([]int, error) {
	if m.LoadScheduleIDsFunc != nil {
		return m.LoadScheduleIDsFunc(id)
	}
	return []int{}, nil
}
func (m *MockDeviceRepo) FindByID(id int) (*model.Device, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(id)
	}
	return nil, fmt.Errorf("not found")
}
func (m *MockDeviceRepo) FindByHash(hash string) (*model.Device, error) {
	if m.FindByHashFunc != nil {
		return m.FindByHashFunc(hash)
	}
	return nil, fmt.Errorf("not found")
}
func (m *MockDeviceRepo) FindByUser(userID int) ([]*model.Device, error) {
	if m.FindByUserFunc != nil {
		return m.FindByUserFunc(userID)
	}
	return []*model.Device{}, nil
}
func (m *MockDeviceRepo) FindNamesByScheduleID(scheduleID int) ([]string, error) {
	if m.FindNamesByScheduleIDFunc != nil {
		return m.FindNamesByScheduleIDFunc(scheduleID)
	}
	return []string{}, nil
}
func (m *MockDeviceRepo) FindBySchedule(scheduleID int) ([]*model.Device, error) {
	return []*model.Device{}, nil
}
func (m *MockDeviceRepo) FindAll() ([]*model.Device, error) { return []*model.Device{}, nil }
func (m *MockDeviceRepo) FindDeviceSchedule() ([]*model.DeviceSchedule, error) {
	return []*model.DeviceSchedule{}, nil
}

// MockScheduleRepo satisfies repos.ScheduleRepo
type MockScheduleRepo struct {
	FindNamesByDeviceIDFunc func(deviceID int) ([]string, error)
}

func (m *MockScheduleRepo) Create(s *model.Schedule) (int, error)      { return 0, nil }
func (m *MockScheduleRepo) Update(id int, s *model.Schedule) error      { return nil }
func (m *MockScheduleRepo) Delete(id int) error                          { return nil }
func (m *MockScheduleRepo) LinkDevice(scheduleID int, deviceID int) error { return nil }
func (m *MockScheduleRepo) LinkRule(id int, ruleID int) error            { return nil }
func (m *MockScheduleRepo) LinkList(id int, listID int) error            { return nil }
func (m *MockScheduleRepo) LoadDeviceIDs(id int) ([]int, error)          { return []int{}, nil }
func (m *MockScheduleRepo) LoadRuleIDs(id int) ([]int, error)            { return []int{}, nil }
func (m *MockScheduleRepo) LoadListIDs(id int) ([]int, error)            { return []int{}, nil }
func (m *MockScheduleRepo) LoadRelations(s *model.Schedule) error        { return nil }
func (m *MockScheduleRepo) FindByID(id int) (*model.Schedule, error)     { return nil, fmt.Errorf("not found") }
func (m *MockScheduleRepo) FindByUser(userID int) ([]*model.Schedule, error) {
	return []*model.Schedule{}, nil
}
func (m *MockScheduleRepo) FindNamesByDeviceID(deviceID int) ([]string, error) {
	if m.FindNamesByDeviceIDFunc != nil {
		return m.FindNamesByDeviceIDFunc(deviceID)
	}
	return []string{}, nil
}
func (m *MockScheduleRepo) FindByDevice(deviceHash string) ([]*model.Schedule, error) {
	return []*model.Schedule{}, nil
}
func (m *MockScheduleRepo) FindByRule(ruleID int) ([]*model.Schedule, error) {
	return []*model.Schedule{}, nil
}
func (m *MockScheduleRepo) FindByList(listID int) ([]*model.Schedule, error) {
	return []*model.Schedule{}, nil
}
func (m *MockScheduleRepo) FindAll() ([]*model.Schedule, error) { return []*model.Schedule{}, nil }
func (m *MockScheduleRepo) FindScheduleRule() ([]*model.ScheduleRule, error) {
	return []*model.ScheduleRule{}, nil
}

// MockCryptoService satisfies services.CryptoService
type MockCryptoService struct {
	RandomHashFunc func() (string, error)
}

func (m *MockCryptoService) RandomHash() (string, error) {
	if m.RandomHashFunc != nil {
		return m.RandomHashFunc()
	}
	return "mockhash123", nil
}
func (m *MockCryptoService) HashPassword(password string) (string, error) { return "hashed", nil }
func (m *MockCryptoService) VerifyPassword(password, hashed string) (bool, error) {
	return true, nil
}

// withParam attaches chi URL params to a request context.
func withParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestDeviceController_Create_Success(t *testing.T) {
	deviceRepo := &MockDeviceRepo{
		CreateFunc: func(d *model.Device) error {
			d.ID = 42
			return nil
		},
	}
	controller := NewDeviceController(deviceRepo, &MockScheduleRepo{}, &MockCryptoService{})

	body, _ := json.Marshal(map[string]interface{}{"name": "Test Device", "os": "iOS", "userId": 1})
	req := httptest.NewRequest(http.MethodPost, "/devices", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	controller.Create(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeviceController_Create_InvalidJSON(t *testing.T) {
	controller := NewDeviceController(&MockDeviceRepo{}, &MockScheduleRepo{}, &MockCryptoService{})

	req := httptest.NewRequest(http.MethodPost, "/devices", bytes.NewBufferString("not-json"))
	rr := httptest.NewRecorder()
	controller.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestDeviceController_Create_HashError(t *testing.T) {
	crypto := &MockCryptoService{
		RandomHashFunc: func() (string, error) { return "", fmt.Errorf("crypto fail") },
	}
	controller := NewDeviceController(&MockDeviceRepo{}, &MockScheduleRepo{}, crypto)

	body, _ := json.Marshal(map[string]interface{}{"name": "D", "os": "iOS", "userId": 1})
	req := httptest.NewRequest(http.MethodPost, "/devices", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	controller.Create(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestDeviceController_FindByID_Success(t *testing.T) {
	deviceRepo := &MockDeviceRepo{
		FindByIDFunc: func(id int) (*model.Device, error) {
			return &model.Device{ID: id, Name: "Found", OS: model.MacOS}, nil
		},
	}
	controller := NewDeviceController(deviceRepo, &MockScheduleRepo{}, &MockCryptoService{})

	req := withParam(httptest.NewRequest(http.MethodGet, "/devices/1", nil), "id", "1")
	rr := httptest.NewRecorder()
	controller.FindByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestDeviceController_FindByID_BadParam(t *testing.T) {
	controller := NewDeviceController(&MockDeviceRepo{}, &MockScheduleRepo{}, &MockCryptoService{})

	req := withParam(httptest.NewRequest(http.MethodGet, "/devices/abc", nil), "id", "abc")
	rr := httptest.NewRecorder()
	controller.FindByID(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestDeviceController_FindByID_NotFound(t *testing.T) {
	deviceRepo := &MockDeviceRepo{
		FindByIDFunc: func(id int) (*model.Device, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	controller := NewDeviceController(deviceRepo, &MockScheduleRepo{}, &MockCryptoService{})

	req := withParam(httptest.NewRequest(http.MethodGet, "/devices/99", nil), "id", "99")
	rr := httptest.NewRecorder()
	controller.FindByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestDeviceController_Delete_Success(t *testing.T) {
	deviceRepo := &MockDeviceRepo{
		DeleteFunc: func(id int) error { return nil },
	}
	controller := NewDeviceController(deviceRepo, &MockScheduleRepo{}, &MockCryptoService{})

	req := withParam(httptest.NewRequest(http.MethodDelete, "/devices/1", nil), "id", "1")
	rr := httptest.NewRecorder()
	controller.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestDeviceController_Delete_BadParam(t *testing.T) {
	controller := NewDeviceController(&MockDeviceRepo{}, &MockScheduleRepo{}, &MockCryptoService{})

	req := withParam(httptest.NewRequest(http.MethodDelete, "/devices/abc", nil), "id", "abc")
	rr := httptest.NewRecorder()
	controller.Delete(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
