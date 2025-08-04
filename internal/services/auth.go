package services

import (
	"errors"
	"server/internal/model"
	"server/internal/repos"
)

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	UserHash  string `json"userHash"`
	Email     string `json:"email"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	SignIn(email, password string) (*model.User, error)
}

// AuthServiceImpl implements the AuthService interface
type AuthServiceImpl struct {
	userRepo      repos.UserRepo
	cryptoService CryptoService
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userRepo repos.UserRepo, cryptoService CryptoService) AuthService {
	return &AuthServiceImpl{
		userRepo:      userRepo,
		cryptoService: cryptoService,
	}
}

// SignIn authenticates a user with email and password
func (as *AuthServiceImpl) SignIn(email, password string) (*model.User, error) {
	// Find user by email
	user, err := as.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Verify the password
	isValid, err := as.cryptoService.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return nil, err
	}

	if !isValid {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}
