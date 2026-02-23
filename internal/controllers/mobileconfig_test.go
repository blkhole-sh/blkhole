package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lemon3studio/blkhole/internal/model"
	"github.com/lemon3studio/blkhole/internal/services"
)

// Mock DeviceRepo
type mockDeviceRepo struct {
	devices map[int]*model.Device
}

func (m *mockDeviceRepo) FindByID(id int) (*model.Device, error) {
	if d, ok := m.devices[id]; ok {
		return d, nil
	}
	return nil, errors.New("device not found")
}

func (m *mockDeviceRepo) Create(d *model.Device) error                           { return nil }
func (m *mockDeviceRepo) Update(id int, d *model.Device) error                   { return nil }
func (m *mockDeviceRepo) Delete(id int) error                                    { return nil }
func (m *mockDeviceRepo) LinkSchedule(id int, scheduleID int) error              { return nil }
func (m *mockDeviceRepo) LoadScheduleIDs(id int) ([]int, error)                  { return nil, nil }
func (m *mockDeviceRepo) FindByHash(hash string) (*model.Device, error)          { return nil, nil }
func (m *mockDeviceRepo) FindByUser(userID int) ([]*model.Device, error)         { return nil, nil }
func (m *mockDeviceRepo) FindNamesByScheduleID(scheduleID int) ([]string, error) { return nil, nil }
func (m *mockDeviceRepo) FindBySchedule(scheduleID int) ([]*model.Device, error) { return nil, nil }
func (m *mockDeviceRepo) FindAll() ([]*model.Device, error)                      { return nil, nil }
func (m *mockDeviceRepo) FindDeviceSchedule() ([]*model.DeviceSchedule, error)   { return nil, nil }

// Mock AuthService
type mockAuthService struct {
	currentUser *model.User
	err         error
}

func (m *mockAuthService) UserFromContext(ctx context.Context) (*model.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.currentUser, nil
}

func (m *mockAuthService) Login(email, password string) (*services.LoginResult, error) {
	return nil, nil
}
func (m *mockAuthService) RefreshToken(refreshToken string) (*services.LoginResult, error) {
	return nil, nil
}

func TestGenerateConfig(t *testing.T) {
	// Setup
	deviceRepo := &mockDeviceRepo{
		devices: map[int]*model.Device{
			1: {ID: 1, UserID: 100, Name: "User's Device", Hash: "hash1"},
			2: {ID: 2, UserID: 200, Name: "Other's Device", Hash: "hash2"},
		},
	}

	tests := []struct {
		name           string
		deviceID       string
		currentUser    *model.User
		authErr        error
		expectedStatus int
	}{
		{
			name:           "Authorized Access",
			deviceID:       "1",
			currentUser:    &model.User{ID: 100},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized Access (Wrong User)",
			deviceID:       "1",
			currentUser:    &model.User{ID: 200},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Unauthorized Access (Device belongs to another)",
			deviceID:       "2",
			currentUser:    &model.User{ID: 100},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Device Not Found",
			deviceID:       "999",
			currentUser:    &model.User{ID: 100},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "User Not Authenticated",
			deviceID:       "1",
			currentUser:    nil,
			authErr:        errors.New("no token"),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Device ID",
			deviceID:       "abc",
			currentUser:    &model.User{ID: 100},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := &mockAuthService{
				currentUser: tt.currentUser,
				err:         tt.authErr,
			}
			controller := NewMobileConfigController("example.com", deviceRepo, authService)

			req := httptest.NewRequest("GET", "/devices/"+tt.deviceID+"/config", nil)

			// Setup chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.deviceID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			controller.GenerateConfig(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
