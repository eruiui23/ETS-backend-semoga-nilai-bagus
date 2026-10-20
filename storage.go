package main

import (
	"sort"
	"sync"
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
