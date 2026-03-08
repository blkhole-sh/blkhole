package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/lemon3studio/blkhole/internal/model"
	"github.com/lemon3studio/blkhole/internal/repos"

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
	Register(email, password string) (*LoginResult, error)
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

	return as.generateTokens(user)
}

// Register creates a new user and returns tokens
func (as *authService) Register(email, password string) (*LoginResult, error) {
	// Check if user already exists
	existingUser, err := as.userRepo.FindByEmail(email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Hash password
	passwordHash, err := as.cryptoService.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		Name:         email, // Use email as name initially
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := as.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return as.generateTokens(user)
}

// generateTokens creates access and refresh tokens for a user
func (as *authService) generateTokens(user *model.User) (*LoginResult, error) {
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

	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token subject")
	}

	userID, err := strconv.Atoi(sub)
	if err != nil {
		return nil, fmt.Errorf("invalid token subject format")
	}

	user, err := as.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return as.generateTokens(user)
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

	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token subject")
	}

	userID, err := strconv.Atoi(sub)
	if err != nil {
		return nil, fmt.Errorf("invalid token subject format")
	}

	user, err := as.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
