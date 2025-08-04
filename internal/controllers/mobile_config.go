package controllers

import (
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var mobileConfigTemplate = template.Must(template.ParseFiles("internal/templates/leo.mobileconfig.tmpl"))

// MobileConfigController defines the interface for mobile config operations
type MobileConfigController interface {
	GenerateConfig(http.ResponseWriter, *http.Request)
}

// MobileConfigControllerImpl implements the MobileConfigController interface
type MobileConfigControllerImpl struct{}

// NewMobileConfigController creates a new MobileConfigController instance
func NewMobileConfigController() MobileConfigController {
	return &MobileConfigControllerImpl{}
}

func (mc *MobileConfigControllerImpl) GenerateConfig(w http.ResponseWriter, r *http.Request) {
	// Get SERVER_hashes from url params
	userHash := chi.URLParam(r, "userHash")
	deviceHash := chi.URLParam(r, "deviceHash")

	// Generate UUIDs for mobile config
	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()

	// Load domain from env
	domain := os.Getenv("SERVER_PORT")

	// Build server URL
	serverURL := "https://" + domain + "/" + userHash + "/" + deviceHash + "/dns-query"

	// Prepare template data
	data := struct {
		UserHash  string
		UUID      string
		DNSUUID   string
		ServerURL string
	}{
		UserHash:  userHash,
		UUID:      uuid1,
		DNSUUID:   uuid2,
		ServerURL: serverURL,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/x-apple-aspen-config")
	w.Header().Set("Content-Disposition", `attachment; filename="leo.mobileconfig"`)

	// Execute template
	if err := mobileConfigTemplate.Execute(w, data); err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to generate mobile config", http.StatusInternalServerError)
	}
}
