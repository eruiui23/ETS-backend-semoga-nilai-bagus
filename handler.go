package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
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

	mux.HandleFunc("POST /peminjaman", h.CreatePeminjaman)
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

// CreatePeminjaman menangani pembuatan peminjaman baru (B2).
func (h *Handler) CreatePeminjaman(w http.ResponseWriter, r *http.Request) {
	var req Peminjaman
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	// 1. Validasi keberadaan sumber_daya_id
	if req.SumberDayaID <= 0 {
		writeError(w, http.StatusBadRequest, "sumber_daya_id harus berupa angka positif")
		return
	}
	sd, exists := h.storage.GetSumberDaya(req.SumberDayaID)
	if !exists {
		writeError(w, http.StatusNotFound, "sumber daya tidak ditemukan")
		return
	}

	// 2. Validasi peminjam
	req.Peminjam = strings.TrimSpace(req.Peminjam)
	if req.Peminjam == "" || len(req.Peminjam) > 50 {
		writeError(w, http.StatusBadRequest, "peminjam wajib diisi dan maksimal 50 karakter")
		return
	}

	// 3. Validasi tanggal format YYYY-MM-DD
	if _, err := time.Parse("2006-01-02", req.Tanggal); err != nil {
		writeError(w, http.StatusBadRequest, "tanggal harus dalam format YYYY-MM-DD yang valid")
		return
	}

	// 4. Validasi slot_mulai dan slot_selesai (1-8, selesai >= mulai)
	if req.SlotMulai < 1 || req.SlotMulai > 8 {
		writeError(w, http.StatusBadRequest, "slot_mulai harus di antara 1 sampai 8")
		return
	}
	if req.SlotSelesai < 1 || req.SlotSelesai > 8 {
		writeError(w, http.StatusBadRequest, "slot_selesai harus di antara 1 sampai 8")
		return
	}
	if req.SlotSelesai < req.SlotMulai {
		writeError(w, http.StatusBadRequest, "slot_selesai harus lebih besar atau sama dengan slot_mulai")
		return
	}

	// 5. Validasi jumlah (1 sampai stok sumber daya)
	if req.Jumlah < 1 || req.Jumlah > sd.Stok {
		writeError(w, http.StatusBadRequest, "jumlah harus antara 1 sampai stok sumber daya")
		return
	}

	// 6. Default status: "aktif"
	if strings.TrimSpace(req.Status) == "" {
		req.Status = "aktif"
	}

	// 7. Simpan peminjaman & cek deteksi bentrok (409)
	created, err := h.storage.CreatePeminjaman(req)
	if err != nil {
		if errors.Is(err, ErrConflict) {
			writeError(w, http.StatusConflict, "jadwal bentrok pada slot waktu yang dipilih")
			return
		}
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "sumber daya tidak ditemukan")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, created)
}
