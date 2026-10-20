# PROSES

## 1. Persona dan masalahnya
seorang petugas gedung yang bertanggung jawab atas ketersediaan suatu ruangan. Biasanya membuka daftar secara manual ruangan apa yang kosong pada suatu tanggal

## 2. Rancangan struktur data
type SumberDaya struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Jenis string `json:"jenis"`
	Stok  int    `json:"stok"`
}

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

type Storage struct {
	mu               sync.Mutex
	sumberDayaMap    map[int]SumberDaya
	peminjamanMap    map[int]Peminjaman
	nextSumberDayaID int
	nextPeminID      int
}

## 3. Satu keputusan desain
* **Keputusan:** Menghitung sisa ketersediaan slot secara langsung (*on-the-fly*) dari data peminjaman saat endpoint dipanggil.
* **Pilihan lain:** Membuat *map* memori terpisah khusus ketersediaan slot yang di-update setiap kali ada `POST`, `PUT`, atau `DELETE`.
* **Alasan:** Menghindari risiko *bug* data tidak sinkron. Menghitung langsung dari sumber utama jauh lebih aman dan lebih cepat dikerjakan dalam waktu terbatas.

## 4. Bug atau kendala (minimal dua)
### Bug 1
- Gejala:
- Cara menemukan:
- Penyebab dan solusi:

### Bug 2
- Gejala:
- Cara menemukan:
- Penyebab dan solusi:

## 5. Change request
(Apa yang diminta, bagian kode mana yang kamu ubah, dan satu keputusan desain yang kamu ambil.)
