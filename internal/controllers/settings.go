package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/blkhole-sh/blkhole/internal/repos"
	"github.com/blkhole-sh/blkhole/internal/services"
)

// SettingsController defines the interface for settings operations
type SettingsController interface {
	GetSettings(http.ResponseWriter, *http.Request)
	UpdateSettings(http.ResponseWriter, *http.Request)
}

// settingsController implements the SettingsController interface
type settingsController struct {
	settings    repos.SettingsRepo
	resolver    services.MutableResolver
	authService services.AuthService
}

// NewSettingsController creates a new SettingsController instance
func NewSettingsController(settings repos.SettingsRepo, resolver services.MutableResolver, authService services.AuthService) SettingsController {
	return &settingsController{settings: settings, resolver: resolver, authService: authService}
}

// GetSettings returns server configuration that is safe to expose publicly
func (sc *settingsController) GetSettings(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"upstreamDns": sc.resolver.UpstreamDNS(),
	})
}

// UpdateSettings persists and applies editable server settings.
func (sc *settingsController) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if _, ok := currentUser(w, r, sc.authService); !ok {
		return
	}

	var request struct {
		UpstreamDNS string `json:"upstreamDns"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Unable to decode settings", http.StatusBadRequest)
		return
	}

	upstreamDNS, err := services.ValidateUpstreamDNS(strings.TrimSpace(request.UpstreamDNS))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := sc.settings.UpdateUpstreamDNS(upstreamDNS); err != nil {
		log.Printf("failed to persist upstream DNS: %v", err)
		http.Error(w, "Unable to update settings", http.StatusInternalServerError)
		return
	}

	sc.resolver.SetUpstreamDNS(upstreamDNS)
	json.NewEncoder(w).Encode(map[string]string{"upstreamDns": upstreamDNS})
}
