# Alat Gudang

Aplikasi web sederhana untuk mencatat **peminjaman alat di gudang sparepart** pabrik, supaya kepala gudang langsung tahu **siapa memegang alat apa, sejak kapan, dan alat mana yang sudah terlalu lama belum kembali**.

## Latar Belakang

Alat yang boleh dipinjam teknisi (kunci khusus, obeng listrik, bor, multimeter, jangka sorong) selama ini dicatat di buku besar. Banyak teknisi lupa menulis atau lupa mengembalikan, sehingga kepala gudang sering harus keliling pabrik mencari alat. Aplikasi ini menggantikan buku besar dengan layar yang mudah dipakai, tanpa membuat teknisi merasa diawasi atau dihukum.

## Fitur Utama

- **Beranda "siapa pegang apa"**: daftar alat yang sedang dipinjam, pemegangnya, dan sejak kapan. Menyegarkan diri otomatis.
- **Penanda "Perlu dicek"**: alat yang sudah lama belum kembali (default 24 jam, bisa diatur per alat).
- **Pinjam alat**: pilih nama teknisi, centang alat (boleh lebih dari satu), simpan. Tanpa mengetik.
- **Kembalikan alat**: dua klik (*Dikembalikan* lalu *Baik*), atau tandai *Ada kerusakan* dengan catatan cepat.
- **Alat rusak + peminjam terakhir**: alat rusak tidak bisa dipinjam sampai ditandai sudah diperbaiki.
- **Riwayat per alat** dan **pencarian** cepat.
- **Batalkan salah catat**: data tidak pernah dihapus, hanya ditandai batal.
- **Kelola alat dan teknisi** serta pengaturan batas waktu dan jam shift.

## Prinsip Desain

- Hanya butuh **mouse**, hampir tanpa mengetik.
- Tulisan dan tombol **besar**, kontras tinggi.
- Bahasa **netral**: "Perlu dicek", "Peminjam terakhir" (bukan "terlambat" atau "pelaku").
- Tanpa login, tanpa ranking atau sanksi, tanpa notifikasi otomatis ke teknisi.
- Berjalan di komputer lama **tanpa internet**.

## Tech Stack

| Lapisan | Teknologi |
|---------|-----------|
| Backend | **Go** (stdlib `net/http`, `html/template`, `embed`), Go 1.22+ |
| Frontend | **htmx** (server-rendered, tanpa SPA dan tanpa build JavaScript) |
| Styling | **Tailwind CSS** (standalone CLI, tanpa Node) |
| Database | **SQLite** (mode WAL) via `modernc.org/sqlite` (pure Go, tanpa CGO) |
| Deployment | Satu binary Go (`alat-gudang.exe`) dengan aset ter-embed, untuk komputer pos Windows 10+ |

## Arsitektur Singkat

```
Browser (htmx) --HTTP--> Go server
                           |-- handler  : parse form, render
                           |-- store    : SQL + aturan bisnis (transaksi)
                           |-- template : halaman penuh / fragmen
                           `-- SQLite   : data/gudang.db (+ backup harian)
```

- Status alat diturunkan dari data peminjaman, jadi tidak bisa tidak sinkron.
- Satu alat hanya boleh punya satu peminjaman terbuka (dijaga unique index di database).
- Tidak ada penghapusan data; pembatalan memakai penanda `voided_at`.
- Waktu disimpan UTC dan ditampilkan dalam WIB.

## Struktur Proyek

```
alat-gudang/
├── cmd/server/main.go      # entry point
├── internal/
│   ├── config/             # konfigurasi
│   ├── store/              # SQL, aturan bisnis, migrasi
│   ├── web/                # handler, template, static
│   ├── backup/             # backup harian
│   └── timefmt/            # waktu relatif, shift, zona WIB
└── Makefile
```

## Menjalankan (Pengembangan)

Prasyarat: Go 1.22+ dan Tailwind CSS standalone CLI.

```bash
make css          # build Tailwind menjadi app.css
make run          # jalankan di http://127.0.0.1:8080
make test         # go vet + go test
make build-win    # build alat-gudang.exe untuk Windows
```

## Konfigurasi

| Variabel | Default | Keterangan |
|----------|---------|------------|
| `ADDR` | `127.0.0.1:8080` | Alamat server |
| `DB_PATH` | `data/gudang.db` | Lokasi database |
| `BACKUP_DIR` | `backup` | Folder backup harian |
| `TZ` | `Asia/Jakarta` | Zona waktu tampilan |

## Deployment

1. Salin `alat-gudang.exe` ke komputer pos (Windows 10 atau lebih baru).
2. Jalankan otomatis saat Windows menyala (Startup folder atau Task Scheduler).
3. Buat pintasan desktop yang membuka `http://localhost:8080`.
4. Backup berjalan otomatis setiap hari. Untuk memulihkan, salin backup terbaru ke `data/gudang.db`.

## Status

MVP dalam pengembangan.

## Lisensi

Belum ditentukan.
