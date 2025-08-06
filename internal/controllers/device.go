package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/lemon3studio/leo/internal/model"
	"github.com/lemon3studio/leo/internal/repos"
	"github.com/lemon3studio/leo/internal/services"

	"github.com/go-chi/chi/v5"
)

// DeviceController defines the interface for device operations
type DeviceController interface {
	Create(http.ResponseWriter, *http.Request)
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
	dc.devices.Create(&d)

	// Respond with msgpack encoded device
	json.NewEncoder(w).Encode(d)
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

	json.NewEncoder(w).Encode(d)
}

func (dc *deviceController) FindByUser(w http.ResponseWriter, r *http.Request) {
	// Get user hash from url params
	userHash := chi.URLParam(r, "userHash")

	// Find devices in db
	d, err := dc.devices.FindByUser(userHash)
	if err != nil {
		log.Printf("failed to find devices for user %s: %v", userHash, err)
		http.Error(w, "Unable to find devices in db", http.StatusNotFound)
		return
	}

	// Ensure we return an empty array instead of null for empty results
	if d == nil {
		d = []*model.Device{}
	}

	// Respond with msgpack encoded devices
	json.NewEncoder(w).Encode(d)
}

func (dc *deviceController) Update(w http.ResponseWriter, r *http.Request) {
	// Initialize device
	var d model.Device

	// Get hash from url params
	hash := chi.URLParam(r, "hash")

	// Encode device from request body
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		log.Printf("failed to decode device from request body: %v", err)
		http.Error(w, "Unable to decode device from request body", http.StatusBadRequest)
		return
	}

	// Update device in db
	dc.devices.Update(hash, &d)

	// Respond with msgpack encoded device
	json.NewEncoder(w).Encode(d)
}

func (dc *deviceController) Delete(w http.ResponseWriter, r *http.Request) {
	// Get hash from url params
	hash := chi.URLParam(r, "hash")

	// Delete device from db
	dc.devices.Delete(hash)

	// Respond with status no content
	w.WriteHeader(http.StatusNoContent)
}
