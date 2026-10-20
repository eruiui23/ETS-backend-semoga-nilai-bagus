package main

import (
	"errors"
	"sort"
	"sync"
)

var (
	ErrNotFound = errors.New("data tidak ditemukan")
	ErrConflict = errors.New("jadwal bentrok pada slot tertentu")
)

// Storage mengelola data in-memory dengan aman terhadap konkurensi.
type Storage struct {
	mu               sync.Mutex
	sumberDayaMap    map[int]SumberDaya
	peminjamanMap    map[int]Peminjaman
	nextSumberDayaID int
	nextPeminID      int
}

// NewStorage membuat dan menginisialisasi instance Storage baru.
func NewStorage() *Storage {
	return &Storage{
		sumberDayaMap:    make(map[int]SumberDaya),
		peminjamanMap:    make(map[int]Peminjaman),
		nextSumberDayaID: 1,
		nextPeminID:      1,
	}
}

// CreateSumberDaya menyimpan data sumber daya baru secara thread-safe.
func (s *Storage) CreateSumberDaya(sd SumberDaya) SumberDaya {
	s.mu.Lock()
	defer s.mu.Unlock()

	sd.ID = s.nextSumberDayaID
	s.nextSumberDayaID++
	s.sumberDayaMap[sd.ID] = sd
	return sd
}

// GetSumberDaya mengambil data sumber daya berdasarkan ID secara thread-safe.
func (s *Storage) GetSumberDaya(id int) (SumberDaya, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sd, ok := s.sumberDayaMap[id]
	return sd, ok
}

// GetAllSumberDaya mengambil semua sumber daya terurut berdasarkan ID naik secara thread-safe.
// Mengembalikan slice kosong (bukan nil) jika belum ada data.
func (s *Storage) GetAllSumberDaya() []SumberDaya {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := make([]SumberDaya, 0, len(s.sumberDayaMap))
	for _, sd := range s.sumberDayaMap {
		res = append(res, sd)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res
}

// hasConflictLocked memeriksa apakah ada bentrok pada salah satu slot antara slotMulai dan slotSelesai.
// Dipanggil saat mutex sudah terkunci.
func (s *Storage) hasConflictLocked(sumberDayaID int, tanggal string, slotMulai, slotSelesai, jumlah, stok, excludeID int) bool {
	for slot := slotMulai; slot <= slotSelesai; slot++ {
		terpakai := 0
		for _, p := range s.peminjamanMap {
			if p.ID == excludeID {
				continue
			}
			if p.SumberDayaID == sumberDayaID && p.Tanggal == tanggal && p.Status == "aktif" {
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

// CreatePeminjaman menyimpan peminjaman baru secara atomik dan thread-safe.
func (s *Storage) CreatePeminjaman(p Peminjaman) (Peminjaman, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sd, ok := s.sumberDayaMap[p.SumberDayaID]
	if !ok {
		return Peminjaman{}, ErrNotFound
	}

	if p.Jumlah > sd.Stok {
		return Peminjaman{}, errors.New("jumlah melebihi stok sumber daya")
	}

	if p.Status == "aktif" && s.hasConflictLocked(p.SumberDayaID, p.Tanggal, p.SlotMulai, p.SlotSelesai, p.Jumlah, sd.Stok, 0) {
		return Peminjaman{}, ErrConflict
	}

	p.ID = s.nextPeminID
	s.nextPeminID++
	s.peminjamanMap[p.ID] = p
	return p, nil
}
