package main

import (
	"net/http"
)

// Handler mengelola semua endpoint HTTP untuk PinjamKampus API.
type Handler struct {
	storage *Storage
}

// NewHandler membuat instance Handler baru dengan storage yang diberikan.
func NewHandler(storage *Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

// RegisterRoutes mendaftarkan routing endpoint ke HTTP ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.HealthCheck)
}

// HealthCheck handler untuk memeriksa status aplikasi.
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "PinjamKampus API is running",
	})
}
