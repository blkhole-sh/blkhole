package controllers

import (
	"encoding/json"
	"net/http"
	"server/internal/services"
	"strings"
)

// AuthController defines the interface for authentication operations
type AuthController interface {
	Login(http.ResponseWriter, *http.Request)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetCurrentUser returns the current authenticated user
func (c *authController) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := c.authService.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

