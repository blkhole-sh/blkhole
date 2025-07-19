package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"server/model"
	"server/repos"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// ListController defines the interface for list operations
type ListController interface {
	Create(http.ResponseWriter, *http.Request)
	FindByID(http.ResponseWriter, *http.Request)
	FindByUser(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
}

// ListControllerImpl implements the ListController interface
type ListControllerImpl struct {
	listRepo repos.ListRepo
}

// NewListController creates a new ListController instance
func NewListController(listRepo repos.ListRepo) ListController {
	return &ListControllerImpl{
		listRepo: listRepo,
	}
}

func (lc *ListControllerImpl) Create(w http.ResponseWriter, r *http.Request) {
	// Initialize list
	var l model.List

	// Encode list from request body
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to decode list from request body", http.StatusBadRequest)
	}

	// Store list into db
	lc.listRepo.Create(&l)

	// Respond with json encoded list
	json.NewEncoder(w).Encode(l)
}

func (lc *ListControllerImpl) FindByID(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
	}

	// Find list in db
	l, err := lc.listRepo.FindByID(id)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to find blocklist in db", http.StatusNotFound)
	}

	// Respond with json encoded list
	json.NewEncoder(w).Encode(l)
}

func (lc *ListControllerImpl) FindByUser(w http.ResponseWriter, r *http.Request) {
	// Get user hash from url params
	userHash := chi.URLParam(r, "userHash")

	// Find lists in db
	l, err := lc.listRepo.FindByUser(userHash)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to find blocklists in db", http.StatusNotFound)
	}

	// Respond with json encoded lists
	json.NewEncoder(w).Encode(l)
}

func (lc *ListControllerImpl) Update(w http.ResponseWriter, r *http.Request) {
	// Initialize list
	var l model.List

	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
	}

	// Encode list from request body
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to decode list from request body", http.StatusBadRequest)
	}

	// Update list in db
	lc.listRepo.Update(id, &l)

	// Respond with json encoded list
	json.NewEncoder(w).Encode(l)
}

func (lc *ListControllerImpl) Delete(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
	}

	// Delete blocklist from db
	lc.listRepo.Delete(id)

	// Respond with status no content
	w.WriteHeader(http.StatusNoContent)
}
