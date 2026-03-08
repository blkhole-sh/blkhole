package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lemon3studio/blkhole/internal/model"
	"github.com/lemon3studio/blkhole/internal/services"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	LoginFunc           func(email, password string) (*services.LoginResult, error)
	RefreshTokenFunc    func(refreshToken string) (*services.LoginResult, error)
	UserFromContextFunc func(ctx context.Context) (*model.User, error)
}

func (m *MockAuthService) Login(email, password string) (*services.LoginResult, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(email, password)
	}
	return nil, fmt.Errorf("LoginFunc not implemented")
}

func (m *MockAuthService) RefreshToken(refreshToken string) (*services.LoginResult, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(refreshToken)
	}
	return nil, fmt.Errorf("RefreshTokenFunc not implemented")
}

func (m *MockAuthService) UserFromContext(ctx context.Context) (*model.User, error) {
	if m.UserFromContextFunc != nil {
		return m.UserFromContextFunc(ctx)
	}
	return nil, fmt.Errorf("UserFromContextFunc not implemented")
}

func TestAuthController_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		mockLogin      func(email, password string) (*services.LoginResult, error)
		expectedStatus int
		expectedBody   string
		checkCookies   bool
	}{
		{
			name: "Success",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password",
			},
			mockLogin: func(email, password string) (*services.LoginResult, error) {
				return &services.LoginResult{
					User: &model.User{
						ID:    1,
						Name:  "Test User",
						Email: "test@example.com",
					},
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"email":"test@example.com"`,
			checkCookies:   true,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid-json",
			mockLogin:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
			checkCookies:   false,
		},
		{
			name: "Missing Credentials",
			requestBody: map[string]string{
				"email":    "",
				"password": "",
			},
			mockLogin:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Email and password required",
			checkCookies:   false,
		},
		{
			name: "Invalid Credentials",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "wrong",
			},
			mockLogin: func(email, password string) (*services.LoginResult, error) {
				return nil, fmt.Errorf("invalid credentials")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid credentials",
			checkCookies:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{
				LoginFunc: tt.mockLogin,
			}
			controller := NewAuthController(mockService)

			var body []byte
			if s, ok := tt.requestBody.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			controller.Login(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}

			if tt.checkCookies {
				cookies := w.Result().Cookies()
				foundAccess := false
				foundRefresh := false
				for _, c := range cookies {
					if c.Name == "access_token" && c.Value == "access_token" && c.HttpOnly {
						foundAccess = true
					}
					if c.Name == "refresh_token" && c.Value == "refresh_token" && c.HttpOnly {
						foundRefresh = true
					}
				}
				if !foundAccess || !foundRefresh {
					t.Errorf("expected secure cookies to be set")
				}
			}
		})
	}
}

func TestAuthController_RefreshToken(t *testing.T) {
	tests := []struct {
		name             string
		cookieValue      string
		mockRefresh      func(refreshToken string) (*services.LoginResult, error)
		expectedStatus   int
		expectedBody     string
		checkCookies     bool
		checkClearCookie bool
	}{
		{
			name:        "Success",
			cookieValue: "valid_refresh_token",
			mockRefresh: func(refreshToken string) (*services.LoginResult, error) {
				return &services.LoginResult{
					User: &model.User{
						ID:    1,
						Name:  "Test User",
						Email: "test@example.com",
					},
					AccessToken:  "new_access_token",
					RefreshToken: "new_refresh_token",
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"email":"test@example.com"`,
			checkCookies:   true,
		},
		{
			name:           "Missing Cookie",
			cookieValue:    "",
			mockRefresh:    nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Refresh token not found",
			checkCookies:   false,
		},
		{
			name:        "Invalid Token",
			cookieValue: "invalid_refresh_token",
			mockRefresh: func(refreshToken string) (*services.LoginResult, error) {
				return nil, fmt.Errorf("invalid refresh token")
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedBody:     "invalid refresh token",
			checkCookies:     false,
			checkClearCookie: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{
				RefreshTokenFunc: tt.mockRefresh,
			}
			controller := NewAuthController(mockService)

			req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: tt.cookieValue})
			}
			w := httptest.NewRecorder()

			controller.RefreshToken(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}

			if tt.checkCookies {
				cookies := w.Result().Cookies()
				foundAccess := false
				foundRefresh := false
				for _, c := range cookies {
					if c.Name == "access_token" && c.Value == "new_access_token" {
						foundAccess = true
					}
					if c.Name == "refresh_token" && c.Value == "new_refresh_token" {
						foundRefresh = true
					}
				}
				if !foundAccess || !foundRefresh {
					t.Errorf("expected new cookies to be set")
				}
			}

			if tt.checkClearCookie {
				cookies := w.Result().Cookies()
				cleared := false
				for _, c := range cookies {
					if c.Name == "access_token" && c.Value == "" && c.Expires.Before(time.Now()) {
						cleared = true
					}
				}
				if !cleared {
					t.Errorf("expected cookies to be cleared")
				}
			}
		})
	}
}

func TestAuthController_Logout(t *testing.T) {
	controller := NewAuthController(&MockAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	controller.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	cookies := w.Result().Cookies()
	clearedAccess := false
	clearedRefresh := false

	for _, c := range cookies {
		if c.Name == "access_token" && c.Value == "" && c.Expires.Unix() == 0 {
			clearedAccess = true
		}
		if c.Name == "refresh_token" && c.Value == "" && c.Expires.Unix() == 0 {
			clearedRefresh = true
		}
	}

	if !clearedAccess || !clearedRefresh {
		t.Errorf("expected cookies to be cleared (expired)")
	}
}

func TestAuthController_GetCurrentUser(t *testing.T) {
	tests := []struct {
		name           string
		mockUser       func(ctx context.Context) (*model.User, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			mockUser: func(ctx context.Context) (*model.User, error) {
				return &model.User{
					ID:    1,
					Name:  "Test User",
					Email: "test@example.com",
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"email":"test@example.com"`,
		},
		{
			name: "Unauthorized",
			mockUser: func(ctx context.Context) (*model.User, error) {
				return nil, fmt.Errorf("unauthorized")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockAuthService{
				UserFromContextFunc: tt.mockUser,
			}
			controller := NewAuthController(mockService)

			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			w := httptest.NewRecorder()

			controller.GetCurrentUser(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}
