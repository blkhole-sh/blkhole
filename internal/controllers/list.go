package controllers

import (
	"log"
	"net/http"
	"github.com/lemon3studio/leo/internal/model"
	"github.com/lemon3studio/leo/internal/repos"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/vmihailenco/msgpack/v5"
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
	if err := msgpack.NewDecoder(r.Body).Decode(&l); err != nil {
		log.Printf("unable to decode list from request body: %v", err)
		http.Error(w, "Unable to decode list from request body", http.StatusBadRequest)
		return
	}

	// Store list into db
	lc.lists.Create(&l)

	// Respond with msgpack encoded list
	msgpack.NewEncoder(w).Encode(l)
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

	// Respond with msgpack encoded list
	msgpack.NewEncoder(w).Encode(l)
}

func (lc *listController) FindByUser(w http.ResponseWriter, r *http.Request) {
	// Get user hash from url params
	userHash := chi.URLParam(r, "userHash")

	// Find lists in db
	l, err := lc.lists.FindByUser(userHash)
	if err != nil {
		log.Printf("unable to find blocklists in db: %v", err)
		http.Error(w, "Unable to find blocklists in db", http.StatusNotFound)
		return
	}

	// Ensure we return an empty array instead of null for empty results
	if l == nil {
		l = []*model.List{}
	}

	// Respond with msgpack encoded lists
	msgpack.NewEncoder(w).Encode(l)
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
	if err := msgpack.NewDecoder(r.Body).Decode(&l); err != nil {
		log.Printf("unable to decode list from request body: %v", err)
		http.Error(w, "Unable to decode list from request body", http.StatusBadRequest)
		return
	}

	// Update list in db
	lc.lists.Update(id, &l)

	// Respond with msgpack encoded list
	msgpack.NewEncoder(w).Encode(l)
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
