package services

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/blkhole-sh/blkhole/internal/cache"
	"github.com/blkhole-sh/blkhole/internal/model"
	"github.com/blkhole-sh/blkhole/internal/repos"
)

func newBrowserServiceTestData(t *testing.T) (*browserService, repos.BrowserRepo, *model.Device, *sql.DB) {
	t.Helper()
	database := setupTestDB(t)
	t.Cleanup(func() { database.Close() })
	users := repos.NewUserRepo(database)
	devices := repos.NewDeviceRepo(database)
	userID := createTestUser(t, users)
	device := &model.Device{Hash: "device-hash", Name: "MacBook", OS: model.MacOS, UserID: userID}
	if err := devices.Create(device); err != nil {
		t.Fatalf("create device: %v", err)
	}
	browsers := repos.NewBrowserRepo(database)
	blocker := NewContentBlocker(devices, repos.NewRuleRepo(database), repos.NewScheduleRepo(database), repos.NewDomainRepo(database), cache.NewDeviceCache())
	return &browserService{browsers: browsers, devices: devices, contentBlocker: blocker, now: time.Now}, browsers, device, database
}

func TestBrowserServicePairingLifecycleAndTokenHashing(t *testing.T) {
	service, _, device, database := newBrowserServiceTestData(t)
	now := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	pairing, err := service.CreatePairing(device.ID)
	if err != nil {
		t.Fatalf("CreatePairing: %v", err)
	}
	if pairing.ExpiresAt.Sub(now) != browserPairingTTL {
		t.Fatalf("pairing TTL = %s, want %s", pairing.ExpiresAt.Sub(now), browserPairingTTL)
	}

	result, err := service.Pair(pairing.Token, "Safari on MacBook", "Safari")
	if err != nil {
		t.Fatalf("Pair: %v", err)
	}
	if result.Device.ID != device.ID || result.AccessToken == "" {
		t.Fatalf("unexpected pair result: %#v", result)
	}
	if result.Client.TokenHash == result.AccessToken || result.Client.TokenHash != hashBrowserToken(result.AccessToken) {
		t.Fatal("long-lived browser token was not hashed before storage")
	}

	if _, err := service.Pair(pairing.Token, "Replay", "Safari"); !errors.Is(err, repos.ErrInvalidBrowserPairing) {
		t.Fatalf("pairing replay error = %v, want ErrInvalidBrowserPairing", err)
	}
	if _, _, err := service.Authenticate(result.AccessToken); err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if _, err := service.CreatePairing(device.ID); err != nil {
		t.Fatalf("create next pairing: %v", err)
	}
	var consumedPairings int
	if err := database.QueryRow("SELECT COUNT(*) FROM browser_pairing WHERE consumed_at IS NOT NULL").Scan(&consumedPairings); err != nil {
		t.Fatalf("count consumed pairings: %v", err)
	}
	if consumedPairings != 0 {
		t.Fatalf("consumed pairings = %d, want 0", consumedPairings)
	}
	if err := service.RevokeClient(device.ID, result.Client.ID); err != nil {
		t.Fatalf("RevokeClient: %v", err)
	}
	if _, _, err := service.Authenticate(result.AccessToken); !errors.Is(err, ErrInvalidBrowserToken) {
		t.Fatalf("Authenticate after revoke = %v, want ErrInvalidBrowserToken", err)
	}
}

func TestBrowserServicePairingExpires(t *testing.T) {
	service, _, device, _ := newBrowserServiceTestData(t)
	now := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	pairing, err := service.CreatePairing(device.ID)
	if err != nil {
		t.Fatalf("CreatePairing: %v", err)
	}
	service.now = func() time.Time { return now.Add(browserPairingTTL + time.Second) }
	if _, err := service.Pair(pairing.Token, "Expired", "Firefox"); !errors.Is(err, repos.ErrInvalidBrowserPairing) {
		t.Fatalf("expired pairing error = %v, want ErrInvalidBrowserPairing", err)
	}
}

func TestBrowserServiceDeviceDeletionInvalidatesToken(t *testing.T) {
	service, _, device, database := newBrowserServiceTestData(t)
	pairing, err := service.CreatePairing(device.ID)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.Pair(pairing.Token, "Firefox", "Firefox")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.devices.Delete(device.ID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM browser_client WHERE id = ?", result.Client.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("browser client remained after device deletion")
	}
	if _, _, err := service.Authenticate(result.AccessToken); !errors.Is(err, ErrInvalidBrowserToken) {
		t.Fatalf("Authenticate after device deletion = %v, want ErrInvalidBrowserToken", err)
	}
}
