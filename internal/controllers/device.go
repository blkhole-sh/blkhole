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

// DeviceController defines the interface for device operations
type DeviceController interface {
	Create(http.ResponseWriter, *http.Request)
	FindByID(http.ResponseWriter, *http.Request)
	FindByHash(http.ResponseWriter, *http.Request)
	FindByUser(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
}

// deviceController implements the DeviceController interface
type deviceController struct {
	devices       repos.DeviceRepo
	schedules     repos.ScheduleRepo
	cryptoService services.CryptoService
	authService   services.AuthService
}

// NewDeviceController creates a new DeviceController instance
func NewDeviceController(deviceRepo repos.DeviceRepo, scheduleRepo repos.ScheduleRepo, cryptoService services.CryptoService, authService services.AuthService) DeviceController {
	return &deviceController{
		devices:       deviceRepo,
		schedules:     scheduleRepo,
		cryptoService: cryptoService,
		authService:   authService,
	}
}

// requireDevice loads the device with the given id and verifies it belongs to
// the authenticated user. On failure it writes an error response and returns false.
func (dc *deviceController) requireDevice(w http.ResponseWriter, r *http.Request, id int) (*model.Device, bool) {
	user, ok := currentUser(w, r, dc.authService)
	if !ok {
		return nil, false
	}

	d, err := dc.devices.FindByID(id)
	if err != nil {
		log.Printf("failed to find device by id %d: %v", id, err)
		http.Error(w, "Unable to find device in db", http.StatusNotFound)
		return nil, false
	}

	if d.UserID != user.ID {
		log.Printf("user %d attempted to access device %d owned by %d", user.ID, d.ID, d.UserID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil, false
	}

	return d, true
}

func (dc *deviceController) Create(w http.ResponseWriter, r *http.Request) {
	// Initialize device
	var d model.Device

	// Encode device from request body
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		log.Printf("failed to decode device from request body: %v", err)
		http.Error(w, "Unable to decode device from request body", http.StatusBadRequest)
		return
	}

	// Always create the device for the authenticated user
	user, ok := currentUser(w, r, dc.authService)
	if !ok {
		return
	}
	d.UserID = user.ID

	hash, err := dc.cryptoService.RandomHash()
	if err != nil {
		log.Printf("failed to create hash for device: %v", err)
		http.Error(w, "Unable to create hash for user", http.StatusInternalServerError)
		return
	}

	d.Hash = hash

	// Store device into db
	if err := dc.devices.Create(&d); err != nil {
		log.Printf("failed to create device: %v", err)
		http.Error(w, "Unable to create device", http.StatusInternalServerError)
		return
	}

	// Respond with JSON encoded device DTO
	json.NewEncoder(w).Encode(d.ToDTO(nil, nil))
}

func (dc *deviceController) FindByID(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	d, ok := dc.requireDevice(w, r, id)
	if !ok {
		return
	}

	json.NewEncoder(w).Encode(d.ToDTO(nil, nil))
}

func (dc *deviceController) FindByHash(w http.ResponseWriter, r *http.Request) {
	// Get hash from url params
	hash := chi.URLParam(r, "hash")

	d, err := dc.devices.FindByHash(hash)
	if err != nil {
		log.Printf("failed to find device by hash %s: %v", hash, err)
		http.Error(w, "Unable to find device in db", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(d.ToDTO(nil, nil))
}

func (dc *deviceController) FindByUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from url params
	userID, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil {
		log.Printf("unable to parse userId from path parameter: %v", err)
		http.Error(w, "Unable to parse userId from path parameter", http.StatusBadRequest)
		return
	}

	// Users may only list their own devices
	user, ok := currentUser(w, r, dc.authService)
	if !ok {
		return
	}
	if userID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Find devices in db
	d, err := dc.devices.FindByUser(userID)
	if err != nil {
		log.Printf("failed to find devices for user %d: %v", userID, err)
		http.Error(w, "Unable to find devices in db", http.StatusNotFound)
		return
	}

	// Convert to DTOs with schedule IDs and names
	var deviceDTOs []model.DeviceDTO
	if d == nil {
		deviceDTOs = []model.DeviceDTO{}
	} else {
		deviceDTOs = make([]model.DeviceDTO, len(d))
		for i, device := range d {
			ids, err := dc.devices.LoadScheduleIDs(device.ID)
			if err != nil {
				log.Printf("failed to load schedule IDs for device %d: %v", device.ID, err)
				ids = []int{}
			}
			names, err := dc.schedules.FindNamesByDeviceID(device.ID)
			if err != nil {
				log.Printf("failed to load schedule names for device %d: %v", device.ID, err)
				names = []string{}
			}
			deviceDTOs[i] = device.ToDTO(ids, names)
		}
	}

	// Respond with JSON encoded device DTOs
	json.NewEncoder(w).Encode(deviceDTOs)
}

func (dc *deviceController) Update(w http.ResponseWriter, r *http.Request) {
	// Initialize device
	var d model.Device

	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	// Encode device from request body
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		log.Printf("failed to decode device from request body: %v", err)
		http.Error(w, "Unable to decode device from request body", http.StatusBadRequest)
		return
	}

	// Verify device belongs to the authenticated user
	if _, ok := dc.requireDevice(w, r, id); !ok {
		return
	}

	// Update device in db
	if err := dc.devices.Update(id, &d); err != nil {
		log.Printf("failed to update device with id %d: %v", id, err)
		http.Error(w, "Unable to update device", http.StatusInternalServerError)
		return
	}

	// Fetch the updated device from database
	updatedDevice, err := dc.devices.FindByID(id)
	if err != nil {
		log.Printf("failed to fetch updated device with id %d: %v", id, err)
		http.Error(w, "Unable to fetch updated device", http.StatusInternalServerError)
		return
	}

	// Load schedule IDs and names for the device
	scheduleIDs, err := dc.devices.LoadScheduleIDs(id)
	if err != nil {
		log.Printf("failed to load schedule IDs for device %d: %v", id, err)
		scheduleIDs = []int{}
	}
	scheduleNames, err := dc.schedules.FindNamesByDeviceID(id)
	if err != nil {
		log.Printf("failed to load schedule names for device %d: %v", id, err)
		scheduleNames = []string{}
	}

	// Respond with JSON encoded device DTO
	json.NewEncoder(w).Encode(updatedDevice.ToDTO(scheduleIDs, scheduleNames))
}

func (dc *deviceController) Delete(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	// Verify device belongs to the authenticated user
	if _, ok := dc.requireDevice(w, r, id); !ok {
		return
	}

	// Delete device from db
	if err := dc.devices.Delete(id); err != nil {
		log.Printf("failed to delete device with id %d: %v", id, err)
		http.Error(w, "Unable to delete device", http.StatusInternalServerError)
		return
	}

	// Respond with status no content
	w.WriteHeader(http.StatusNoContent)
}
