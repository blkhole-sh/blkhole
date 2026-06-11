package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/blkhole-sh/blkhole/internal/model"
	"github.com/blkhole-sh/blkhole/internal/repos"
	"github.com/blkhole-sh/blkhole/internal/services"

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
	lists       repos.ListRepo
	listService services.ListService
	authService services.AuthService
}

// NewListController creates a new ListController instance
func NewListController(listRepo repos.ListRepo, listService services.ListService, authService services.AuthService) ListController {
	return &listController{
		lists:       listRepo,
		listService: listService,
		authService: authService,
	}
}

// requireList loads the list with the given id and verifies it belongs to the
// authenticated user. On failure it writes an error response and returns false.
func (lc *listController) requireList(w http.ResponseWriter, r *http.Request, id int) (*model.List, bool) {
	user, ok := currentUser(w, r, lc.authService)
	if !ok {
		return nil, false
	}

	l, err := lc.lists.FindByID(id)
	if err != nil {
		log.Printf("unable to find blocklist in db: %v", err)
		http.Error(w, "Unable to find blocklist in db", http.StatusNotFound)
		return nil, false
	}

	if l.UserID != user.ID {
		log.Printf("user %d attempted to access list %d owned by %d", user.ID, l.ID, l.UserID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil, false
	}

	return l, true
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

	// Always create the list for the authenticated user
	user, ok := currentUser(w, r, lc.authService)
	if !ok {
		return
	}
	l.UserID = user.ID

	// Store list into db
	if _, err := lc.lists.Create(&l); err != nil {
		log.Printf("failed to create list: %v", err)
		http.Error(w, "Unable to create list", http.StatusInternalServerError)
		return
	}

	// Load domains from source if set
	if l.Source != "" {
		if err := lc.listService.LoadList(&l); err != nil {
			log.Printf("failed to load list %d from source: %v", l.ID, err)
		}
	}

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

	// Find list in db and verify ownership
	l, ok := lc.requireList(w, r, id)
	if !ok {
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

	// Users may only list their own blocklists
	user, ok := currentUser(w, r, lc.authService)
	if !ok {
		return
	}
	if userID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
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

	// Verify list belongs to the authenticated user
	if _, ok := lc.requireList(w, r, id); !ok {
		return
	}

	// Update list in db
	if err := lc.lists.Update(id, &l); err != nil {
		log.Printf("failed to update list with id %d: %v", id, err)
		http.Error(w, "Unable to update list", http.StatusInternalServerError)
		return
	}

	// Fetch the updated list with all relations
	updatedList, err := lc.lists.FindByID(id)
	if err != nil {
		log.Printf("failed to fetch updated list with id %d: %v", id, err)
		http.Error(w, "Unable to fetch updated list", http.StatusInternalServerError)
		return
	}

	// Load domains from source if set
	if updatedList.Source != "" {
		if err := lc.listService.LoadList(updatedList); err != nil {
			log.Printf("failed to load list %d from source: %v", id, err)
		}
	}

	// Respond with JSON encoded list DTO
	json.NewEncoder(w).Encode(updatedList.ToDTO())
}

func (lc *listController) Delete(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	l, ok := lc.requireList(w, r, id)
	if !ok {
		return
	}

	if l.IsDefault {
		http.Error(w, "Default lists cannot be deleted", http.StatusForbidden)
		return
	}

	// Delete blocklist from db
	lc.lists.Delete(id)

	// Respond with status no content
	w.WriteHeader(http.StatusNoContent)
}
