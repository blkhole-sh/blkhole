// Package controllers provides HTTP request handlers for the Leo DNS blocker API.
package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"server/model"
	"server/repos"
	"server/services"

	"github.com/go-chi/chi/v5"
)

// UserController defines the interface for user operations
type UserController interface {
	Create(http.ResponseWriter, *http.Request)
	FindByHash(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
}

// UserControllerImpl implements the UserController interface
type UserControllerImpl struct {
	userRepo      repos.UserRepo
	cryptoService services.CryptoService
}

// NewUserController creates a new UserController instance
func NewUserController(userRepo repos.UserRepo, cryptoService services.CryptoService) UserController {
	return &UserControllerImpl{
		userRepo:      userRepo,
		cryptoService: cryptoService,
	}
}

func (uc *UserControllerImpl) Create(w http.ResponseWriter, r *http.Request) {
	// Initialize user
	var u model.User

	// Encode user from request body
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to decode user from request body", http.StatusBadRequest)
	}

	hash, err := uc.cryptoService.RandomHash()
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to create hash for user", http.StatusInternalServerError)
	}

	u.Hash = hash

	// Store user into db
	uc.userRepo.Create(&u)

	// Respond with json encoded user
	json.NewEncoder(w).Encode(u)
}

func (uc *UserControllerImpl) FindByHash(w http.ResponseWriter, r *http.Request) {
	// Get hash from url params
	hash := chi.URLParam(r, "hash")

	u, err := uc.userRepo.FindByHash(hash)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to find user in db", http.StatusNotFound)
	}

	// Respond with json encoded user
	json.NewEncoder(w).Encode(u)
}

func (uc *UserControllerImpl) Update(w http.ResponseWriter, r *http.Request) {
	// Initialize user
	var u model.User

	// Get hash from url params
	hash := chi.URLParam(r, "hash")

	// Encode user from request body
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to decode user from request body", http.StatusBadRequest)
	}

	// Update user in db
	uc.userRepo.Update(hash, &u)

	// Respond with json encoded user
	json.NewEncoder(w).Encode(u)
}

func (uc *UserControllerImpl) Delete(w http.ResponseWriter, r *http.Request) {
	// Get hash from url params
	hash := chi.URLParam(r, "hash")

	// Delete user from db
	uc.userRepo.Delete(hash)
	w.WriteHeader(http.StatusNoContent)

	// Respond with status no content
	w.WriteHeader(http.StatusNoContent)
}
