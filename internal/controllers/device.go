package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/lemon3studio/leo/internal/model"
	"github.com/lemon3studio/leo/internal/repos"
	"github.com/lemon3studio/leo/internal/services"

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
	cryptoService services.CryptoService
}

// NewDeviceController creates a new DeviceController instance
func NewDeviceController(deviceRepo repos.DeviceRepo, cryptoService services.CryptoService) DeviceController {
	return &deviceController{
		devices:       deviceRepo,
		cryptoService: cryptoService,
	}
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
	json.NewEncoder(w).Encode(d.ToDTO())
}

func (dc *deviceController) FindByID(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	d, err := dc.devices.FindByID(id)
	if err != nil {
		log.Printf("failed to find device by id %d: %v", id, err)
		http.Error(w, "Unable to find device in db", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(d.ToDTO())
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

	json.NewEncoder(w).Encode(d.ToDTO())
}

func (dc *deviceController) FindByUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from url params
	userID, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil {
		log.Printf("unable to parse userId from path parameter: %v", err)
		http.Error(w, "Unable to parse userId from path parameter", http.StatusBadRequest)
		return
	}

	// Find devices in db
	d, err := dc.devices.FindByUser(userID)
	if err != nil {
		log.Printf("failed to find devices for user %d: %v", userID, err)
		http.Error(w, "Unable to find devices in db", http.StatusNotFound)
		return
	}

	// Convert to DTOs with counts instead of arrays
	var deviceDTOs []model.DeviceDTO
	if d == nil {
		deviceDTOs = []model.DeviceDTO{}
	} else {
		deviceDTOs = make([]model.DeviceDTO, len(d))
		for i, device := range d {
			deviceDTOs[i] = device.ToDTO()
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

	// Update device in db
	if err := dc.devices.Update(id, &d); err != nil {
		log.Printf("failed to update device with id %d: %v", id, err)
		http.Error(w, "Unable to update device", http.StatusInternalServerError)
		return
	}

	// Respond with JSON encoded device DTO
	json.NewEncoder(w).Encode(d.ToDTO())
}

func (dc *deviceController) Delete(w http.ResponseWriter, r *http.Request) {
	// Get id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
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
