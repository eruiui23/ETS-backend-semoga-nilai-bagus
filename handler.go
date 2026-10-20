package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Handler mengelola semua data dan endpoint HTTP untuk PinjamKampus.
type Handler struct {
	mu          sync.Mutex
	sumberDaya  []SumberDaya
	peminjaman  []Peminjaman
	nextSDID    int
	nextPeminID int
}

// NewHandler membuat instance Handler yang siap digunakan (wajib diekspos).
func NewHandler() http.Handler {
	h := &Handler{
		sumberDaya:  make([]SumberDaya, 0),
		peminjaman:  make([]Peminjaman, 0),
		nextSDID:    1,
		nextPeminID: 1,
	}

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Middleware normalisasi trailing slash
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > 1 && strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		mux.ServeHTTP(w, r)
	})
}

// RegisterRoutes mendaftarkan seluruh endpoint.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// B1: Kelola Sumber Daya
	mux.HandleFunc("POST /sumber-daya", h.CreateSumberDaya)
	mux.HandleFunc("GET /sumber-daya", h.ListSumberDaya)
	mux.HandleFunc("GET /sumber-daya/{id}", h.GetSumberDayaByID)

	// B2: Buat Peminjaman
	mux.HandleFunc("POST /peminjaman", h.CreatePeminjaman)
}

// cariSD mencari SumberDaya berdasarkan ID (harus dipanggil saat lock aktif).
func (h *Handler) cariSD(id int) (SumberDaya, bool) {
	for _, sd := range h.sumberDaya {
		if sd.ID == id {
			return sd, true
		}
	}
	return SumberDaya{}, false
}

// isBentrok memeriksa apakah kuota melebihi stok pada slot mana pun (harus dipanggil saat lock aktif).
func (h *Handler) isBentrok(sdID int, tanggal string, mulai, selesai, jumlah, stok, excludeID int) bool {
	for slot := mulai; slot <= selesai; slot++ {
		terpakai := 0
		for _, p := range h.peminjaman {
			if p.ID != excludeID && p.SumberDayaID == sdID && p.Tanggal == tanggal && p.Status == "aktif" {
				if slot >= p.SlotMulai && slot <= p.SlotSelesai {
					terpakai += p.Jumlah
				}
			}
		}
		if terpakai+jumlah > stok {
			return true
		}
	}
	return false
}

// ==================== B1: Kelola Sumber Daya ====================

// POST /sumber-daya
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

	h.mu.Lock()
	req.ID = h.nextSDID
	h.nextSDID++
	h.sumberDaya = append(h.sumberDaya, req)
	h.mu.Unlock()

	writeJSON(w, http.StatusCreated, req)
}

// GET /sumber-daya
func (h *Handler) ListSumberDaya(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Karena slice di-append berurutan, otomatis terurut ID menaik
	writeJSON(w, http.StatusOK, h.sumberDaya)
}

// GET /sumber-daya/{id}
func (h *Handler) GetSumberDayaByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "id harus berupa angka")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	sd, found := h.cariSD(id)
	if !found {
		writeError(w, http.StatusNotFound, "sumber daya tidak ditemukan")
		return
	}

	writeJSON(w, http.StatusOK, sd)
}

// ==================== B2: Buat Peminjaman ====================

// POST /peminjaman
func (h *Handler) CreatePeminjaman(w http.ResponseWriter, r *http.Request) {
	var req Peminjaman
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// 1. Cek keberadaan sumber daya
	if req.SumberDayaID <= 0 {
		writeError(w, http.StatusBadRequest, "sumber_daya_id harus berupa angka positif")
		return
	}
	sd, found := h.cariSD(req.SumberDayaID)
	if !found {
		writeError(w, http.StatusNotFound, "sumber daya tidak ditemukan")
		return
	}

	// 2. Validasi input
	req.Peminjam = strings.TrimSpace(req.Peminjam)
	if req.Peminjam == "" || len(req.Peminjam) > 50 {
		writeError(w, http.StatusBadRequest, "peminjam wajib diisi dan maksimal 50 karakter")
		return
	}
	if _, err := time.Parse("2006-01-02", req.Tanggal); err != nil {
		writeError(w, http.StatusBadRequest, "tanggal harus dalam format YYYY-MM-DD yang valid")
		return
	}
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
	if req.Jumlah < 1 || req.Jumlah > sd.Stok {
		writeError(w, http.StatusBadRequest, "jumlah harus antara 1 sampai stok sumber daya")
		return
	}
	if strings.TrimSpace(req.Status) == "" {
		req.Status = "aktif"
	}

	// 3. Deteksi bentrok (409)
	if req.Status == "aktif" && h.isBentrok(req.SumberDayaID, req.Tanggal, req.SlotMulai, req.SlotSelesai, req.Jumlah, sd.Stok, 0) {
		writeError(w, http.StatusConflict, "jadwal bentrok pada slot waktu yang dipilih")
		return
	}

	// 4. Simpan
	req.ID = h.nextPeminID
	h.nextPeminID++
	h.peminjaman = append(h.peminjaman, req)

	writeJSON(w, http.StatusCreated, req)
}
