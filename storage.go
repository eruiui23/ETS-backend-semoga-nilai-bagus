package main

import (
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
