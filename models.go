package main

// SumberDaya merepresentasikan entitas barang atau ruangan yang dapat dipinjam.
type SumberDaya struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Jenis string `json:"jenis"`
	Stok  int    `json:"stok"`
}

// Peminjaman merepresentasikan data transaksi peminjaman sumber daya.
type Peminjaman struct {
	ID           int    `json:"id"`
	SumberDayaID int    `json:"sumber_daya_id"`
	Peminjam     string `json:"peminjam"`
	Tanggal      string `json:"tanggal"`
	SlotMulai    int    `json:"slot_mulai"`
	SlotSelesai  int    `json:"slot_selesai"`
	Jumlah       int    `json:"jumlah"`
	Status       string `json:"status"`
}

// SlotKetersediaan merepresentasikan status kuota ketersediaan pada satu slot.
type SlotKetersediaan struct {
	Slot     int  `json:"slot"`
	Sisa     int  `json:"sisa"`
	Tersedia bool `json:"tersedia"`
}

// KetersediaanResponse adalah respons untuk endpoint ketersediaan Persona A.
type KetersediaanResponse struct {
	SumberDayaID int                `json:"sumber_daya_id"`
	Tanggal      string             `json:"tanggal"`
	Slot         []SlotKetersediaan `json:"slot"`
}
