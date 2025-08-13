package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/lemon3studio/leo/internal/model"
	"github.com/lemon3studio/leo/internal/repos"

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

// listController implements the ListController interface
type listController struct {
	lists repos.ListRepo
}

// NewListController creates a new ListController instance
func NewListController(listRepo repos.ListRepo) ListController {
	return &listController{
		lists: listRepo,
	}
}

func (lc *listController) Create(w http.ResponseWriter, r *http.Request) {
	// Initialize list
	var l model.List

	// Encode list from request body
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		log.Printf("unable to decode list from request body: %v", err)
		http.Error(w, "Unable to decode list from request body", http.StatusBadRequest)
		return
	}

	// Store list into db
	lc.lists.Create(&l)

	// Respond with JSON encoded list DTO
	json.NewEncoder(w).Encode(l.ToDTO())
}

func (lc *listController) FindByID(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	// Find list in db
	l, err := lc.lists.FindByID(id)
	if err != nil {
		log.Printf("unable to find blocklist in db: %v", err)
		http.Error(w, "Unable to find blocklist in db", http.StatusNotFound)
		return
	}

	// Respond with JSON encoded list DTO
	json.NewEncoder(w).Encode(l.ToDTO())
}

func (lc *listController) FindByUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from url params
	userID, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil {
		log.Printf("unable to parse userId from path parameter: %v", err)
		http.Error(w, "Unable to parse userId from path parameter", http.StatusBadRequest)
		return
	}

	// Find lists in db
	l, err := lc.lists.FindByUser(userID)
	if err != nil {
		log.Printf("unable to find blocklists in db: %v", err)
		http.Error(w, "Unable to find blocklists in db", http.StatusNotFound)
		return
	}

	// Convert to DTOs with counts instead of arrays
	var listDTOs []model.ListDTO
	if l == nil {
		listDTOs = []model.ListDTO{}
	} else {
		listDTOs = make([]model.ListDTO, len(l))
		for i, list := range l {
			listDTOs[i] = list.ToDTO()
		}
	}

	// Respond with JSON encoded list DTOs
	json.NewEncoder(w).Encode(listDTOs)
}

func (lc *listController) Update(w http.ResponseWriter, r *http.Request) {
	// Initialize list
	var l model.List

	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	// Encode list from request body
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		log.Printf("unable to decode list from request body: %v", err)
		http.Error(w, "Unable to decode list from request body", http.StatusBadRequest)
		return
	}

	// Update list in db
	lc.lists.Update(id, &l)

	// Respond with JSON encoded list DTO
	json.NewEncoder(w).Encode(l.ToDTO())
}

func (lc *listController) Delete(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	// Delete blocklist from db
	lc.lists.Delete(id)

	// Respond with status no content
	w.WriteHeader(http.StatusNoContent)
}
