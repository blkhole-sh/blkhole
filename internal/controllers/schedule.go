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
	devices        repos.DeviceRepo
	lists          repos.ListRepo
	contentBlocker services.ContentBlocker
}

// NewScheduleController creates a new ScheduleController instance
func NewScheduleController(scheduleRepo repos.ScheduleRepo, deviceRepo repos.DeviceRepo, listRepo repos.ListRepo, contentBlocker services.ContentBlocker) ScheduleController {
	return &scheduleController{
		contentBlocker: contentBlocker,
		schedules:      scheduleRepo,
		devices:        deviceRepo,
		lists:          listRepo,
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

	// Return JSON result
	json.NewEncoder(w).Encode(map[string]bool{"blocked": blocked})
}

func (sc *scheduleController) Create(w http.ResponseWriter, r *http.Request) {
	// Initialize schedule
	var s model.Schedule

	// Encode schedule from request body
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		log.Printf("unable to decode schedule from request body: %v", err)
		http.Error(w, "Unable to decode schedule from request body", http.StatusBadRequest)
		return
	}

	// Store schedule into db
	sc.schedules.Create(&s)

	// Respond with JSON encoded schedule DTO
	json.NewEncoder(w).Encode(s.ToDTO(nil, nil))
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

	// Respond with JSON encoded schedule DTO
	json.NewEncoder(w).Encode(s.ToDTO(nil, nil))
}

func (sc *scheduleController) FindByUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from url params
	userID, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil {
		log.Printf("unable to parse userId from path parameter: %v", err)
		http.Error(w, "Unable to parse userId from path parameter", http.StatusBadRequest)
		return
	}

	// Find schedules in db
	s, err := sc.schedules.FindByUser(userID)
	if err != nil {
		log.Printf("unable to find schedules in db: %v", err)
		http.Error(w, "Unable to find schedules in db", http.StatusNotFound)
		return
	}

	// Convert to DTOs with device and list names
	var scheduleDTOs []model.ScheduleDTO
	if s == nil {
		scheduleDTOs = []model.ScheduleDTO{}
	} else {
		scheduleDTOs = make([]model.ScheduleDTO, len(s))
		for i, schedule := range s {
			deviceNames, err := sc.devices.FindNamesByScheduleID(schedule.ID)
			if err != nil {
				log.Printf("failed to load device names for schedule %d: %v", schedule.ID, err)
				deviceNames = []string{}
			}
			listNames, err := sc.lists.FindNamesByScheduleID(schedule.ID)
			if err != nil {
				log.Printf("failed to load list names for schedule %d: %v", schedule.ID, err)
				listNames = []string{}
			}
			scheduleDTOs[i] = schedule.ToDTO(deviceNames, listNames)
		}
	}

	// Respond with JSON encoded schedule DTOs
	json.NewEncoder(w).Encode(scheduleDTOs)
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
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		log.Printf("unable to decode schedule from request body: %v", err)
		http.Error(w, "Unable to decode schedule from request body", http.StatusBadRequest)
		return
	}

	existing, err := sc.schedules.FindByID(id)
	if err != nil {
		log.Printf("unable to find schedule with id %d: %v", id, err)
		http.Error(w, "Unable to find schedule", http.StatusNotFound)
		return
	}
	if existing.IsDefault && !s.Active {
		http.Error(w, "Default schedules cannot be deactivated", http.StatusForbidden)
		return
	}

	// Update schedule in db
	if err := sc.schedules.Update(id, &s); err != nil {
		log.Printf("failed to update schedule with id %d: %v", id, err)
		http.Error(w, "Unable to update schedule", http.StatusInternalServerError)
		return
	}

	// Fetch the updated schedule from database
	updatedSchedule, err := sc.schedules.FindByID(id)
	if err != nil {
		log.Printf("failed to fetch updated schedule with id %d: %v", id, err)
		http.Error(w, "Unable to fetch updated schedule", http.StatusInternalServerError)
		return
	}

	// Load device and list names for the schedule
	deviceNames, err := sc.devices.FindNamesByScheduleID(id)
	if err != nil {
		log.Printf("failed to load device names for schedule %d: %v", id, err)
		deviceNames = []string{}
	}
	listNames, err := sc.lists.FindNamesByScheduleID(id)
	if err != nil {
		log.Printf("failed to load list names for schedule %d: %v", id, err)
		listNames = []string{}
	}

	// Respond with JSON encoded schedule DTO
	json.NewEncoder(w).Encode(updatedSchedule.ToDTO(deviceNames, listNames))
}

func (sc *scheduleController) Delete(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	s, err := sc.schedules.FindByID(id)
	if err != nil {
		log.Printf("unable to find schedule with id %d: %v", id, err)
		http.Error(w, "Unable to find schedule", http.StatusNotFound)
		return
	}

	if s.IsDefault {
		http.Error(w, "Default schedules cannot be deleted", http.StatusForbidden)
		return
	}

	// Delete schedule from db
	sc.schedules.Delete(id)

	// Respond with status no content
	w.WriteHeader(http.StatusNoContent)
}
