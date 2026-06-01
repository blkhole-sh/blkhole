package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/jwtauth/v5"
)

func TestCookieAuthenticator_ValidToken(t *testing.T) {
	ja := jwtauth.New("HS256", []byte("test-secret"), nil)
	_, tokenStr, _ := ja.Encode(map[string]interface{}{"user_id": 1})

	reached := false
	handler := CookieAuthenticator(ja)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenStr})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 with valid token, got %d", rr.Code)
	}
	if !reached {
		t.Error("expected next handler to be called")
	}
}

func TestCookieAuthenticator_NoCookie(t *testing.T) {
	ja := jwtauth.New("HS256", []byte("test-secret"), nil)

	handler := CookieAuthenticator(ja)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no cookie, got %d", rr.Code)
	}
}

func TestCookieAuthenticator_InvalidToken(t *testing.T) {
	ja := jwtauth.New("HS256", []byte("test-secret"), nil)

	handler := CookieAuthenticator(ja)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "not.a.valid.jwt"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with invalid token, got %d", rr.Code)
	}
}

func TestCookieAuthenticator_ExpiredToken(t *testing.T) {
	ja := jwtauth.New("HS256", []byte("test-secret"), nil)
	_, tokenStr, _ := ja.Encode(map[string]interface{}{
		"user_id": 1,
		"exp":     1, // Unix epoch 1 = long expired
	})

	handler := CookieAuthenticator(ja)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenStr})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with expired token, got %d", rr.Code)
	}
}
