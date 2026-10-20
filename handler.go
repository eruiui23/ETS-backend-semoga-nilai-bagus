package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// Handler mengelola semua endpoint HTTP untuk PinjamKampus API.
type Handler struct {
	storage *Storage
}

// NewHandler membuat instance Handler baru dengan storage default dan mengembalikan http.Handler.
func NewHandler() http.Handler {
	storage := NewStorage()
	return NewHandlerWithStorage(storage)
}

// NewHandlerWithStorage membuat instance http.Handler dengan storage tertentu.
func NewHandlerWithStorage(storage *Storage) http.Handler {
	h := &Handler{
		storage: storage,
	}

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Middleware untuk menormalkan trailing slash agar routing konsisten
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > 1 && strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		mux.ServeHTTP(w, r)
	})
}

// RegisterRoutes mendaftarkan routing endpoint ke HTTP ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /sumber-daya", h.CreateSumberDaya)
	mux.HandleFunc("GET /sumber-daya", h.ListSumberDaya)
	mux.HandleFunc("GET /sumber-daya/{id}", h.GetSumberDayaByID)
}

// CreateSumberDaya menangani pembuatan sumber daya baru (B1).
func (h *Handler) CreateSumberDaya(w http.ResponseWriter, r *http.Request) {
	var req SumberDaya
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	req.Nama = strings.TrimSpace(req.Nama)
	if req.Nama == "" || len(req.Nama) > 50 {
		writeError(w, http.StatusBadRequest, "nama wajib diisi dan maksimal 50 karakter")
		return
	}

	if req.Jenis != "ruang" && req.Jenis != "alat" {
		writeError(w, http.StatusBadRequest, "jenis harus 'ruang' atau 'alat'")
		return
	}

	if req.Stok < 1 {
		writeError(w, http.StatusBadRequest, "stok minimal 1")
		return
	}

	created := h.storage.CreateSumberDaya(req)
	writeJSON(w, http.StatusCreated, created)
}

// ListSumberDaya mengembalikan seluruh daftar sumber daya (B1).
func (h *Handler) ListSumberDaya(w http.ResponseWriter, r *http.Request) {
	list := h.storage.GetAllSumberDaya()
	writeJSON(w, http.StatusOK, list)
}

// GetSumberDayaByID mengembalikan detail satu sumber daya berdasarkan ID (B1).
// Mematuhi urutan validasi ketat:
// 1. ID bukan angka -> 400 Bad Request
// 2. Data tidak ditemukan -> 404 Not Found
func (h *Handler) GetSumberDayaByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id harus berupa angka")
		return
	}

	sd, ok := h.storage.GetSumberDaya(id)
	if !ok {
		writeError(w, http.StatusNotFound, "sumber daya tidak ditemukan")
		return
	}

	writeJSON(w, http.StatusOK, sd)
}
