package web

import (
	"embed"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

// Dist contains the Vite production build. The all: prefix keeps the placeholder
// file embeddable before the first frontend build.
//
//go:embed all:dist
var dist embed.FS

func Handler() (http.Handler, error) {
	root, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil, fmt.Errorf("open embedded frontend: %w", err)
	}
	return &spaHandler{root: root}, nil
}

type spaHandler struct{ root fs.FS }

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name == "." || name == "" {
		name = "index.html"
	}

	if info, err := fs.Stat(h.root, name); err == nil && !info.IsDir() {
		h.serveFile(w, r, name)
		return
	}
	if _, err := fs.Stat(h.root, "index.html"); err == nil {
		h.serveFile(w, r, "index.html")
		return
	}

	http.Error(w, "frontend build missing; run `npm run build` in web/", http.StatusServiceUnavailable)
}

func (h *spaHandler) serveFile(w http.ResponseWriter, r *http.Request, name string) {
	body, err := fs.ReadFile(h.root, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if contentType := mime.TypeByExtension(path.Ext(name)); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	if strings.HasPrefix(name, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	_, _ = w.Write(body)
}
