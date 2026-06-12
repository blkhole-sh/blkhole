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

// revisionPlaceholder is the token baked into index.html at build time and
// replaced with the running binary's revision when served.
const revisionPlaceholder = "__BLKHOLE_REVISION__"

// webController implements the WebController interface
type webController struct {
	webFS    fs.FS
	revision string
}

// NewWebController creates a new WebController instance
func NewWebController(webFS fs.FS, revision string) WebController {
	return &webController{webFS: webFS, revision: revision}
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
		path = "index.html"
	}

	switch filepath.Ext(path) {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	if path == "index.html" {
		content = []byte(strings.Replace(string(content), revisionPlaceholder, wc.revision, 1))
	}

	w.Write(content)
}
