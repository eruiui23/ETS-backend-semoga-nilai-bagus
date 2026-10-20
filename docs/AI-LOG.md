# AI-LOG

Catat setiap penggunaan AI yang menghasilkan kode. Jika tidak memakai AI sama sekali, tulis satu baris: **tidak memakai AI**.

Log yang jujur tidak mengurangi nilai, sebanyak apa pun AI dipakai. Yang mengurangi nilai adalah log yang tidak sesuai dengan kode, atau isi log yang tidak bisa dijelaskan.

## (Nama fitur atau fungsi, misalnya: Deteksi bentrok slot, B5)
- Prompt: (ringkas, apa yang kamu minta)
- Dipakai: (bagian kode yang kamu ambil, misalnya nama fungsi dan file)
- Diubah: (apa yang kamu ubah atau perbaiki dari hasil AI, dan kenapa)

## Struktur projek
- Prompt: tolong rancang struktur projek
- Dipakai: file main, models, storage, utils
- Diubah: blm ada rencana di ubah karena sesuai namun mungkin menghapus storage.go

## Memberi konteks untuk Ai
- prompt : aturan aturan yang ada di req.md
- dipakai : kedepannya
- diubah : models.go 

## mengerjakan kebutuhan wajib B1 kelola sumber daya
- prompt : mengerjakan B1 sesuai aturan aturan yang ada di req.md
- dipakai : file readme mengenai fitur B1
- diubah : models definisi struct, handler menambah fungsi dan enpoint , main, storage auto increment id

## mengerjakan kebutuhan wajib B2 Buat peminjaman
- prompt : mengerjakan B2 sesuai aturan aturan yang ada di req.md
- dipakai : file readme mengenai fitur B1
- diubah : models definisi struct, handler menambah fungsi dan enpoint , main, storage auto increment id

## Simplifikasi arsitektur & Change Request C
- prompt : simplifikasi backend, implementasi Change Request C data peminjaman tidak boleh hilang, dan fitur Persona A
- dipakai : handler.go, utils.go, models.go, PROSES.md
- diubah : menghapus storage.go dan menggantinya dengan in-memory slice di struct Handler; mengimplementasikan soft delete pada DELETE /peminjaman/{id} (status "dibatalkan", 204, dan 409 jika sudah batal); kalkulasi slot on-the-fly pada Persona A
