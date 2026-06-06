package controllers

import (
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/lemon3studio/blkhole/internal/services"
)

// AuthController defines the interface for authentication operations
type AuthController interface {
	Login(http.ResponseWriter, *http.Request)
	Register(http.ResponseWriter, *http.Request)
	RefreshToken(http.ResponseWriter, *http.Request)
	Logout(http.ResponseWriter, *http.Request)
	GetCurrentUser(http.ResponseWriter, *http.Request)
}

// authController implements the AuthController interface
type authController struct {
	authService services.AuthService
	listService services.ListService
}

// NewAuthController creates a new authentication controller
func NewAuthController(authService services.AuthService, listService services.ListService) AuthController {
	return &authController{authService: authService, listService: listService}
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

// Register handles user registration
func (c *authController) Register(w http.ResponseWriter, r *http.Request) {
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

	if len(req.Password) < 12 {
		http.Error(w, "Password must be at least 12 characters", http.StatusBadRequest)
		return
	}

	if pwned, err := isPasswordPwned(req.Password); err == nil && pwned {
		http.Error(w, "Password has appeared in known data breaches, please choose a different one", http.StatusBadRequest)
		return
	}

	result, err := c.authService.Register(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	go func() {
		if err := c.listService.SeedDefaults(result.User.ID); err != nil {
			log.Printf("failed to seed default lists for user %d: %v", result.User.ID, err)
		}
	}()

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
		Secure:   false,                // Set to false for development (HTTP)
		SameSite: http.SameSiteLaxMode, // CSRF protection but allows same-site navigation
	}
	http.SetCookie(w, cookie)
}

// isPasswordPwned checks the HIBP Pwned Passwords API using k-anonymity.
// Returns (true, nil) if the password appears in known breaches.
// Returns (false, nil) if it does not, or (false, err) if the API is unreachable (fail open).
func isPasswordPwned(password string) (bool, error) {
	sum := sha1.Sum([]byte(password))
	hash := strings.ToUpper(fmt.Sprintf("%x", sum))
	prefix, suffix := hash[:5], hash[5:]

	req, err := http.NewRequest(http.MethodGet, "https://api.pwnedpasswords.com/range/"+prefix, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Add-Padding", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("hibp returned %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], suffix) && parts[1] != "0" {
			return true, nil
		}
	}
	return false, scanner.Err()
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
