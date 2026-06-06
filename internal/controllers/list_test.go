package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lemon3studio/blkhole/internal/model"
)

// MockListService satisfies services.ListService
type MockListService struct{}

func (m *MockListService) LoadList(l *model.List) error  { return nil }
func (m *MockListService) SeedDefaults(userID int) error { return nil }

// MockListRepo satisfies repos.ListRepo
type MockListRepo struct {
	CreateFunc    func(l *model.List) (int, error)
	UpdateFunc    func(id int, l *model.List) error
	DeleteFunc    func(id int) error
	FindByIDFunc  func(id int) (*model.List, error)
	FindByUserFunc func(userID int) ([]*model.List, error)
}

func (m *MockListRepo) Create(l *model.List) (int, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(l)
	}
	l.ID = 1
	return 1, nil
}
func (m *MockListRepo) Update(id int, l *model.List) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(id, l)
	}
	return nil
}
func (m *MockListRepo) Delete(id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}
func (m *MockListRepo) LinkSchedule(id int, scheduleID int) error     { return nil }
func (m *MockListRepo) LoadRules(id int) ([]model.Rule, error)        { return []model.Rule{}, nil }
func (m *MockListRepo) GetRuleCount(id int) (int, error)              { return 0, nil }
func (m *MockListRepo) LoadScheduleIDs(id int) ([]int, error)         { return []int{}, nil }
func (m *MockListRepo) LoadRelations(l *model.List) error             { return nil }
func (m *MockListRepo) FindByID(id int) (*model.List, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(id)
	}
	return nil, fmt.Errorf("not found")
}
func (m *MockListRepo) FindAll() ([]*model.List, error) { return []*model.List{}, nil }
func (m *MockListRepo) FindByUser(userID int) ([]*model.List, error) {
	if m.FindByUserFunc != nil {
		return m.FindByUserFunc(userID)
	}
	return []*model.List{}, nil
}
func (m *MockListRepo) FindNamesByScheduleID(scheduleID int) ([]string, error) {
	return []string{}, nil
}
func (m *MockListRepo) FindBySchedule(scheduleID int) ([]*model.List, error) {
	return []*model.List{}, nil
}

func (m *MockListRepo) HasDefaultsForUser(userID int) (bool, error) { return false, nil }

func TestListController_Create_Success(t *testing.T) {
	listRepo := &MockListRepo{
		CreateFunc: func(l *model.List) (int, error) {
			l.ID = 10
			return 10, nil
		},
	}
	controller := NewListController(listRepo, &MockListService{})

	body, _ := json.Marshal(map[string]interface{}{"name": "My List", "userId": 1})
	req := httptest.NewRequest(http.MethodPost, "/lists", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	controller.Create(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestListController_Create_InvalidJSON(t *testing.T) {
	controller := NewListController(&MockListRepo{}, &MockListService{})

	req := httptest.NewRequest(http.MethodPost, "/lists", bytes.NewBufferString("bad"))
	rr := httptest.NewRecorder()
	controller.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestListController_FindByID_Success(t *testing.T) {
	listRepo := &MockListRepo{
		FindByIDFunc: func(id int) (*model.List, error) {
			return &model.List{ID: id, Name: "Found List"}, nil
		},
	}
	controller := NewListController(listRepo, &MockListService{})

	req := withParam(httptest.NewRequest(http.MethodGet, "/lists/1", nil), "id", "1")
	rr := httptest.NewRecorder()
	controller.FindByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestListController_FindByID_BadParam(t *testing.T) {
	controller := NewListController(&MockListRepo{}, &MockListService{})

	req := withParam(httptest.NewRequest(http.MethodGet, "/lists/abc", nil), "id", "abc")
	rr := httptest.NewRecorder()
	controller.FindByID(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestListController_FindByID_NotFound(t *testing.T) {
	listRepo := &MockListRepo{
		FindByIDFunc: func(id int) (*model.List, error) { return nil, fmt.Errorf("not found") },
	}
	controller := NewListController(listRepo, &MockListService{})

	req := withParam(httptest.NewRequest(http.MethodGet, "/lists/99", nil), "id", "99")
	rr := httptest.NewRecorder()
	controller.FindByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestListController_FindByUser_Success(t *testing.T) {
	listRepo := &MockListRepo{
		FindByUserFunc: func(userID int) ([]*model.List, error) {
			return []*model.List{{ID: 1, Name: "L1"}, {ID: 2, Name: "L2"}}, nil
		},
	}
	controller := NewListController(listRepo, &MockListService{})

	req := withParam(httptest.NewRequest(http.MethodGet, "/users/1/lists", nil), "userId", "1")
	rr := httptest.NewRecorder()
	controller.FindByUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []model.ListDTO
	json.NewDecoder(rr.Body).Decode(&result)
	if len(result) != 2 {
		t.Errorf("expected 2 lists, got %d", len(result))
	}
}

func TestListController_Delete_Success(t *testing.T) {
	repo := &MockListRepo{
		FindByIDFunc: func(id int) (*model.List, error) {
			return &model.List{ID: id, IsDefault: false}, nil
		},
	}
	controller := NewListController(repo, &MockListService{})

	req := withParam(httptest.NewRequest(http.MethodDelete, "/lists/1", nil), "id", "1")
	rr := httptest.NewRecorder()
	controller.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}
