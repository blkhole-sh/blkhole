package services

import (
	"testing"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lemon3studio/blkhole/internal/repos"
)

func TestAuthService_Register(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer db.Close()

	userRepo := repos.NewUserRepo(db)
	scheduleRepo := repos.NewScheduleRepo(db)
	cryptoService := NewCryptoService([]byte("secret"))
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	authService := NewAuthService(userRepo, scheduleRepo, cryptoService, tokenAuth)

	// Test Case 1: Successful Registration
	email := "test@example.com"
	password := "password123"

	result, err := authService.Register(email, password)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if result.User.Email != email {
		t.Errorf("Expected email %s, got %s", email, result.User.Email)
	}
	if result.AccessToken == "" {
		t.Error("Access token is empty")
	}
	if result.RefreshToken == "" {
		t.Error("Refresh token is empty")
	}

	// Verify user in DB
	user, err := userRepo.FindByEmail(email)
	if err != nil {
		t.Fatalf("User not found in DB: %v", err)
	}
	if user.Email != email {
		t.Errorf("Expected DB email %s, got %s", email, user.Email)
	}

	// Test Case 2: Duplicate Email
	_, err = authService.Register(email, "newpassword")
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}
	if err.Error() != "email already registered" {
		t.Errorf("Expected error 'email already registered', got '%v'", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer db.Close()

	userRepo := repos.NewUserRepo(db)
	scheduleRepo := repos.NewScheduleRepo(db)
	cryptoService := NewCryptoService([]byte("secret"))
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	authService := NewAuthService(userRepo, scheduleRepo, cryptoService, tokenAuth)

	// Create user
	email := "login@example.com"
	password := "password123"
	_, err := authService.Register(email, password)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Test Case 1: Successful Login
	result, err := authService.Login(email, password)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if result.User.Email != email {
		t.Errorf("Expected email %s, got %s", email, result.User.Email)
	}

	// Test Case 2: Wrong Password
	_, err = authService.Login(email, "wrongpassword")
	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}

	// Test Case 3: Non-existent User
	_, err = authService.Login("nonexistent@example.com", password)
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}
