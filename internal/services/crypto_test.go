package services

import (
	"strings"
	"testing"
)

func TestCryptoService_RandomHash_Format(t *testing.T) {
	cs := NewCryptoService([]byte("test-secret"))

	hash, err := cs.RandomHash()
	if err != nil {
		t.Fatalf("RandomHash() unexpected error: %v", err)
	}
	if hash == "" {
		t.Error("RandomHash() returned empty string")
	}
	if hash != strings.ToLower(hash) {
		t.Errorf("RandomHash() result is not lowercase: %q", hash)
	}
}

func TestCryptoService_RandomHash_Uniqueness(t *testing.T) {
	cs := NewCryptoService([]byte("test-secret"))

	h1, err := cs.RandomHash()
	if err != nil {
		t.Fatalf("first RandomHash() error: %v", err)
	}
	h2, err := cs.RandomHash()
	if err != nil {
		t.Fatalf("second RandomHash() error: %v", err)
	}
	if h1 == h2 {
		t.Error("RandomHash() produced identical hashes on consecutive calls")
	}
}

func TestCryptoService_RandomHash_LargeSecret(t *testing.T) {
	// blake2b accepts keys 0-64 bytes; >64 bytes is invalid
	cs := NewCryptoService(make([]byte, 65))
	_, err := cs.RandomHash()
	if err == nil {
		t.Error("RandomHash() with 65-byte secret expected error, got nil")
	}
}

func TestCryptoService_HashPassword_Format(t *testing.T) {
	cs := NewCryptoService([]byte("test-secret"))

	h, err := cs.HashPassword("mypassword")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	parts := strings.Split(h, "$")
	if len(parts) != 2 {
		t.Errorf("HashPassword() expected 'salt$hash' format, got %q", h)
	}
	for _, p := range parts {
		if p == "" {
			t.Errorf("HashPassword() has empty segment in %q", h)
		}
	}
}

func TestCryptoService_HashPassword_Uniqueness(t *testing.T) {
	cs := NewCryptoService([]byte("test-secret"))

	h1, _ := cs.HashPassword("password")
	h2, _ := cs.HashPassword("password")
	if h1 == h2 {
		t.Error("HashPassword() produced identical hashes for the same password (salt not random)")
	}
}

func TestCryptoService_VerifyPassword_Correct(t *testing.T) {
	cs := NewCryptoService([]byte("test-secret"))

	password := "correct-horse-battery"
	hashed, err := cs.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	ok, err := cs.VerifyPassword(password, hashed)
	if err != nil {
		t.Fatalf("VerifyPassword() error: %v", err)
	}
	if !ok {
		t.Error("VerifyPassword() returned false for correct password")
	}
}

func TestCryptoService_VerifyPassword_Wrong(t *testing.T) {
	cs := NewCryptoService([]byte("test-secret"))

	hashed, _ := cs.HashPassword("correct")
	ok, err := cs.VerifyPassword("wrong", hashed)
	if err != nil {
		t.Fatalf("VerifyPassword() unexpected error: %v", err)
	}
	if ok {
		t.Error("VerifyPassword() returned true for wrong password")
	}
}

func TestCryptoService_VerifyPassword_InvalidFormat(t *testing.T) {
	cs := NewCryptoService([]byte("test-secret"))

	_, err := cs.VerifyPassword("password", "noDollarSign")
	if err == nil {
		t.Error("VerifyPassword() expected error for invalid hash format")
	}
}

func TestCryptoService_VerifyPassword_RoundTrip(t *testing.T) {
	cs := NewCryptoService([]byte("secret"))

	passwords := []string{"", "unicode-✓", strings.Repeat("a", 200)}
	for _, pw := range passwords {
		t.Run(pw, func(t *testing.T) {
			h, err := cs.HashPassword(pw)
			if err != nil {
				t.Fatalf("HashPassword(%q) error: %v", pw, err)
			}
			ok, err := cs.VerifyPassword(pw, h)
			if err != nil {
				t.Fatalf("VerifyPassword(%q) error: %v", pw, err)
			}
			if !ok {
				t.Errorf("VerifyPassword(%q) round-trip failed", pw)
			}
		})
	}
}
