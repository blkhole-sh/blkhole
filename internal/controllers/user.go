// Package controllers provides HTTP request handlers for the Leo DNS blocker API.
package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/lemon3studio/leo/internal/model"
	"github.com/lemon3studio/leo/internal/repos"
	"github.com/lemon3studio/leo/internal/services"

	"github.com/go-chi/chi/v5"
)

// UserController defines the interface for user operations
type UserController interface {
	Create(http.ResponseWriter, *http.Request)
	FindByHash(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
}

// userController implements the UserController interface
type userController struct {
	users         repos.UserRepo
	cryptoService services.CryptoService
}

// NewUserController creates a new UserController instance
func NewUserController(userRepo repos.UserRepo, cryptoService services.CryptoService) UserController {
	return &userController{
		users:         userRepo,
		cryptoService: cryptoService,
	}
}

func (uc *userController) Create(w http.ResponseWriter, r *http.Request) {
	// Initialize user
	var u model.User

	// Encode user from request body
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		log.Printf("unable to decode user from request body: %v", err)
		http.Error(w, "Unable to decode user from request body", http.StatusBadRequest)
		return
	}

	hash, err := uc.cryptoService.RandomHash()
	if err != nil {
		log.Printf("unable to create hash for user: %v", err)
		http.Error(w, "Unable to create hash for user", http.StatusInternalServerError)
		return
	}

	u.Hash = hash

	// Store user into db
	uc.users.Create(&u)

	// Respond with JSON encoded user
	json.NewEncoder(w).Encode(u)
}

func (uc *userController) FindByHash(w http.ResponseWriter, r *http.Request) {
	// Get hash from url params
	hash := chi.URLParam(r, "hash")

	u, err := uc.users.FindByHash(hash)
	if err != nil {
		log.Printf("unable to find user in db: %v", err)
		http.Error(w, "Unable to find user in db", http.StatusNotFound)
		return
	}

	// Respond with JSON encoded user
	json.NewEncoder(w).Encode(u)
}

func (uc *userController) Update(w http.ResponseWriter, r *http.Request) {
	// Initialize user
	var u model.User

	// Get hash from url params
	hash := chi.URLParam(r, "hash")

	// Encode user from request body
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		log.Printf("unable to decode user from request body: %v", err)
		http.Error(w, "Unable to decode user from request body", http.StatusBadRequest)
		return
	}

	// Update user in db
	uc.users.Update(hash, &u)

	// Respond with JSON encoded user
	json.NewEncoder(w).Encode(u)
}

func (uc *userController) Delete(w http.ResponseWriter, r *http.Request) {
	// Get hash from url params
	hash := chi.URLParam(r, "hash")

	// Delete user from db
	uc.users.Delete(hash)
	w.WriteHeader(http.StatusNoContent)

	// Respond with status no content
	w.WriteHeader(http.StatusNoContent)
}
