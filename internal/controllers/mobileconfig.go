package controllers

import (
	"crypto/rand"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/lemon3studio/leo/internal/repos"
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
	domain  string
	devices repos.DeviceRepo
}

// NewMobileConfigController creates a new MobileConfigController instance
func NewMobileConfigController(domain string, devices repos.DeviceRepo) MobileConfigController {
	return &mobileConfigController{
		domain:  domain,
		devices: devices,
	}
}

func (mc *mobileConfigController) GenerateConfig(w http.ResponseWriter, r *http.Request) {
	// Get device id from url params
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("unable to parse id from path parameter: %v", err)
		http.Error(w, "Unable to parse id from path parameter", http.StatusBadRequest)
		return
	}

	// Find device with given id in db
	device, err := mc.devices.FindByID(id)
	if err != nil {
		log.Printf("failed to find device by id %d: %v", id, err)
		http.Error(w, "Unable to find device in db", http.StatusNotFound)
		return
	}

	// Generate UUIDs for mobile config
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	// Build server URL using device hash (required for DNS query endpoint)
	serverURL := "https://" + mc.domain + "/" + device.Hash + "/dns-query"

	// Prepare template data
	data := struct {
		DeviceHash string
		DeviceName string
		UUID       string
		DNSUUID    string
		ServerURL  string
	}{
		DeviceHash: device.Hash,
		DeviceName: device.Name,
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
