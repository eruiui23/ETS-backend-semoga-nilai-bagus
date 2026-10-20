package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Handler mengelola seluruh state in-memory dan HTTP endpoint untuk PinjamKampus.
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

// RegisterRoutes mendaftarkan seluruh endpoint API.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// B1: Kelola Sumber Daya
	mux.HandleFunc("POST /sumber-daya", h.CreateSumberDaya)
	mux.HandleFunc("GET /sumber-daya", h.ListSumberDaya)
	mux.HandleFunc("GET /sumber-daya/{id}", h.GetSumberDayaByID)

	// Persona A: Ketersediaan per slot
	mux.HandleFunc("GET /sumber-daya/{id}/ketersediaan", h.GetKetersediaan)

	// B2, B3, B6, CR-C: Peminjaman
	mux.HandleFunc("POST /peminjaman", h.CreatePeminjaman)
	mux.HandleFunc("GET /peminjaman", h.ListPeminjaman)
	mux.HandleFunc("GET /peminjaman/{id}", h.GetPeminjamanByID)
	mux.HandleFunc("PUT /peminjaman/{id}", h.UpdatePeminjaman)
	mux.HandleFunc("DELETE /peminjaman/{id}", h.DeletePeminjaman)
}

// cariSD mencari SumberDaya berdasarkan ID (dipanggil saat lock aktif).
func (h *Handler) cariSD(id int) (SumberDaya, bool) {
	for _, sd := range h.sumberDaya {
		if sd.ID == id {
			return sd, true
		}
	}
	return SumberDaya{}, false
}

// cariPeminjamanIndex mencari indeks peminjaman berdasarkan ID (dipanggil saat lock aktif).
func (h *Handler) cariPeminjamanIndex(id int) int {
	for i, p := range h.peminjaman {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// isBentrok mengecek apakah ada bentrok pada salah satu slot antara mulai dan selesai (dipanggil saat lock aktif).
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

// ==================== Persona A: Ketersediaan Slot ====================

// GET /sumber-daya/{id}/ketersediaan?tanggal=YYYY-MM-DD
// Urutan validasi:
// 1. ID bukan angka -> 400
// 2. Sumber daya tidak ada -> 404
// 3. Tanggal kosong atau tidak valid -> 400
func (h *Handler) GetKetersediaan(w http.ResponseWriter, r *http.Request) {
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

	tanggal := r.URL.Query().Get("tanggal")
	if tanggal == "" {
		writeError(w, http.StatusBadRequest, "parameter tanggal wajib diisi")
		return
	}
	if _, err := time.Parse("2006-01-02", tanggal); err != nil {
		writeError(w, http.StatusBadRequest, "tanggal harus dalam format YYYY-MM-DD yang valid")
		return
	}

	type slotInfo struct {
		Slot     int  `json:"slot"`
		Sisa     int  `json:"sisa"`
		Tersedia bool `json:"tersedia"`
	}

	slots := make([]slotInfo, 8)
	for i := 1; i <= 8; i++ {
		terpakai := 0
		for _, p := range h.peminjaman {
			if p.SumberDayaID == id && p.Tanggal == tanggal && p.Status == "aktif" {
				if i >= p.SlotMulai && i <= p.SlotSelesai {
					terpakai += p.Jumlah
				}
			}
		}
		sisa := sd.Stok - terpakai
		if sisa < 0 {
			sisa = 0
		}
		slots[i-1] = slotInfo{
			Slot:     i,
			Sisa:     sisa,
			Tersedia: sisa > 0,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"sumber_daya_id": id,
		"tanggal":        tanggal,
		"slot":           slots,
	})
}

// ==================== B2, B3, B6, CR-C: Peminjaman ====================

// POST /peminjaman
func (h *Handler) CreatePeminjaman(w http.ResponseWriter, r *http.Request) {
	var req Peminjaman
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if req.SumberDayaID <= 0 {
		writeError(w, http.StatusBadRequest, "sumber_daya_id harus berupa angka positif")
		return
	}
	sd, found := h.cariSD(req.SumberDayaID)
	if !found {
		writeError(w, http.StatusNotFound, "sumber daya tidak ditemukan")
		return
	}

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

	if req.Status == "aktif" && h.isBentrok(req.SumberDayaID, req.Tanggal, req.SlotMulai, req.SlotSelesai, req.Jumlah, sd.Stok, 0) {
		writeError(w, http.StatusConflict, "jadwal bentrok pada slot waktu yang dipilih")
		return
	}

	req.ID = h.nextPeminID
	h.nextPeminID++
	h.peminjaman = append(h.peminjaman, req)

	writeJSON(w, http.StatusCreated, req)
}

// GET /peminjaman (mendukung filter: status, sumber_daya_id, tanggal, peminjam)
func (h *Handler) ListPeminjaman(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fStatus := q.Get("status")
	fTanggal := q.Get("tanggal")
	fPeminjam := q.Get("peminjam")
	fSDStr := q.Get("sumber_daya_id")

	var fSD int
	if fSDStr != "" {
		fSD, _ = strconv.Atoi(fSDStr)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	res := make([]Peminjaman, 0)
	for _, p := range h.peminjaman {
		if fSD > 0 && p.SumberDayaID != fSD {
			continue
		}
		if fTanggal != "" && p.Tanggal != fTanggal {
			continue
		}
		if fPeminjam != "" && !strings.EqualFold(p.Peminjam, fPeminjam) {
			continue
		}
		if fStatus != "" && !strings.EqualFold(p.Status, fStatus) {
			continue
		}
		res = append(res, p)
	}

	writeJSON(w, http.StatusOK, res)
}

// GET /peminjaman/{id}
func (h *Handler) GetPeminjamanByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "id harus berupa angka")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	idx := h.cariPeminjamanIndex(id)
	if idx == -1 {
		writeError(w, http.StatusNotFound, "peminjaman tidak ditemukan")
		return
	}

	writeJSON(w, http.StatusOK, h.peminjaman[idx])
}

// PUT /peminjaman/{id} (B6)
// Urutan validasi:
// 1. ID bukan angka -> 400
// 2. Data tidak ditemukan -> 404
// 3. Body JSON tidak valid / format salah -> 400
// 4. Cek aturan bisnis / bentrok (tidak bentrok dengan diri sendiri) -> 409
func (h *Handler) UpdatePeminjaman(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "id harus berupa angka")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	idx := h.cariPeminjamanIndex(id)
	if idx == -1 {
		writeError(w, http.StatusNotFound, "peminjaman tidak ditemukan")
		return
	}

	var req Peminjaman
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	if req.SumberDayaID <= 0 {
		writeError(w, http.StatusBadRequest, "sumber_daya_id harus berupa angka positif")
		return
	}
	sd, found := h.cariSD(req.SumberDayaID)
	if !found {
		writeError(w, http.StatusNotFound, "sumber daya tidak ditemukan")
		return
	}

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
		req.Status = h.peminjaman[idx].Status
	}

	// Deteksi bentrok: excludeID = id agar tidak bentrok dengan diri sendiri
	if req.Status == "aktif" && h.isBentrok(req.SumberDayaID, req.Tanggal, req.SlotMulai, req.SlotSelesai, req.Jumlah, sd.Stok, id) {
		writeError(w, http.StatusConflict, "jadwal bentrok pada slot waktu yang dipilih")
		return
	}

	req.ID = id
	h.peminjaman[idx] = req

	writeJSON(w, http.StatusOK, req)
}

// DELETE /peminjaman/{id} (CR-C: Pembatalan / Soft-Delete)
// Urutan validasi:
// 1. ID bukan angka -> 400
// 2. Data tidak ditemukan -> 404
// 3. Status sudah "dibatalkan" -> 409
// Berhasil -> 204 No Content
func (h *Handler) DeletePeminjaman(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "id harus berupa angka")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	idx := h.cariPeminjamanIndex(id)
	if idx == -1 {
		writeError(w, http.StatusNotFound, "peminjaman tidak ditemukan")
		return
	}

	if h.peminjaman[idx].Status == "dibatalkan" {
		writeError(w, http.StatusConflict, "peminjaman sudah dibatalkan sebelumnya")
		return
	}

	// Ubah status menjadi dibatalkan, data tetap ada (soft delete)
	h.peminjaman[idx].Status = "dibatalkan"

	w.WriteHeader(http.StatusNoContent)
}
