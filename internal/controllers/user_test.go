package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/blkhole-sh/blkhole/internal/model"
	"github.com/blkhole-sh/blkhole/internal/services"
)

// MockAuthService
type mockUserAuthService struct {
	user *model.User
	err  error
}

func (m *mockUserAuthService) Login(email, password string) (*services.LoginResult, error) {
	return nil, nil
}
func (m *mockUserAuthService) RefreshToken(refreshToken string) (*services.LoginResult, error) {
	return nil, nil
}

func (m *mockUserAuthService) Register(email, password string) (*services.LoginResult, error) {
	return nil, nil
}
func (m *mockUserAuthService) UserFromContext(ctx context.Context) (*model.User, error) {
	return m.user, m.err
}

// MockUserRepo
type mockUserRepo struct {
	findByIdUser *model.User
	findByIdErr  error
}

func (m *mockUserRepo) Create(u *model.User) error                    { return nil }
func (m *mockUserRepo) Update(id int, u *model.User) error            { return nil }
func (m *mockUserRepo) Delete(id int) error                           { return nil }
func (m *mockUserRepo) LoadDeviceIDs(id int) ([]int, error)           { return nil, nil }
func (m *mockUserRepo) LoadListIDs(id int) ([]int, error)             { return nil, nil }
func (m *mockUserRepo) LoadScheduleIDs(id int) ([]int, error)         { return nil, nil }
func (m *mockUserRepo) LoadRelations(u *model.User) error             { return nil }
func (m *mockUserRepo) FindByEmail(email string) (*model.User, error) { return nil, nil }
func (m *mockUserRepo) FindAllIDs() ([]int, error)                    { return []int{}, nil }
func (m *mockUserRepo) FindByID(id int) (*model.User, error) {
	return m.findByIdUser, m.findByIdErr
}

// MockCryptoService
type mockCryptoService struct{}

func (m *mockCryptoService) HashPassword(password string) (string, error)       { return "", nil }
func (m *mockCryptoService) VerifyPassword(password, hash string) (bool, error) { return true, nil }
func (m *mockCryptoService) RandomHash() (string, error)                        { return "", nil }

func TestUserController_FindByID_IDOR(t *testing.T) {
	mockAuth := &mockUserAuthService{
		user: &model.User{ID: 1},
	}
	mockUserRepo := &mockUserRepo{
		findByIdUser: &model.User{ID: 2},
	}
	mockCrypto := &mockCryptoService{}

	uc := NewUserController(mockUserRepo, mockAuth, mockCrypto)

	r := httptest.NewRequest("GET", "/users/2", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "2")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	uc.FindByID(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status Forbidden (403), got %d", w.Code)
	}
}

func TestUserController_FindByID_Success(t *testing.T) {
	mockAuth := &mockUserAuthService{
		user: &model.User{ID: 1},
	}
	mockUserRepo := &mockUserRepo{
		findByIdUser: &model.User{ID: 1},
	}
	mockCrypto := &mockCryptoService{}

	uc := NewUserController(mockUserRepo, mockAuth, mockCrypto)

	r := httptest.NewRequest("GET", "/users/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	uc.FindByID(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK (200), got %d", w.Code)
	}
}

func TestUserController_Update_IDOR(t *testing.T) {
	mockAuth := &mockUserAuthService{
		user: &model.User{ID: 1},
	}
	mockUserRepo := &mockUserRepo{}
	mockCrypto := &mockCryptoService{}

	uc := NewUserController(mockUserRepo, mockAuth, mockCrypto)

	r := httptest.NewRequest("PATCH", "/users/2", strings.NewReader(`{"name":"New Name"}`))
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "2")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	uc.Update(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status Forbidden (403), got %d", w.Code)
	}
}

func TestUserController_Delete_IDOR(t *testing.T) {
	mockAuth := &mockUserAuthService{
		user: &model.User{ID: 1},
	}
	mockUserRepo := &mockUserRepo{}
	mockCrypto := &mockCryptoService{}

	uc := NewUserController(mockUserRepo, mockAuth, mockCrypto)

	r := httptest.NewRequest("DELETE", "/users/2", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "2")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	uc.Delete(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status Forbidden (403), got %d", w.Code)
	}
}

func TestUserController_Update_Success(t *testing.T) {
	mockAuth := &mockUserAuthService{
		user: &model.User{ID: 1},
	}
	mockUserRepo := &mockUserRepo{}
	mockCrypto := &mockCryptoService{}

	uc := NewUserController(mockUserRepo, mockAuth, mockCrypto)

	r := httptest.NewRequest("PATCH", "/users/1", strings.NewReader(`{"name":"New Name"}`))
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	uc.Update(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK (200), got %d", w.Code)
	}
}

func TestUserController_Delete_Success(t *testing.T) {
	mockAuth := &mockUserAuthService{
		user: &model.User{ID: 1},
	}
	mockUserRepo := &mockUserRepo{}
	mockCrypto := &mockCryptoService{}

	uc := NewUserController(mockUserRepo, mockAuth, mockCrypto)

	r := httptest.NewRequest("DELETE", "/users/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	uc.Delete(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status No Content (204), got %d", w.Code)
	}
}
