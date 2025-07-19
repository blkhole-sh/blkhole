package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"server/model"
	"server/repos"
	"server/services"
	"strconv"

	"github.com/go-chi/chi/v5"
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

// ScheduleControllerImpl implements the ScheduleController interface
type ScheduleControllerImpl struct {
	scheduleRepo   repos.ScheduleRepo
	contentBlocker services.ContentBlocker
}

// NewScheduleController creates a new ScheduleController instance
func NewScheduleController(scheduleRepo repos.ScheduleRepo, contentBlocker services.ContentBlocker) ScheduleController {
	return &ScheduleControllerImpl{
		contentBlocker: contentBlocker,
		scheduleRepo:   scheduleRepo,
	}
}

// IsBlocked handles GET requests to check if a domain is blocked
func (sc *ScheduleControllerImpl) IsBlocked(w http.ResponseWriter, r *http.Request) {
	// Parse domain from request param
	domain := r.URL.Query().Get("domain")
	deviceHash := r.URL.Query().Get("deviceHash")

	// Check if domain is blocked, throw error if domain is invalid
	blocked, err := sc.contentBlocker.IsBlocked(domain, deviceHash)
	if err != nil {
		log.Fatal(err)

		// Return HTTP Bad Gateway if domain is invalid
		http.Error(w, "Invalid domain", http.StatusBadGateway)
		return
	}

	// Return HTTP 200 and true if domain is blocked or false if else
	w.Write([]byte(strconv.FormatBool(blocked)))
}

func (sc *ScheduleControllerImpl) Create(w http.ResponseWriter, r *http.Request) {
	// Initialize schedule
	var s model.Schedule

	// Encode schedule from request body
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to decode schedule from request body", http.StatusBadRequest)
	}

	// Store schedule into db
	sc.scheduleRepo.Create(&s)

	// Respond with json encoded schedule
	json.NewEncoder(w).Encode(s)
}

func (sc *ScheduleControllerImpl) FindByID(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
	}

	// Find schedule in db
	s, err := sc.scheduleRepo.FindByID(id)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to find schedule in db", http.StatusNotFound)
	}

	// Respond with json encoded schedule
	json.NewEncoder(w).Encode(s)
}

func (sc *ScheduleControllerImpl) FindByUser(w http.ResponseWriter, r *http.Request) {
	// Get user hash from url params
	userHash := chi.URLParam(r, "userHash")

	// Find schedules in db
	s, err := sc.scheduleRepo.FindByUser(userHash)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to find schedules in db", http.StatusNotFound)
	}

	// Respond with json encoded schedules
	json.NewEncoder(w).Encode(s)
}

func (sc *ScheduleControllerImpl) Update(w http.ResponseWriter, r *http.Request) {
	// Initialize schedule
	var s model.Schedule

	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
	}

	// Encode schedule from request body
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to decode schedule from request body", http.StatusBadRequest)
	}

	// Update schedule in db
	sc.scheduleRepo.Update(id, &s)

	// Respond with json encoded schedule
	json.NewEncoder(w).Encode(s)
}

func (sc *ScheduleControllerImpl) Delete(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
	}

	// Delete schedule from db
	sc.scheduleRepo.Delete(id)

	// Respond with status no content
	w.WriteHeader(http.StatusNoContent)
}
