package services

import (
	"context"
	"fmt"
	"server/internal/model"
	"server/internal/repos"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

const TokenExpiry = 24 * time.Hour

// LoginResult contains successful login response data
type LoginResult struct {
	User  *model.User `json:"user"`
	Token string      `json:"token"`
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	Login(email, password string) (*LoginResult, error)
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

// Login authenticates user credentials and returns a token
func (as *authService) Login(email, password string) (*LoginResult, error) {
	user, err := as.userRepo.FindByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	valid, err := as.cryptoService.VerifyPassword(password, user.PasswordHash)
	if err != nil || !valid {
		return nil, fmt.Errorf("invalid credentials")
	}

	claims := map[string]any{
		"sub":   user.Hash,
		"email": user.Email,
		"exp":   time.Now().Add(TokenExpiry).Unix(),
	}

	_, token, err := as.tokenAuth.Encode(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	return &LoginResult{User: user, Token: token}, nil
}

// UserFromContext extracts user from JWT context
func (as *authService) UserFromContext(ctx context.Context) (*model.User, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("no token in context")
	}

	userHash, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token subject")
	}

	user, err := as.userRepo.FindByHash(userHash)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
