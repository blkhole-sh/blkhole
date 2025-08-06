package controllers

import (
	"crypto/rand"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
)

//go:embed mobileconfig.tmpl
var mobileConfigTemplateContent string

// generateUUID creates a UUID v4 using standard library
func generateUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	
	// Set version (4) and variant bits
	bytes[6] = (bytes[6] & 0x0f) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3f) | 0x80 // Variant bits
	
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

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
	// Get device hash from url params
	deviceHash := chi.URLParam(r, "hash")

	// Generate UUIDs for mobile config using standard library
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	// Build server URL
	serverURL := "https://" + mc.domain + "/" + deviceHash + "/dns-query"

	// Prepare template data
	data := struct {
		DeviceHash string
		UUID       string
		DNSUUID    string
		ServerURL  string
	}{
		DeviceHash: deviceHash,
		UUID:       uuid1,
		DNSUUID:    uuid2,
		ServerURL:  serverURL,
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
