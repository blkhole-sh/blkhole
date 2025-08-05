package controllers

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

// WebController defines the interface for web frontend operations
type WebController interface {
	Serve(w http.ResponseWriter, r *http.Request)
}

// webController implements the WebController interface
type webController struct {
	webFS fs.FS
}

// NewWebController creates a new WebController instance
func NewWebController(webFS fs.FS) WebController {
	return &webController{webFS: webFS}
}

func (wc *webController) Serve(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	content, err := fs.ReadFile(wc.webFS, path)
	if err != nil {
		content, err = fs.ReadFile(wc.webFS, "index.html")
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
		return
	}

	switch filepath.Ext(path) {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	w.Write(content)
}
