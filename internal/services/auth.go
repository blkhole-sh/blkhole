package services

import (
	"context"
	"fmt"
	"time"

	"github.com/lemon3studio/leo/internal/model"
	"github.com/lemon3studio/leo/internal/repos"

	"github.com/go-chi/jwtauth/v5"
)

const (
	TokenExpiry        = 1 * time.Hour      // Short-lived access token
	RefreshTokenExpiry = 7 * 24 * time.Hour // 7 days refresh token
)

// LoginResult contains successful login response data
type LoginResult struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"-"` // Hidden from JSON, set as HttpOnly cookie
	RefreshToken string      `json:"-"` // Hidden from JSON, set as HttpOnly cookie
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	Login(email, password string) (*LoginResult, error)
	RefreshToken(refreshToken string) (*LoginResult, error)
	UserFromContext(ctx context.Context) (*model.User, error)
}

// authService implements the AuthService interface
type authService struct {
	userRepo      repos.UserRepo
	cryptoService CryptoService
	tokenAuth     *jwtauth.JWTAuth
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repos.UserRepo, cryptoService CryptoService, tokenAuth *jwtauth.JWTAuth) AuthService {
	return &authService{
		userRepo:      userRepo,
		cryptoService: cryptoService,
		tokenAuth:     tokenAuth,
	}
}

// Login authenticates user credentials and returns tokens
func (as *authService) Login(email, password string) (*LoginResult, error) {
	user, err := as.userRepo.FindByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	valid, err := as.cryptoService.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !valid {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Create access token (short-lived)
	accessClaims := map[string]any{
		"sub":   fmt.Sprintf("%d", user.ID),
		"email": user.Email,
		"type":  "access",
		"exp":   time.Now().Add(TokenExpiry).Unix(),
		"iat":   time.Now().Unix(),
	}

	_, accessToken, err := as.tokenAuth.Encode(accessClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// Create refresh token (long-lived)
	refreshClaims := map[string]any{
		"sub":  fmt.Sprintf("%d", user.ID),
		"type": "refresh",
		"exp":  time.Now().Add(RefreshTokenExpiry).Unix(),
		"iat":  time.Now().Unix(),
	}

	_, refreshToken, err := as.tokenAuth.Encode(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken creates new tokens using a valid refresh token
func (as *authService) RefreshToken(refreshToken string) (*LoginResult, error) {
	token, err := as.tokenAuth.Decode(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	claims := token.PrivateClaims()

	// Verify it's a refresh token
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	userID, ok := claims["sub"].(float64) // JSON numbers are float64 in Go
	if !ok {
		return nil, fmt.Errorf("invalid token subject")
	}

	user, err := as.userRepo.FindByID(int(userID))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Create new access token
	accessClaims := map[string]any{
		"sub":   fmt.Sprintf("%d", user.ID),
		"email": user.Email,
		"type":  "access",
		"exp":   time.Now().Add(TokenExpiry).Unix(),
		"iat":   time.Now().Unix(),
	}

	_, accessToken, err := as.tokenAuth.Encode(accessClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// Create new refresh token
	refreshClaims := map[string]any{
		"sub":  user.ID,
		"type": "refresh",
		"exp":  time.Now().Add(RefreshTokenExpiry).Unix(),
		"iat":  time.Now().Unix(),
	}

	_, newRefreshToken, err := as.tokenAuth.Encode(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// UserFromContext extracts user from JWT context
func (as *authService) UserFromContext(ctx context.Context) (*model.User, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("no token in context")
	}

	// Verify it's an access token
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return nil, fmt.Errorf("invalid token type")
	}

	userID, ok := claims["sub"].(float64) // JSON numbers are float64 in Go
	if !ok {
		return nil, fmt.Errorf("invalid token subject")
	}

	user, err := as.userRepo.FindByID(int(userID))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
