package controllers

import (
	"net/http"

	"github.com/lemon3studio/blkhole/internal/model"
	"github.com/lemon3studio/blkhole/internal/services"
)

// currentUser extracts the authenticated user from the request context.
// On failure it writes a 401 response and returns false.
func currentUser(w http.ResponseWriter, r *http.Request, authService services.AuthService) (*model.User, bool) {
	user, err := authService.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	return user, true
}
