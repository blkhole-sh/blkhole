package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"server/model"
	"server/repos"
	"server/services"

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

// DeviceControllerImpl implements the DeviceController interface
type DeviceControllerImpl struct {
	deviceRepo    repos.DeviceRepo
	cryptoService services.CryptoService
}

// NewDeviceController creates a new DeviceController instance
func NewDeviceController(deviceRepo repos.DeviceRepo, cryptoService services.CryptoService) DeviceController {
	return &DeviceControllerImpl{
		deviceRepo:    deviceRepo,
		cryptoService: cryptoService,
	}
}

func (dc *DeviceControllerImpl) Create(w http.ResponseWriter, r *http.Request) {
	// Initialize device
	var d model.Device

	// Encode device from request body
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to decode device from request body", http.StatusBadRequest)
	}

	hash, err := dc.cryptoService.RandomHash()
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to create hash for user", http.StatusInternalServerError)
	}

	d.Hash = hash

	// Store device into db
	dc.deviceRepo.Create(&d)

	// Respond with json encoded device
	json.NewEncoder(w).Encode(d)
}

func (dc *DeviceControllerImpl) FindByHash(w http.ResponseWriter, r *http.Request) {
	// Get hash from url params
	hash := chi.URLParam(r, "hash")

	d, err := dc.deviceRepo.FindByHash(hash)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to find device in db", http.StatusNotFound)
	}

	json.NewEncoder(w).Encode(d)
}

func (dc *DeviceControllerImpl) FindByUser(w http.ResponseWriter, r *http.Request) {
	// Get user hash from url params
	userHash := chi.URLParam(r, "userHash")

	// Find devices in db
	d, err := dc.deviceRepo.FindByUser(userHash)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to find devices in db", http.StatusNotFound)
	}

	// Respond with json encoded devices
	json.NewEncoder(w).Encode(d)
}

func (dc *DeviceControllerImpl) Update(w http.ResponseWriter, r *http.Request) {
	// Initialize device
	var d model.Device

	// Get hash from url params
	hash := chi.URLParam(r, "hash")

	// Encode device from request body
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to decode device from request body", http.StatusBadRequest)
	}

	// Update device in db
	dc.deviceRepo.Update(hash, &d)

	// Respond with json encoded device
	json.NewEncoder(w).Encode(d)
}

func (dc *DeviceControllerImpl) Delete(w http.ResponseWriter, r *http.Request) {
	// Get hash from url params
	hash := chi.URLParam(r, "hash")

	// Delete device from db
	dc.deviceRepo.Delete(hash)

	// Respond with status no content
	w.WriteHeader(http.StatusNoContent)
}
