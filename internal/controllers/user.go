// Package controllers provides HTTP request handlers for the blkhole DNS blocker API.
package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/lemon3studio/blkhole/internal/model"
	"github.com/lemon3studio/blkhole/internal/repos"
	"github.com/lemon3studio/blkhole/internal/services"

	"github.com/go-chi/chi/v5"
)

// UserController defines the interface for user operations
type UserController interface {
	FindByID(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
}

// userController implements the UserController interface
type userController struct {
	users         repos.UserRepo
	authService   services.AuthService
	cryptoService services.CryptoService
}

// NewUserController creates a new UserController instance
func NewUserController(userRepo repos.UserRepo, authService services.AuthService, cryptoService services.CryptoService) UserController {
	return &userController{
		users:         userRepo,
		authService:   authService,
		cryptoService: cryptoService,
	}
}

func (uc *userController) FindByID(w http.ResponseWriter, r *http.Request) {
	// Get current user
	currentUser, err := uc.authService.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	if currentUser.ID != id {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	u, err := uc.users.FindByID(id)
	if err != nil {
		log.Printf("unable to find user in db: %v", err)
		http.Error(w, "Unable to find user in db", http.StatusNotFound)
		return
	}

	// Respond with JSON encoded user
	json.NewEncoder(w).Encode(u.ToDTO())
}

func (uc *userController) Update(w http.ResponseWriter, r *http.Request) {
	// Get current user
	currentUser, err := uc.authService.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	if currentUser.ID != id {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Decode request body — supports optional password change fields
	var req struct {
		model.User
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("unable to decode user from request body: %v", err)
		http.Error(w, "Unable to decode user from request body", http.StatusBadRequest)
		return
	}

	u := req.User

	// If a password change is requested, verify current password first
	if req.CurrentPassword != "" || req.NewPassword != "" {
		if req.CurrentPassword == "" || req.NewPassword == "" {
			http.Error(w, "Both currentPassword and newPassword are required", http.StatusBadRequest)
			return
		}

		valid, err := uc.cryptoService.VerifyPassword(req.CurrentPassword, currentUser.PasswordHash)
		if err != nil || !valid {
			http.Error(w, "Current password is incorrect", http.StatusBadRequest)
			return
		}

		newHash, err := uc.cryptoService.HashPassword(req.NewPassword)
		if err != nil {
			log.Printf("failed to hash new password for user %d: %v", id, err)
			http.Error(w, "Unable to update password", http.StatusInternalServerError)
			return
		}

		u.PasswordHash = newHash
	} else {
		// Preserve existing password hash when not changing password
		u.PasswordHash = currentUser.PasswordHash
	}

	// Preserve name and email from current user if not provided
	if u.Name == "" {
		u.Name = currentUser.Name
	}
	if u.Email == "" {
		u.Email = currentUser.Email
	}

	// Update user in db
	if err := uc.users.Update(id, &u); err != nil {
		log.Printf("failed to update user with id %d: %v", id, err)
		http.Error(w, "Unable to update user", http.StatusInternalServerError)
		return
	}

	// Load relations for the DTO
	if err := uc.users.LoadRelations(&u); err != nil {
		log.Printf("failed to load relations for user %d: %v", id, err)
	}
	u.ID = id

	// Respond with JSON encoded user
	json.NewEncoder(w).Encode(u.ToDTO())
}

func (uc *userController) Delete(w http.ResponseWriter, r *http.Request) {
	// Get current user
	currentUser, err := uc.authService.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	if currentUser.ID != id {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Delete user from db
	if err := uc.users.Delete(id); err != nil {
		log.Printf("failed to delete user with id %d: %v", id, err)
		http.Error(w, "Unable to delete user", http.StatusInternalServerError)
		return
	}

	// Respond with status no content
	w.WriteHeader(http.StatusNoContent)
}
