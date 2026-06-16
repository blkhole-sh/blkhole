package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestWebController_Serve_ExistingFile(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>index</html>")},
		"app.js":     &fstest.MapFile{Data: []byte("console.log('hi')")},
		"style.css":  &fstest.MapFile{Data: []byte("body{}")},
	}
	controller := NewWebController(mockFS, "")

	tests := []struct {
		path           string
		expectedStatus int
		expectedBody   string
		expectedCT     string
	}{
		{"/app.js", http.StatusOK, "console.log('hi')", "application/javascript"},
		{"/style.css", http.StatusOK, "body{}", "text/css"},
		{"/index.html", http.StatusOK, "<html>index</html>", "text/html; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()
			controller.Serve(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d", tt.expectedStatus, rr.Code)
			}
			if body := rr.Body.String(); body != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, body)
			}
			if ct := rr.Header().Get("Content-Type"); ct != tt.expectedCT {
				t.Errorf("expected Content-Type %q, got %q", tt.expectedCT, ct)
			}
		})
	}
}

func TestWebController_Serve_RootFallsBackToIndex(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>root</html>")},
	}
	controller := NewWebController(mockFS, "")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	controller.Serve(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); body != "<html>root</html>" {
		t.Errorf("expected index.html content, got %q", body)
	}
}

func TestWebController_Serve_UnknownPathFallsBackToIndex(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>spa</html>")},
	}
	controller := NewWebController(mockFS, "")

	req := httptest.NewRequest(http.MethodGet, "/some/spa/route", nil)
	rr := httptest.NewRecorder()
	controller.Serve(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for SPA fallback, got %d", rr.Code)
	}
	if body := rr.Body.String(); body != "<html>spa</html>" {
		t.Errorf("expected index.html content, got %q", body)
	}
}

func TestWebController_Serve_InjectsRevision(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`<meta content="__BLKHOLE_REVISION__" />`)},
	}
	controller := NewWebController(mockFS, "abc1234")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	controller.Serve(rr, req)

	if body := rr.Body.String(); body != `<meta content="abc1234" />` {
		t.Errorf("expected revision injected, got %q", body)
	}
}

func TestWebController_Serve_NoIndexReturns404(t *testing.T) {
	mockFS := fstest.MapFS{}
	controller := NewWebController(mockFS, "")

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()
	controller.Serve(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
