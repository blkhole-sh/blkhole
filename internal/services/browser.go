package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/blkhole-sh/blkhole/internal/model"
	"github.com/blkhole-sh/blkhole/internal/repos"
)

const browserPairingTTL = 90 * time.Second

var ErrInvalidBrowserToken = errors.New("invalid browser token")

type BrowserPairing struct {
	Token     string
	ExpiresAt time.Time
}

type BrowserPairResult struct {
	AccessToken string
	Client      *model.BrowserClient
	Device      *model.Device
}

// BrowserService handles browser pairing and device-scoped rule access.
type BrowserService interface {
	CreatePairing(deviceID int) (*BrowserPairing, error)
	Pair(pairingToken, clientName, browser string) (*BrowserPairResult, error)
	Authenticate(accessToken string) (*model.BrowserClient, *model.Device, error)
	ListClients(deviceID int) ([]*model.BrowserClient, error)
	RevokeClient(deviceID, clientID int) error
	Rules(deviceHash string) ([]string, error)
}

type browserService struct {
	browsers       repos.BrowserRepo
	devices        repos.DeviceRepo
	contentBlocker ContentBlocker
	now            func() time.Time
}

func NewBrowserService(browsers repos.BrowserRepo, devices repos.DeviceRepo, contentBlocker ContentBlocker) BrowserService {
	return &browserService{browsers: browsers, devices: devices, contentBlocker: contentBlocker, now: time.Now}
}

func randomBrowserToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func hashBrowserToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (bs *browserService) CreatePairing(deviceID int) (*BrowserPairing, error) {
	token, err := randomBrowserToken()
	if err != nil {
		return nil, err
	}
	now := bs.now().UTC()
	expiresAt := now.Add(browserPairingTTL)
	if err := bs.browsers.CreatePairing(deviceID, hashBrowserToken(token), now, expiresAt); err != nil {
		return nil, err
	}
	return &BrowserPairing{Token: token, ExpiresAt: expiresAt}, nil
}

func (bs *browserService) Pair(pairingToken, clientName, browser string) (*BrowserPairResult, error) {
	clientName = strings.TrimSpace(clientName)
	browser = strings.TrimSpace(browser)
	if pairingToken == "" || clientName == "" || len(clientName) > 100 || len(browser) > 50 {
		return nil, repos.ErrInvalidBrowserPairing
	}

	accessToken, err := randomBrowserToken()
	if err != nil {
		return nil, err
	}
	client, err := bs.browsers.ConsumePairing(hashBrowserToken(pairingToken), clientName, browser, hashBrowserToken(accessToken), bs.now().UTC())
	if err != nil {
		return nil, err
	}
	device, err := bs.devices.FindByID(client.DeviceID)
	if err != nil {
		return nil, err
	}
	return &BrowserPairResult{AccessToken: accessToken, Client: client, Device: device}, nil
}

func (bs *browserService) Authenticate(accessToken string) (*model.BrowserClient, *model.Device, error) {
	if accessToken == "" {
		return nil, nil, ErrInvalidBrowserToken
	}
	client, err := bs.browsers.FindActiveClientByTokenHash(hashBrowserToken(accessToken), bs.now().UTC())
	if errors.Is(err, repos.ErrBrowserClientNotFound) {
		return nil, nil, ErrInvalidBrowserToken
	}
	if err != nil {
		return nil, nil, err
	}
	device, err := bs.devices.FindByID(client.DeviceID)
	if err != nil {
		return nil, nil, ErrInvalidBrowserToken
	}
	return client, device, nil
}

func (bs *browserService) ListClients(deviceID int) ([]*model.BrowserClient, error) {
	return bs.browsers.FindClientsByDevice(deviceID)
}

func (bs *browserService) RevokeClient(deviceID, clientID int) error {
	return bs.browsers.RevokeClient(deviceID, clientID, bs.now().UTC())
}

func (bs *browserService) Rules(deviceHash string) ([]string, error) {
	return bs.contentBlocker.EffectiveBlockedDomains(deviceHash)
}
