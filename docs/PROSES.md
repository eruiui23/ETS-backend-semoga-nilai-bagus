# PROSES

## 1. Persona dan masalahnya
seorang petugas gedung yang bertanggung jawab atas ketersediaan suatu ruangan. Biasanya membuka daftar secara manual ruangan apa yang kosong pada suatu tanggal

## 2. Rancangan struktur data
```go
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

type Handler struct {
	mu          sync.Mutex
	sumberDaya  []SumberDaya
	peminjaman  []Peminjaman
	nextSDID    int
	nextPeminID int
}
```

## 3. Satu keputusan desain
* **Keputusan:** Menyimpan data in-memory langsung menggunakan slice pada struct `Handler` dan menghitung sisa ketersediaan slot secara langsung (*on-the-fly*) dari data peminjaman saat endpoint dipanggil.
* **Pilihan lain:** Membuat abstraksi layer `Storage` terpisah berbasis `map[int]...` dengan tabel/map terpisah untuk kuota ketersediaan.
* **Alasan:** Menyederhanakan struktur kode agar minim *boilerplate* dalam batas waktu pengerjaan 2.5 jam. Penggunaan slice membuat data otomatis terurut berdasarkan ID naik tanpa perlu algoritma sorting tambahan, serta menghindari risiko bug inkonsistensi data kuota.

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
- **Apa yang diminta:** Bagian audit meminta agar data peminjaman tidak boleh hilang saat dihapus. Endpoint `DELETE /peminjaman/{id}` diubah menjadi pembatalan (*soft delete*) dengan mengubah status menjadi `"dibatalkan"`, mengembalikan status `204 No Content`, dan menghasilkan `409 Conflict` jika peminjaman sudah berstatus dibatalkan. Peminjaman yang dibatalkan tidak boleh menahan slot.
- **Bagian kode mana yang kamu ubah:** 
  1. Method `DeletePeminjaman` di `handler.go`: mengubah aksi dari menghapus elemen slice menjadi hanya mengupdate field `Status` menjadi `"dibatalkan"` serta menambahkan validasi konflik status 409 jika status sudah `"dibatalkan"`.
  2. Logika `isBentrok` dan `GetKetersediaan` di `handler.go`: memastikan hanya peminjaman dengan `Status == "aktif"` yang menahan slot dan dihitung kuotanya.
- **Satu keputusan desain yang diambil:** Menerapkan *soft delete* langsung pada slice peminjaman tanpa memindahkan data ke tabel/slice arsip terpisah, sehingga endpoint `GET /peminjaman` dan filter `?status=dibatalkan` dapat langsung membaca data secara terpadu tanpa kompleksitas *querying* lintas koleksi.
