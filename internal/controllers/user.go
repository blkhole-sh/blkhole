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
	Create(http.ResponseWriter, *http.Request)
	FindByID(http.ResponseWriter, *http.Request)
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

	// Store user into db
	if err := uc.users.Create(&u); err != nil {
		log.Printf("failed to create user: %v", err)
		http.Error(w, "Unable to create user", http.StatusInternalServerError)
		return
	}

	// Respond with JSON encoded user
	json.NewEncoder(w).Encode(u.ToDTO())
}

func (uc *userController) FindByID(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
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
	// Initialize user
	var u model.User

	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	// Encode user from request body
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		log.Printf("unable to decode user from request body: %v", err)
		http.Error(w, "Unable to decode user from request body", http.StatusBadRequest)
		return
	}

	// Update user in db
	if err := uc.users.Update(id, &u); err != nil {
		log.Printf("failed to update user with id %d: %v", id, err)
		http.Error(w, "Unable to update user", http.StatusInternalServerError)
		return
	}

	// Respond with JSON encoded user
	json.NewEncoder(w).Encode(u.ToDTO())
}

func (uc *userController) Delete(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
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
