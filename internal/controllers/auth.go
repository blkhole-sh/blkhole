package controllers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/lemon3studio/leo/internal/services"
)

// AuthController defines the interface for authentication operations
type AuthController interface {
	Login(http.ResponseWriter, *http.Request)
	RefreshToken(http.ResponseWriter, *http.Request)
	Logout(http.ResponseWriter, *http.Request)
	GetCurrentUser(http.ResponseWriter, *http.Request)
}

// authController implements the AuthController interface
type authController struct {
	authService services.AuthService
}

// NewAuthController creates a new authentication controller
func NewAuthController(authService services.AuthService) AuthController {
	return &authController{authService: authService}
}

// Login handles user authentication
func (c *authController) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		http.Error(w, "Email and password required", http.StatusBadRequest)
		return
	}

	result, err := c.authService.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Set secure HttpOnly cookies
	c.setSecureCookie(w, "access_token", result.AccessToken, services.TokenExpiry)
	c.setSecureCookie(w, "refresh_token", result.RefreshToken, services.RefreshTokenExpiry)

	// Return only user data (tokens are in cookies)
	json.NewEncoder(w).Encode(map[string]any{"user": result.User.ToDTO()})
}

// RefreshToken handles token refresh
func (c *authController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Refresh token not found", http.StatusUnauthorized)
		return
	}

	result, err := c.authService.RefreshToken(refreshCookie.Value)
	if err != nil {
		// Clear invalid cookies
		c.clearAuthCookies(w)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Set new secure HttpOnly cookies
	c.setSecureCookie(w, "access_token", result.AccessToken, services.TokenExpiry)
	c.setSecureCookie(w, "refresh_token", result.RefreshToken, services.RefreshTokenExpiry)

	// Return only user data
	json.NewEncoder(w).Encode(map[string]any{"user": result.User.ToDTO()})
}

// Logout handles user logout
func (c *authController) Logout(w http.ResponseWriter, r *http.Request) {
	c.clearAuthCookies(w)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// GetCurrentUser returns the current authenticated user
func (c *authController) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := c.authService.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(user.ToDTO())
}

// setSecureCookie sets a secure HttpOnly cookie
func (c *authController) setSecureCookie(w http.ResponseWriter, name, value string, expiry time.Duration) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(expiry),
		HttpOnly: true,
		Secure:   false,                   // Set to false for development (HTTP)
		SameSite: http.SameSiteLaxMode, // CSRF protection but allows same-site navigation
	}
	http.SetCookie(w, cookie)
}

// clearAuthCookies clears authentication cookies
func (c *authController) clearAuthCookies(w http.ResponseWriter) {
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)
}
