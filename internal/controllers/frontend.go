package controllers

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

type FrontendController interface {
	Serve(w http.ResponseWriter, r *http.Request)
}

type frontendController struct {
	webFS fs.FS
}

func NewFrontendController(webFS fs.FS) FrontendController {
	return &frontendController{webFS: webFS}
}

func (fc *frontendController) Serve(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	content, err := fs.ReadFile(fc.webFS, path)
	if err != nil {
		content, err = fs.ReadFile(fc.webFS, "index.html")
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
