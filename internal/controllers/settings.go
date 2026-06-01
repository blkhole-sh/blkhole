package controllers

import (
	"encoding/json"
	"net/http"
)

// SettingsController defines the interface for settings operations
type SettingsController interface {
	GetSettings(http.ResponseWriter, *http.Request)
}

// settingsController implements the SettingsController interface
type settingsController struct {
	upstreamDNS string
}

// NewSettingsController creates a new SettingsController instance
func NewSettingsController(upstreamDNS string) SettingsController {
	return &settingsController{upstreamDNS: upstreamDNS}
}

// GetSettings returns server configuration that is safe to expose publicly
func (sc *settingsController) GetSettings(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"upstreamDns": sc.upstreamDNS,
	})
}
