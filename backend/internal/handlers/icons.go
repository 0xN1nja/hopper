package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type IconHandler struct {
	dataPath string
}

func NewIconHandler(dataPath string) *IconHandler {
	return &IconHandler{dataPath: dataPath}
}

func (h *IconHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "file too large (max 5 MB)")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".ico": true}
	if !allowed[ext] {
		writeError(w, http.StatusBadRequest, "unsupported file type")
		return
	}

	iconsDir := filepath.Join(h.dataPath, "icons")
	if err := os.MkdirAll(iconsDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}

	filename := uuid.New().String() + ext
	dst, err := os.Create(filepath.Join(iconsDir, filename))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"url": "/api/icons/" + filename})
}

func (h *IconHandler) Serve(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if strings.ContainsAny(filename, "/\\") || strings.Contains(filename, "..") {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(h.dataPath, "icons", filename))
}
