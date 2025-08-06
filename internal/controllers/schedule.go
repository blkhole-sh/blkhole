package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/lemon3studio/leo/internal/model"
	"github.com/lemon3studio/leo/internal/repos"
	"github.com/lemon3studio/leo/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/vmihailenco/msgpack/v5"
)

// ScheduleController defines the interface for schedule operations
type ScheduleController interface {
	IsBlocked(http.ResponseWriter, *http.Request)
	Create(http.ResponseWriter, *http.Request)
	FindByID(http.ResponseWriter, *http.Request)
	FindByUser(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
}

// scheduleController implements the ScheduleController interface
type scheduleController struct {
	schedules      repos.ScheduleRepo
	contentBlocker services.ContentBlocker
}

// NewScheduleController creates a new ScheduleController instance
func NewScheduleController(scheduleRepo repos.ScheduleRepo, contentBlocker services.ContentBlocker) ScheduleController {
	return &scheduleController{
		contentBlocker: contentBlocker,
		schedules:      scheduleRepo,
	}
}

// IsBlocked handles GET requests to check if a domain is blocked
func (sc *scheduleController) IsBlocked(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	domain := r.URL.Query().Get("domain")
	deviceHash := r.URL.Query().Get("deviceHash")

	// Check if the domain is blocked
	blocked, err := sc.contentBlocker.IsBlocked(domain, deviceHash)
	if err != nil {
		http.Error(w, "Invalid domain", http.StatusBadRequest)
		return
	}

	// Set response content type and return JSON result
	msgpack.NewEncoder(w).Encode(map[string]bool{"blocked": blocked})
}

func (sc *scheduleController) Create(w http.ResponseWriter, r *http.Request) {
	// Initialize schedule
	var s model.Schedule

	// Encode schedule from request body
	if err := msgpack.NewDecoder(r.Body).Decode(&s); err != nil {
		log.Printf("unable to decode schedule from request body: %v", err)
		http.Error(w, "Unable to decode schedule from request body", http.StatusBadRequest)
		return
	}

	// Store schedule into db
	sc.schedules.Create(&s)

	// Respond with msgpack encoded schedule
	msgpack.NewEncoder(w).Encode(s)
}

func (sc *scheduleController) FindByID(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	// Find schedule in db
	s, err := sc.schedules.FindByID(id)
	if err != nil {
		log.Printf("unable to find schedule in db: %v", err)
		http.Error(w, "Unable to find schedule in db", http.StatusNotFound)
		return
	}

	// Respond with msgpack encoded schedule
	msgpack.NewEncoder(w).Encode(s)
}

func (sc *scheduleController) FindByUser(w http.ResponseWriter, r *http.Request) {
	// Get user hash from url params
	userHash := chi.URLParam(r, "userHash")

	// Find schedules in db
	s, err := sc.schedules.FindByUser(userHash)
	if err != nil {
		log.Printf("unable to find schedules in db: %v", err)
		http.Error(w, "Unable to find schedules in db", http.StatusNotFound)
		return
	}

	// Ensure we return an empty array instead of null for empty results
	if s == nil {
		s = []*model.Schedule{}
	}

	// Respond with msgpack encoded schedules
	msgpack.NewEncoder(w).Encode(s)
}

func (sc *scheduleController) Update(w http.ResponseWriter, r *http.Request) {
	// Initialize schedule
	var s model.Schedule

	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	// Encode schedule from request body
	if err := msgpack.NewDecoder(r.Body).Decode(&s); err != nil {
		log.Printf("unable to decode schedule from request body: %v", err)
		http.Error(w, "Unable to decode schedule from request body", http.StatusBadRequest)
		return
	}

	// Update schedule in db
	sc.schedules.Update(id, &s)

	// Respond with msgpack encoded schedule
	msgpack.NewEncoder(w).Encode(s)
}

func (sc *scheduleController) Delete(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	// Delete schedule from db
	sc.schedules.Delete(id)

	// Respond with status no content
	w.WriteHeader(http.StatusNoContent)
}
