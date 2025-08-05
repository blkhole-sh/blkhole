package controllers

import (
	_ "embed"
	"log"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

//go:embed mobileconfig.tmpl
var mobileConfigTemplateContent string

var mobileConfigTemplate = template.Must(template.New("mobileconfig").Parse(mobileConfigTemplateContent))

// MobileConfigController defines the interface for mobile config operations
type MobileConfigController interface {
	GenerateConfig(http.ResponseWriter, *http.Request)
}

// mobileConfigController implements the MobileConfigController interface
type mobileConfigController struct {
	domain string
}

// NewMobileConfigController creates a new MobileConfigController instance
func NewMobileConfigController(domain string) MobileConfigController {
	return &mobileConfigController{domain: domain}
}

func (mc *mobileConfigController) GenerateConfig(w http.ResponseWriter, r *http.Request) {
	// Get SERVER_hashes from url params
	userHash := chi.URLParam(r, "userHash")
	deviceHash := chi.URLParam(r, "deviceHash")

	// Generate UUIDs for mobile config
	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()

	// Build server URL
	serverURL := "https://" + mc.domain + "/" + userHash + "/" + deviceHash + "/dns-query"

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
		log.Printf("failed to execute mobile config template: %v", err)
		http.Error(w, "Unable to generate mobile config", http.StatusInternalServerError)
	}
}
