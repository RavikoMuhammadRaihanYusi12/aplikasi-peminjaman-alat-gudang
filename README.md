# Alat Gudang

Aplikasi pencatatan peminjaman alat gudang untuk **satu PC pos**. Dibangun sebagai
monolit Go yang menyajikan HTML (server-rendered) dengan htmx dan Tailwind CSS,
serta menyimpan data di SQLite. Tidak ada API JSON, tidak ada state di browser,
dan tidak butuh Internet saat berjalan.

- **Target pemakaian:** PC Windows di pos gudang, diakses lewat `http://127.0.0.1:8080`.
- **Versi:** 0.1.0 (MVP).
- **Dokumen acuan:** `documentation/prd.md`, `documentation/planning.md`,
  `documentation/system-design.md`.

## Status fitur

P0 sudah tersedia:

- Beranda "siapa pegang apa" (3 angka besar, perlu dicek, sedang dipinjam, cari, polling 30 detik).
- Pinjamkan alat (pilih teknisi + beberapa alat sekaligus).
- Pengembalian dengan kondisi **Baik** / **Ada kerusakan** + catatan cepat.
- Alat rusak tidak bisa dipinjam sampai ditandai sudah diperbaiki.
- Detail alat + riwayat + pembatalan salah klik.
- Daftar semua alat dan statusnya.
- Kelola alat, teknisi, dan pengaturan batas waktu.

P1 yang sudah tersedia:

- Impor CSV alat dan teknisi lewat CLI.
- Tombol "Tandai Rusak" manual di detail alat.

P1 yang belum dibuat: tampilan per teknisi dan riwayat global.

## Prasyarat

| Kebutuhan | Keterangan |
|---|---|
| Go 1.22+ | Hanya untuk build dari source. |
| Windows 10+ | Target rilis. |
| Node.js/npm | Hanya jika mengubah CSS Tailwind. |
| Browser modern | Edge, Chrome, atau Firefox terbaru. |

## Build dari source

```powershell
# 1. CSS (hanya kalau template berubah)
npx tailwindcss -i ./internal/web/static/input.css -o ./internal/web/static/app.css --minify

# 2. Tes
go test ./...

# 3. Binary rilis
make build-win
```

Jika `make` tidak tersedia:

```powershell
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -trimpath -ldflags "-s -w -X main.Version=0.1.0" -o dist/alat-gudang.exe ./cmd/server
```

Hasil build ada di `dist/alat-gudang.exe`.

## Menjalankan

```powershell
.\dist\alat-gudang.exe
```

Buka `http://127.0.0.1:8080`.

Saat pertama kali dijalankan, aplikasi membuat `data/`, `backup/`, dan `logs/`
di **direktori kerja saat ini**. Database baru dibuat kosong. Isi lewat menu
**Kelola**, atau minta data contoh (hanya untuk demo):

```powershell
.\dist\alat-gudang.exe -seed
```

> `-seed` hanya mengisi tabel yang masih kosong. Jangan jalankan pada database
> produksi yang sudah berisi data.

### Konfigurasi

Prioritas: flag CLI > environment variable > default.

| Flag | Env | Default | Keterangan |
|---|---|---|---|
| `-addr` | `ADDR` | `127.0.0.1:8080` | Alamat HTTP. |
| `-db` | `DB_PATH` | `data/gudang.db` | Lokasi file SQLite. |
| `-backup-dir` | `BACKUP_DIR` | `backup` | Folder backup harian. |
| `-log-dir` | `LOG_DIR` | `logs` | Folder log aplikasi. |
| `-tz` | `TZ` | `Asia/Jakarta` | Zona waktu tampilan. |
| `-seed` | — | mati | Isi data contoh jika database kosong. |
| `-dev` | — | mati | Baca template dari disk (untuk pengembangan). |
| `-import-tools` | — | — | Impor CSV alat saat startup. |
| `-import-technicians` | — | — | Impor CSV teknisi saat startup. |
| `-version` | — | — | Cetak versi lalu keluar. |

**Penting:** path relatif (`data/`, `backup/`, `logs/`) dihitung dari direktori
kerja proses, bukan dari lokasi `.exe`. Untuk autostart, pastikan "Start in"
pintasan atau Task Scheduler menunjuk ke folder aplikasi.

## Impor CSV

Format alat — `code,name,category,max_hours` (header opsional, BOM Excel didukung):

```csv
code,name,category,max_hours
MM-01,Multimeter Digital Fluke,Alat ukur,24
KS-01,Kunci Sok Set 24 pcs,Kunci khusus,
```

Format teknisi — `name,section`:

```csv
name,section
Budi,Listrik
Dedi,Mekanik
```

Jalankan:

```powershell
.\dist\alat-gudang.exe -import-tools .\tools.csv -import-technicians .\teknisi.csv
```

Impor berjalan sebelum server menerima koneksi, lalu aplikasi tetap berjalan.
Baris dengan kode/nama yang sudah ada akan **diperbarui**, bukan diduplikasi.

## Backup dan pemulihan

- Backup dibuat saat startup dan setiap 24 jam ke `backup/gudang-YYYYMMDD-HHMMSS.mmm.db`.
- 14 backup terbaru disimpan; sisanya dihapus otomatis.
- Log harian ada di `logs/` dan disimpan 30 hari.

Pemulihan (target < 15 menit):

1. Tutup aplikasi.
2. Salin backup terbaru menjadi `data/gudang.db`.
3. Hapus `data/gudang.db-wal` dan `data/gudang.db-shm` yang lama (kalau ada).
4. Jalankan aplikasi lagi.

Uji pemulihan satu kali sebelum go-live.

## Deploy di PC pos (Windows)

1. Buat folder tetap, misalnya `C:\AlatGudang`.
2. Salin isi `dist/` ke folder tersebut: `alat-gudang.exe`, `README.txt`, dan
   folder `data/` kosong.
3. Jalankan sekali, lalu tutup, untuk memastikan database dan folder dibuat.
4. Buat pintasan desktop ke `alat-gudang.exe`.
   Di properti pintasan, isi **Start in** dengan `C:\AlatGudang`.
5. Agar menyala otomatis: tekan `Win+R`, jalankan `shell:startup`, lalu salin
   pintasan ke folder Startup. Alternatif: Task Scheduler "At log on" dengan
   **Start in** yang sama.
6. Isi master alat/teknisi (manual atau CSV), pasang stiker kode, dan uji alur
   pinjam → kembali → batal.
7. Salin folder `backup/` secara berkala ke flash disk atau folder jaringan.

## Hosting di jaringan atau publik

Default `127.0.0.1` berarti aplikasi **hanya bisa diakses dari PC itu sendiri**.

- **LAN tepercaya:** jalankan dengan `-addr 0.0.0.0:8080` dan buka port di
  Windows Firewall. Aplikasi belum punya login, jadi siapa pun di jaringan bisa
  mengubah data.
- **VPS/Windows Server:** bisa, asalkan ada disk persisten (SQLite tidak cocok
  untuk filesystem ephemeral/serverless seperti Vercel/Netlify). Jalankan di
  belakang reverse proxy (Caddy/nginx) untuk HTTPS.
- **Internet publik:** **belum siap** tanpa tambahan autentikasi, HTTPS,
  rate limiting, security header, dan backup off-site. Untuk pemakaian saat ini,
  batasi ke localhost atau LAN tepercaya.

Jangan mengekspos aplikasi langsung ke Internet apa adanya.

## Struktur proyek

```
cmd/server/            entrypoint, logging, backup, shutdown
internal/config/       flag dan environment
internal/store/        SQLite, migrasi, validasi, impor CSV, seed
internal/web/          handler, middleware, renderer, template, static
internal/backup/       backup SQLite + retensi
internal/timefmt/      format waktu relatif Indonesia
documentation/         PRD, planning, system design
data/                  database runtime (jangan di-commit)
backup/                hasil backup runtime
logs/                  log runtime
dist/                  paket rilis
```

## Checklist go-live

- [ ] Semua tes lulus: `go test ./...`
- [ ] Database produksi bersih dari data demo.
- [ ] Master alat dan teknisi sudah diisi; stiker kode terpasang.
- [ ] Aplikasi menyala otomatis dan pintasan desktop berfungsi.
- [ ] Alur pinjam ≤ 4 klik dan kembali ≤ 2 klik sudah dicoba orang awam.
- [ ] Backup harian berjalan dan pemulihan sudah diuji sekali.
- [ ] Panduan 1 halaman ditempel di pos.
- [ ] Teknisi dan petugas malam sudah diberi penjelasan singkat.
