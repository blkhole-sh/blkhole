package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blkhole-sh/blkhole/internal/model"
	"github.com/blkhole-sh/blkhole/internal/repos"
	"github.com/blkhole-sh/blkhole/internal/services"
	"github.com/go-chi/chi/v5"
)

const (
	browserFailureLimit  = 10
	browserFailureWindow = time.Minute
)

// BrowserController defines browser pairing and rules API handlers.
type BrowserController interface {
	CreatePairing(http.ResponseWriter, *http.Request)
	Pair(http.ResponseWriter, *http.Request)
	Rules(http.ResponseWriter, *http.Request)
	ListClients(http.ResponseWriter, *http.Request)
	RevokeClient(http.ResponseWriter, *http.Request)
}

type failureWindow struct {
	count int
	until time.Time
}

type failureLimiter struct {
	mu       sync.Mutex
	failures map[string]failureWindow
}

func newFailureLimiter() *failureLimiter {
	return &failureLimiter{failures: make(map[string]failureWindow)}
}

func (l *failureLimiter) allowed(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	window := l.failures[key]
	if now.After(window.until) {
		delete(l.failures, key)
		return true
	}
	return window.count < browserFailureLimit
}

func (l *failureLimiter) failed(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	window := l.failures[key]
	if now.After(window.until) {
		window = failureWindow{until: now.Add(browserFailureWindow)}
	}
	window.count++
	l.failures[key] = window
}

func (l *failureLimiter) succeeded(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
}

type browserController struct {
	browsers services.BrowserService
	devices  repos.DeviceRepo
	auth     services.AuthService
	limiter  *failureLimiter
	now      func() time.Time
}

func NewBrowserController(browsers services.BrowserService, devices repos.DeviceRepo, auth services.AuthService) BrowserController {
	return &browserController{browsers: browsers, devices: devices, auth: auth, limiter: newFailureLimiter(), now: time.Now}
}

func (bc *browserController) requireDevice(w http.ResponseWriter, r *http.Request) (*model.Device, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return nil, false
	}
	user, ok := currentUser(w, r, bc.auth)
	if !ok {
		return nil, false
	}
	device, err := bc.devices.FindByID(id)
	if err != nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return nil, false
	}
	if device.UserID != user.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil, false
	}
	return device, true
}

func browserClientDTO(client *model.BrowserClient) model.BrowserClientDTO {
	return model.BrowserClientDTO{
		ID:           client.ID,
		Name:         client.Name,
		Browser:      client.Browser,
		CreatedAt:    time.Unix(client.CreatedAt, 0).UTC().Format(time.RFC3339),
		LastActiveAt: time.Unix(client.LastUsedAt, 0).UTC().Format(time.RFC3339),
	}
}

func (bc *browserController) CreatePairing(w http.ResponseWriter, r *http.Request) {
	device, ok := bc.requireDevice(w, r)
	if !ok {
		return
	}
	pairing, err := bc.browsers.CreatePairing(device.ID)
	if err != nil {
		http.Error(w, "Unable to create browser pairing", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"pairingToken": pairing.Token,
		"expiresAt":    pairing.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

func requestKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func (bc *browserController) Pair(w http.ResponseWriter, r *http.Request) {
	key := "pair:" + requestKey(r)
	now := bc.now()
	if !bc.limiter.allowed(key, now) {
		http.Error(w, "Too many failed pairing attempts", http.StatusTooManyRequests)
		return
	}

	var request struct {
		PairingToken string `json:"pairingToken"`
		ClientName   string `json:"clientName"`
		Browser      string `json:"browser"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := bc.browsers.Pair(request.PairingToken, request.ClientName, request.Browser)
	if errors.Is(err, repos.ErrInvalidBrowserPairing) {
		bc.limiter.failed(key, now)
		http.Error(w, "Invalid or expired pairing token", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "Unable to pair browser", http.StatusInternalServerError)
		return
	}
	bc.limiter.succeeded(key)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"accessToken": result.AccessToken,
		"client":      browserClientDTO(result.Client),
		"device": map[string]any{
			"id":   result.Device.ID,
			"name": result.Device.Name,
		},
	})
}

func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	value := r.Header.Get("Authorization")
	if !strings.HasPrefix(value, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, prefix))
}

func rulesETag(domains []string) string {
	hash := sha256.Sum256([]byte(strings.Join(domains, "\x00")))
	return `"` + hex.EncodeToString(hash[:]) + `"`
}

func (bc *browserController) Rules(w http.ResponseWriter, r *http.Request) {
	key := "auth:" + requestKey(r)
	now := bc.now()
	if !bc.limiter.allowed(key, now) {
		http.Error(w, "Too many failed authentication attempts", http.StatusTooManyRequests)
		return
	}
	_, device, err := bc.browsers.Authenticate(bearerToken(r))
	if errors.Is(err, services.ErrInvalidBrowserToken) {
		bc.limiter.failed(key, now)
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "Unable to authenticate browser", http.StatusInternalServerError)
		return
	}
	bc.limiter.succeeded(key)

	domains, err := bc.browsers.Rules(device.Hash)
	if err != nil {
		http.Error(w, "Unable to load browser rules", http.StatusInternalServerError)
		return
	}
	etag := rulesETag(domains)
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache")
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"domains": domains})
}

func (bc *browserController) ListClients(w http.ResponseWriter, r *http.Request) {
	device, ok := bc.requireDevice(w, r)
	if !ok {
		return
	}
	clients, err := bc.browsers.ListClients(device.ID)
	if err != nil {
		http.Error(w, "Unable to load browser clients", http.StatusInternalServerError)
		return
	}
	result := make([]model.BrowserClientDTO, len(clients))
	for i, client := range clients {
		result[i] = browserClientDTO(client)
	}
	json.NewEncoder(w).Encode(result)
}

func (bc *browserController) RevokeClient(w http.ResponseWriter, r *http.Request) {
	device, ok := bc.requireDevice(w, r)
	if !ok {
		return
	}
	clientID, err := strconv.Atoi(chi.URLParam(r, "clientId"))
	if err != nil {
		http.Error(w, "Invalid browser client ID", http.StatusBadRequest)
		return
	}
	if err := bc.browsers.RevokeClient(device.ID, clientID); errors.Is(err, repos.ErrBrowserClientNotFound) {
		http.Error(w, "Browser client not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Unable to revoke browser client", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
