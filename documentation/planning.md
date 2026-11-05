# Planning MVP: Pencatatan Peminjaman Alat Gudang (Pak Anto)

> **Status:** Draft v0.1, 5 November 2026
> **Stack:** Go (stdlib `net/http` + `html/template`) · htmx · Tailwind CSS · SQLite
> **Nama kerja aplikasi:** *Alat Gudang*

---

## 1. Ringkasan

### 1.1 Masalah

Peminjaman alat di gudang sparepart dicatat di buku besar. Pencatatan sering tidak lengkap (lupa menulis, lupa tanda tangan, mengembalikan tanpa memberi tahu). Akibatnya Pak Anto tidak tahu siapa memegang alat apa, dan harus keliling pabrik mencari. Contoh nyata: multimeter hilang setengah hari, ternyata ada di laci teknisi yang sedang cuti.

### 1.2 Tujuan MVP (satu kalimat)

> Begitu layar dibuka, Pak Anto langsung tahu **siapa memegang alat apa, sejak kapan, dan alat mana yang sudah terlalu lama belum kembali**. Kalau ada alat rusak, ia tahu **siapa peminjam terakhirnya**.

### 1.3 Kriteria berhasil (dicek 2-4 minggu setelah dipakai)

| # | Kriteria | Cara mengukur |
|---|----------|---------------|
| 1 | Pak Anto bisa menjawab "alat X dipegang siapa?" dalam **< 10 detik** tanpa keliling | Tes langsung di pos |
| 2 | Mencatat peminjaman 1 alat **tanpa mengetik**, **≤ 4 klik** | Uji pakai |
| 3 | Mencatat pengembalian **≤ 2 klik** | Uji pakai |
| 4 | Petugas shift malam bisa memakai tanpa dilatih lebih dari 5 menit | Uji pakai |
| 5 | Jumlah alat "tidak jelas keberadaannya" turun dibanding sebelum aplikasi | Tanya Pak Anto / hitung kejadian keliling |

### 1.3.1 Yang BUKAN tujuan MVP

Mengawasi teknisi, memberi sanksi, mengelola stok sparepart, laporan manajemen. Lihat bagian 6.

---

## 2. Pengguna & prinsip desain

### 2.1 Pengguna

| Pengguna | Kondisi | Yang dibutuhkan |
|----------|---------|-----------------|
| **Pak Anto** (58 th, kepala gudang) | Tidak akrab komputer, mengetik 2 jari, nyaman dengan mouse, tidak suka pintasan keyboard | Layar utama yang langsung menjawab "siapa pegang apa"; input minim |
| **Petugas shift malam** | Memakai komputer pos yang sama, tidak ada Pak Anto | Alur yang sama persis, tidak perlu penjelasan |
| **±12 teknisi** (pagi & malam) | Sering pinjam, sering lupa mencatat | Cara mencatat yang lebih cepat daripada menulis di buku |

### 2.2 Prinsip desain (jadi acuan setiap keputusan UI)

1. **Harus lebih cepat dari buku besar.** Kalau tidak, orang kembali ke buku.
2. **Nol mengetik untuk pekerjaan harian.** Nama dan alat dipilih dengan klik. Mengetik hanya untuk mencari (opsional) dan setup data.
3. **Mouse saja.** Tidak ada pintasan keyboard, drag-and-drop, klik kanan, atau double-click.
4. **Besar dan jelas.** Teks dasar ≥ 18px, tombol utama tinggi ≥ 56px, kontras tinggi. Status tidak hanya lewat warna, selalu ada teks/ikon.
5. **Bahasa netral, bukan menghukum.** Pakai "Perlu dicek" (bukan "Terlambat"), "Peminjam terakhir" (bukan "Pelaku/Penyebab"). Tidak ada ranking, skor, atau daftar "teknisi bermasalah".
6. **Susah salah, mudah dibatalkan.** Ada konfirmasi jelas dan tombol "Batalkan" untuk salah klik.
7. **Satu layar utama.** Semua yang penting ada di Beranda; menu lain sedikit dan jelas.
8. **Jalan di komputer tua, tanpa internet.** Satu file program, semua aset (htmx, CSS) ikut tertanam.

---

## 3. Pertanyaan untuk Pak Anto (urut dari yang paling penting)

**Cara bertanya.** Pak Anto sudah bilang "yang penting jalan dulu" dan sulit menjelaskan lewat chat. Jadi: **datangi gudang 15-20 menit**, lihat langsung cara teknisi meminjam, dan tanyakan 3 pertanyaan teratas sambil menunjukkan sketsa layar di kertas. Sisanya bisa dijawab IT/atasan atau memakai nilai default.

| # | Pertanyaan (versi sederhana untuk Pak Anto) | Kenapa penting | Default bila belum dijawab |
|---|---------------------------------------------|----------------|----------------------------|
| **1** | *"Waktu teknisi pinjam alat, siapa yang mencatat: Bapak/petugas malam, atau teknisi sendiri? Dan apakah teknisi bisa ambil alat tanpa lewat Bapak/petugas?"* | Menentukan seluruh alur. Kalau alat bisa diambil bebas, aplikasi apa pun tidak akan menolong. | Petugas gudang mencatat; teknisi boleh klik sendiri di komputer pos. Alat disimpan di area gudang. |
| **2** | *"Berapa lama sebuah alat dipinjam masih wajar? Apakah sama untuk semua alat, atau beda (mis. multimeter vs bor)?"* | Menentukan kapan alat muncul sebagai "Perlu dicek". Fitur inti. | 24 jam untuk semua alat; bisa diubah per alat. |
| **3** | *"Ada berapa alat yang dipinjamkan? Ada yang jumlahnya kembar (mis. 3 multimeter sama)? Boleh kami tempel stiker kode (MM-01, MM-02)?"* | Alat kembar tanpa kode tidak bisa dibedakan, dan "siapa pegang yang mana" jadi kabur. | Setiap alat punya kode unik; Pak Anto/IT memberi stiker saat setup. |
| **4** | *"Mohon daftar nama teknisi (nama + bagian). Seberapa sering ada yang baru/pindah? Siapa yang akan menambah/mengubah daftarnya?"* | Daftar nama adalah pilihan utama di layar pinjam. | 12 teknisi; Pak Anto yang kelola lewat menu Kelola. |
| **5** | *"Shift malam dijaga berapa orang? Perlu tidak kami catat siapa petugas yang mencatat, atau cukup 'shift pagi/malam'?"* | Menentukan perlu tidaknya langkah "pilih petugas" (menambah klik). | Cukup label shift otomatis dari jam (tanpa pilih petugas). |
| **6** | *"Kalau alat dikembalikan rusak, apa yang Bapak perlu tahu? Cuma nama peminjam terakhir, atau juga catatan rusaknya? Alat rusak boleh dipinjam lagi?"* | Menentukan isi form pengembalian dan status "Rusak". | Catat kondisi (Baik/Rusak) + catatan singkat pilihan; alat rusak tidak bisa dipinjam sampai ditandai sudah diperbaiki. |
| **7** | *(Untuk IT/Pak Anto)* *"Komputer pos pakai Windows apa dan browser apa? Ukuran layar? Tersambung jaringan pabrik/internet?"* | Menentukan versi Tailwind, versi htmx, dan cara memasang. | Windows 10/7+, Chrome/Edge/Firefox terbaru, tanpa internet. |
| **8** | *"Siapa boleh melihat layar ini: hanya di pos, atau teknisi/atasan juga bisa buka dari komputer lain? Perlu ditampilkan di TV/monitor tambahan?"* | Terkait rasa "diawasi" dan cara deploy (localhost vs jaringan). | Hanya dibuka di komputer pos. |
| **9** | *"Alat apa saja yang **sekarang** sedang dipinjam? Perlu data dari buku besar lama dimasukkan?"* | Supaya hari pertama dipakai datanya sudah benar. | Hanya alat yang sedang keluar saat ini yang dimasukkan, riwayat lama tidak. |
| **10** | *"Kalau salah catat (salah pilih nama/alat), siapa yang boleh membetulkan?"* | Menentukan fitur pembatalan dan siapa yang berwenang. | Siapa saja di pos boleh membatalkan; semuanya tetap tercatat (tidak dihapus permanen). |
| **11** | *"Kalau aplikasi ini jalan, apa tanda bahwa Bapak merasa terbantu setelah 2-4 minggu?"* | Menyamakan definisi "berhasil". | Kriteria di bagian 1.3. |

---

## 4. Asumsi (karena belum bisa bertanya)

| # | Asumsi | Kalau ternyata salah |
|---|--------|----------------------|
| A1 | Aplikasi hanya dipakai di **satu komputer pos** (Windows), dibuka lewat browser ke `localhost`. | Ubah bind address ke `0.0.0.0` agar bisa diakses dari jaringan pabrik. |
| A2 | Komputer pos punya browser yang cukup modern (Chrome/Edge/Firefox 2020+). | Tailwind v3.4 + htmx dipilih supaya aman untuk browser agak lama. |
| A3 | **Tidak ada internet** di pos, jadi semua aset di-embed ke dalam program (tanpa CDN). | Tidak ada dampak. |
| A4 | Alat sekitar **30-100 item**, tiap alat bisa diberi kode unik (stiker). | Data awal diisi via impor CSV oleh IT. |
| A5 | Teknisi **±12 orang**, dipilih dari daftar nama (tanpa login, tanpa tanda tangan). | Tambah PIN per teknisi di fase 2. |
| A6 | Pencatatan umumnya dilakukan **petugas gudang** atau teknisi di komputer pos; alat tidak diambil tanpa melewati pos. | Proses fisik perlu dibenahi dulu (bukan masalah software). |
| A7 | Batas "Perlu dicek" default **24 jam**, bisa diatur per alat. | Ubah di menu Kelola. |
| A8 | Satu orang bisa meminjam **beberapa alat sekaligus**. | Sudah didukung. |
| A9 | Shift: **pagi 07:00-19:00, malam 19:00-07:00** (hanya label otomatis). | Ubah di Pengaturan. |
| A10 | Zona waktu **WIB (Asia/Jakarta)**. | Ubah lewat konfigurasi `TZ`. |
| A11 | Tidak ada login. Komputer pos dianggap aman secara fisik. Menu Kelola tidak diberi password di MVP. | Tambah PIN sederhana untuk menu Kelola. |
| A12 | Sparepart (barang habis pakai) **bukan** bagian MVP; yang dilacak hanya **alat pinjaman**. | Dibuat modul terpisah di fase berikutnya. |
| A13 | Satu gudang, satu lokasi. | Tambah tabel lokasi nanti. |
| A14 | Selama 2 minggu pertama, **buku besar tetap dipakai berdampingan** (masa transisi). | Tidak ada dampak teknis. |

---

## 5. Scope MVP

### 5.1 P0: wajib ada di versi pertama

1. **Beranda ("Siapa pegang apa")**
   - 3 angka besar: *Sedang dipinjam*, *Perlu dicek*, *Rusak*.
   - Daftar "Perlu dicek" di atas (alat yang melewati batas waktu), lalu daftar "Sedang dipinjam" (terlama di atas).
   - Setiap baris: kode+nama alat, nama teknisi (bagian), sejak kapan (waktu relatif + waktu pasti), tombol **Dikembalikan**.
   - Kotak cari (opsional) yang memfilter langsung. Daftar menyegarkan diri tiap 30 detik.
2. **Pinjamkan alat** (satu halaman): pilih teknisi (kartu nama besar) → centang satu/lebih alat yang tersedia → **Simpan**.
3. **Dikembalikan**: dialog kecil dengan dua tombol besar **Baik** / **Ada kerusakan**. Jika rusak: pilih catatan cepat (chip) + catatan opsional.
4. **Alat rusak**: ditandai "Rusak", tampil di Beranda beserta **peminjam terakhir**, dan tidak bisa dipinjam sampai ditandai sudah diperbaiki.
5. **Detail alat + riwayat**: siapa saja yang meminjam, kapan, kondisi saat kembali. Ada "Peminjam terakhir" di bagian atas.
6. **Semua Alat**: daftar semua alat dengan status (Tersedia / Dipinjam / Rusak), tombol **Pinjamkan** langsung di baris alat yang tersedia.
7. **Kelola**: tambah/ubah/nonaktifkan alat dan teknisi, ubah batas waktu default.
8. **Pembatalan salah klik** (notifikasi "Batalkan" setelah simpan; juga ada di riwayat alat).

### 5.2 P1: dikerjakan jika waktu cukup

- Tampilan **"Per teknisi"** (kelompokkan: Budi memegang 3 alat).
- **Riwayat global** (100 transaksi terakhir).
- **Impor CSV** alat & teknisi lewat perintah CLI (`-import`).
- Tombol "Tandai rusak" manual di detail alat (alat ditemukan rusak setelah dikembalikan).

### 5.3 Definisi selesai (Definition of Done) MVP

Semua P0 jalan, lolos tes di bagian 14, dikemas jadi satu `.exe`, dan sudah dicoba oleh orang yang mirip Pak Anto (tidak terbiasa komputer).

---

## 6. Fitur yang sengaja TIDAK dibuat (dan alasannya)

| Fitur | Alasan tidak dibuat di MVP |
|-------|----------------------------|
| **Login / akun per teknisi / peran** | Menambah hambatan (lupa password, komputer bersama). Pak Anto minta sederhana dan tidak ingin teknisi merasa diawasi. |
| **Tanda tangan digital** | Tanda tangan di buku justru salah satu yang sering terlewat. Memilih nama dengan klik lebih cepat dan cukup untuk tujuan "tahu siapa pegang". |
| **Notifikasi WhatsApp/email ke teknisi (penagihan otomatis)** | Berisiko terasa seperti mengawasi/menagih, dan butuh integrasi eksternal. Pak Anto cukup melihat layar dan menegur secara manusiawi. |
| **Skor, ranking, sanksi, laporan "teknisi bermasalah"** | Bertentangan dengan keinginan Pak Anto yang tidak ingin menghukum. |
| **Scan barcode/QR** | Butuh stiker, perangkat, dan proses baru. Kode alat sudah cukup untuk MVP. Kandidat fase 2. |
| **Manajemen stok sparepart** | Masalah yang berbeda (barang habis pakai, stok minimum, penerimaan barang), jauh lebih besar, dan bukan yang diminta. |
| **Reservasi/booking alat** | Tidak diminta; menambah proses. |
| **Jadwal kalibrasi/perawatan alat ukur** | Di luar kebutuhan "siapa pegang apa"; bisa jadi fase lanjutan. |
| **Integrasi sistem HR/cuti/ERP** | Kompleks dan butuh akses sistem lain. Cukup lihat siapa pemegang alat di layar, sehingga kasus "teknisi cuti" langsung terlihat. |
| **Upload foto kerusakan** | Menambah penyimpanan dan langkah. Catatan teks singkat cukup untuk MVP. |
| **Grafik, dashboard analitik, ekspor Excel** | Beranda sudah menjawab kebutuhan harian. Ekspor CSV mudah ditambah bila diminta. |
| **Aplikasi mobile / PWA / mode offline** | Dipakai di komputer pos; web biasa sudah cukup. |
| **Multi-gudang / multi-lokasi** | Hanya satu gudang. |
| **Real-time (WebSocket)** | Polling tiap 30 detik sudah cukup dan lebih sederhana. |
| **Approval berjenjang** | Menambah birokrasi, bertentangan dengan "yang gampang saja". |

---

## 7. Desain layar

### 7.1 Navigasi (atas, tombol besar)

`[ Beranda ]  [ Semua Alat ]  [ Kelola ]` + tombol hijau besar **＋ Pinjamkan Alat** selalu terlihat.

### 7.2 Beranda

```
+--------------------------------------------------------------------------+
| ALAT GUDANG       [Beranda] [Semua Alat] [Kelola]     Kamis, 5 Nov 15:03 |
+--------------------------------------------------------------------------+
|  +--------------+  +--------------+  +--------------+                    |
|  |  DIPINJAM    |  | PERLU DICEK  |  |    RUSAK     |   [ + PINJAMKAN ]  |
|  |      7       |  |      2       |  |      1       |   [    ALAT     ]  |
|  +--------------+  +--------------+  +--------------+                    |
|  Cari alat / nama:  [____________________]                               |
+--------------------------------------------------------------------------+
|  (!) PERLU DICEK                                                         |
|  MM-02  Multimeter       Budi (Listrik)    Sejak Sel 09:10 (2 hari)      |
|                                                    [ Dikembalikan ]      |
|  BR-01  Bor Listrik      Andi (Mekanik)    Sejak Rab 20:30 (1 hari)      |
|                                                    [ Dikembalikan ]      |
|--------------------------------------------------------------------------|
|  SEDANG DIPINJAM                                                         |
|  JS-03  Jangka Sorong    Dedi (Mekanik)    Sejak Kam 13:40 (1 jam)       |
|                                                    [ Dikembalikan ]      |
|  ...                                                                     |
|--------------------------------------------------------------------------|
|  RUSAK                                                                   |
|  OL-02  Obeng Listrik    Peminjam terakhir: Rudi, dikembalikan Rab 16:20 |
|         Catatan: tidak menyala                  [ Sudah Diperbaiki ]    |
+--------------------------------------------------------------------------+
```

- Baris "Perlu dicek" memakai latar kuning/oranye **dan** label teks "Perlu dicek".
- Daftar dimuat ulang tiap 30 detik, tetapi **ditunda saat dialog sedang terbuka** agar tidak mengganggu.

### 7.3 Pinjamkan Alat (satu halaman, 3 bagian bernomor)

```
  1. SIAPA YANG MEMINJAM?
  [ Andi ]  [ Budi ]  [ Candra ]  [ Dedi ]  [ Eko ]  [ Fajar ] ...   (kartu besar, pilih satu)

  2. ALAT APA? (boleh lebih dari satu)
  [ ] MM-01 Multimeter     [ ] MM-03 Multimeter    [ ] JS-01 Jangka Sorong
  (alat yang sedang dipinjam/rusak abu-abu, tidak bisa dipilih, ada keterangan siapa pemegangnya)

  3. [ SIMPAN PEMINJAMAN ]    (aktif setelah 1 teknisi + >=1 alat dipilih)
     Ringkasan: "Budi meminjam: Multimeter MM-01, Jangka Sorong JS-01"
```

Setelah simpan: kembali ke Beranda dengan pesan hijau "Tersimpan: Budi meminjam 2 alat. [Batalkan]".

### 7.4 Dialog Dikembalikan

```
  Multimeter MM-02 dikembalikan oleh Budi?
  [   BAIK   ]      [ ADA KERUSAKAN ]
  (jika rusak: chip "Tidak menyala" "Patah/aus" "Aksesori hilang" "Lainnya" + catatan opsional + [Simpan])
```

Klik **Baik** langsung menyimpan, jadi total **2 klik** dari Beranda.

### 7.5 Detail Alat

Judul alat + status + **Peminjam terakhir** + tabel riwayat (Peminjam | Dipinjam | Dikembalikan | Kondisi | Catatan). Ada tombol Batalkan pada transaksi terakhir.

### 7.6 Kelola

Tiga tab sederhana: **Alat**, **Teknisi**, **Pengaturan** (batas waktu default, jam shift). Tidak ada hapus permanen, hanya **Nonaktifkan** supaya riwayat tetap utuh.

### 7.7 Panduan teks (copywriting)

| Hindari | Pakai |
|---------|-------|
| Terlambat / Overdue | Perlu dicek |
| Pelanggaran / Pelaku | Peminjam terakhir |
| Checkout / Return | Dipinjam / Dikembalikan |
| Error 500 | "Ada gangguan. Coba lagi, atau hubungi IT." |

---

## 8. Arsitektur & tech stack

### 8.1 Pilihan teknologi

| Lapisan | Pilihan | Alasan |
|---------|---------|--------|
| Bahasa/server | **Go** (stdlib `net/http` dengan ServeMux ber-method, `html/template`, `embed`) | Satu binary tanpa runtime, mudah dipasang di Windows lama, minim dependensi. |
| Database | **SQLite** via `modernc.org/sqlite` (pure Go, tanpa CGO) + mode WAL | Satu file, tanpa server DB, cukup untuk ±12 pengguna. Pure Go memudahkan cross-compile ke Windows. |
| Frontend | **Server-rendered HTML + htmx** | Tanpa build JS, tanpa SPA, interaksi lewat atribut HTML. |
| Styling | **Tailwind CSS** (standalone CLI, tanpa Node) | Cepat membuat UI besar dan konsisten. Hasil `app.css` ikut di-commit sehingga deploy tidak butuh CLI. Default v3.4 (ramah browser lama); naik ke v4 jika browser pos terbukti modern (Q7). |
| Aset | htmx di-*vendor* ke `static/`, semua di-embed | Tidak butuh internet. |
| Zona waktu | `import _ "time/tzdata"` | Windows tidak selalu punya database zona waktu; ini memastikan `Asia/Jakarta` tersedia. |

### 8.2 Alur request

```
Browser (htmx) --HTTP--> Go server
                           |-- handler (parse form, validasi)
                           |-- store (SQL + aturan bisnis, dalam transaksi)
                           |-- template (halaman penuh ATAU fragmen jika header HX-Request)
                           `--> HTML dikirim kembali, htmx menukar bagian halaman
```

### 8.3 Pola htmx yang dipakai

| Kebutuhan | Pola |
|-----------|------|
| Segarkan daftar otomatis | `<div hx-get="/fragmen/beranda" hx-trigger="every 30s, loan-changed from:body" hx-swap="innerHTML">` |
| Buka dialog Dikembalikan | `<button hx-get="/kembali/12" hx-target="#dialog" hx-swap="innerHTML">` |
| Simpan pengembalian | `<button hx-post="/kembali/12" hx-vals='{"kondisi":"baik"}'>`; server membalas dengan header `HX-Trigger: loan-changed` + pesan toast |
| Cari langsung | `<input hx-get="/fragmen/beranda" hx-trigger="input changed delay:300ms" hx-target="#daftar" name="q">` |
| Angka ringkasan ikut berubah | Respons memakai *out-of-band swap* (`hx-swap-oob`) untuk blok angka |
| Satu template, dua mode | Handler mengecek `HX-Request`: fragmen untuk htmx, halaman penuh untuk akses biasa |

### 8.4 Keamanan (secukupnya untuk MVP)

- Bind default ke `127.0.0.1` (hanya komputer pos). Mengubah ke jaringan harus lewat konfigurasi yang disengaja.
- Perlindungan CSRF ringan: cek header `Origin`/`Sec-Fetch-Site` pada semua POST (Go 1.25+ punya `http.CrossOriginProtection`).
- Query SQL selalu berparameter; `html/template` meng-escape otomatis.
- Tidak ada data pribadi sensitif; hanya nama + bagian.

---

## 9. Model data

```sql
CREATE TABLE technicians (
  id          INTEGER PRIMARY KEY,
  name        TEXT    NOT NULL,
  section     TEXT    NOT NULL DEFAULT '',           -- bagian, mis. Listrik/Mekanik
  active      INTEGER NOT NULL DEFAULT 1,
  created_at  TEXT    NOT NULL                       -- UTC, RFC3339
);

CREATE TABLE tools (
  id             INTEGER PRIMARY KEY,
  code           TEXT    NOT NULL UNIQUE,            -- mis. MM-01
  name           TEXT    NOT NULL,                   -- mis. Multimeter
  category       TEXT    NOT NULL DEFAULT '',        -- Alat ukur / Alat listrik / Kunci khusus
  max_hours      INTEGER,                            -- NULL = pakai default dari settings
  condition      TEXT    NOT NULL DEFAULT 'baik' CHECK (condition IN ('baik','rusak')),
  condition_note TEXT    NOT NULL DEFAULT '',
  active         INTEGER NOT NULL DEFAULT 1,
  created_at     TEXT    NOT NULL
);

CREATE TABLE loans (
  id               INTEGER PRIMARY KEY,
  tool_id          INTEGER NOT NULL REFERENCES tools(id),
  technician_id    INTEGER NOT NULL REFERENCES technicians(id),
  borrowed_at      TEXT    NOT NULL,                 -- UTC
  borrow_shift     TEXT    NOT NULL,                 -- 'pagi' | 'malam' (otomatis dari jam)
  returned_at      TEXT,                             -- NULL = masih dipinjam
  return_shift     TEXT,
  return_condition TEXT    CHECK (return_condition IN ('baik','rusak')),
  return_note      TEXT    NOT NULL DEFAULT '',
  voided_at        TEXT                              -- pembatalan salah catat (tidak dihapus)
);

-- Satu alat hanya boleh punya SATU peminjaman yang masih terbuka.
CREATE UNIQUE INDEX ux_loans_one_open_per_tool
  ON loans(tool_id) WHERE returned_at IS NULL AND voided_at IS NULL;
CREATE INDEX ix_loans_tool_history ON loans(tool_id, borrowed_at DESC);
CREATE INDEX ix_loans_technician   ON loans(technician_id);

CREATE TABLE settings (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
-- default: default_max_hours=24, shift_pagi_start=07:00, shift_malam_start=19:00
```

**Catatan desain data**

- **Status alat tidak disimpan**, melainkan diturunkan: `rusak` jika `tools.condition='rusak'`; `dipinjam` jika ada loan terbuka; selain itu `tersedia`. Ini menghindari status yang tidak sinkron.
- **Tidak ada hapus permanen**; hanya `active=0` dan `voided_at`. Riwayat selalu utuh.
- "Peminjam terakhir" = loan non-void terbaru untuk alat itu (`ORDER BY borrowed_at DESC LIMIT 1`).
- Waktu disimpan UTC, ditampilkan WIB.

---

## 10. Endpoint

| Method | Path | Fungsi |
|--------|------|--------|
| GET | `/` | Beranda (halaman penuh) |
| GET | `/fragmen/beranda?q=` | Fragmen daftar + angka (untuk polling dan pencarian) |
| GET | `/pinjam?alat=ID` | Form pinjam (opsional alat sudah terpilih) |
| POST | `/pinjam` | Simpan peminjaman (`technician_id`, `tool_ids[]`), satu transaksi |
| GET | `/kembali/{loanID}` | Dialog pengembalian |
| POST | `/kembali/{loanID}` | Simpan pengembalian (`kondisi`, `catatan`) |
| POST | `/batal/{loanID}` | Batalkan pencatatan (void) / batalkan pengembalian terakhir |
| GET | `/alat` | Semua alat + status |
| GET | `/alat/{id}` | Detail + riwayat |
| POST | `/alat/{id}/diperbaiki` | Tandai sudah diperbaiki (condition kembali `baik`) |
| GET/POST | `/kelola/alat`, `/kelola/teknisi`, `/kelola/pengaturan` | CRUD ringan |
| GET | `/healthz` | Cek hidup (untuk skrip pemantau) |

---

## 11. Aturan bisnis

1. Alat hanya bisa dipinjam jika `active`, `condition='baik'`, dan tidak punya loan terbuka. Ini dijaga **di dua tempat**: validasi aplikasi + unique index database.
2. Peminjaman beberapa alat dilakukan dalam **satu transaksi** (semua berhasil atau tidak sama sekali) dengan `borrowed_at` yang sama.
3. **Perlu dicek** jika `now - borrowed_at > COALESCE(tool.max_hours, default_max_hours)` jam.
4. Mengembalikan dengan kondisi `rusak` otomatis mengubah `tools.condition='rusak'` dan menyimpan catatan pada loan serta alat.
5. "Sudah diperbaiki" mengubah `condition` menjadi `baik` (loan lama tetap menyimpan fakta kerusakannya).
6. Pembatalan hanya untuk **transaksi terakhir** sebuah alat; pembatalan pengembalian mengosongkan `returned_at`, kondisi, dan catatan; pembatalan peminjaman mengisi `voided_at`. Pembatalan hanya boleh jika alat belum dipinjam orang lain.
7. Teknisi/alat nonaktif tidak muncul di form pinjam, tetapi tetap tampil di riwayat.
8. Waktu ditampilkan relatif dan pasti: "2 hari lalu · Sel 09:10".

---

## 12. Struktur proyek

```
alat-gudang/
├── cmd/server/main.go            # entry point: config, DB, migrasi, backup, HTTP server
├── internal/
│   ├── config/config.go          # ADDR, DB_PATH, TZ, BACKUP_DIR dari env/flag
│   ├── store/                    # SQL + aturan bisnis (Borrow, Return, Void, MarkRepaired, ...)
│   │   ├── store.go
│   │   ├── loans.go · tools.go · technicians.go · settings.go
│   │   ├── migrations/001_init.sql   # di-embed, dijalankan saat start (PRAGMA user_version)
│   │   └── *_test.go
│   ├── web/
│   │   ├── routes.go · handlers_*.go · middleware.go · render.go
│   │   ├── templates/            # layout.html, beranda.html, pinjam.html, _dialog_kembali.html, ...
│   │   └── static/               # htmx.min.js, app.css (hasil build), input.css
│   └── timefmt/                  # "2 hari lalu", shift pagi/malam, zona WIB
├── tailwind.config.js
├── Makefile                      # css, run, test, build-win (GOOS=windows GOARCH=amd64)
├── planning.md                   # dokumen ini
└── README.md                     # cara pasang untuk IT
```

---

## 13. Rencana kerja (milestone)

Estimasi kasar untuk satu pengembang: **±3-4 hari kerja**.

| M | Isi | Estimasi | Hasil yang bisa didemokan |
|---|-----|----------|---------------------------|
| **M0** | Inisialisasi repo, Makefile, Tailwind CLI, vendor htmx, layout dasar, `/healthz` | 0,5 hari | Halaman kosong tampil dengan gaya |
| **M1** | Skema DB + migrasi + `store` (Borrow, Return, Void) + unit test aturan bisnis + data contoh | 1 hari | Test hijau; aturan inti terbukti |
| **M2** | Beranda + fragmen + polling + pencarian + dialog Dikembalikan | 1 hari | Pak Anto bisa lihat siapa pegang apa dan mengembalikan |
| **M3** | Form Pinjamkan + Semua Alat + toast "Batalkan" | 0,5 hari | Alur pinjam-kembali utuh |
| **M4** | Detail alat + riwayat + alur Rusak/Sudah diperbaiki + peminjam terakhir | 0,5 hari | Skenario "alat rusak" jalan |
| **M5** | Menu Kelola (alat, teknisi, pengaturan) + impor CSV (P1) | 0,5 hari | Data awal bisa dimasukkan |
| **M6** | Polesan UI (ukuran, kontras), backup otomatis, build `.exe`, panduan pasang, uji pakai dengan orang awam | 0,5-1 hari | Paket siap dipasang di pos |

Urutan sengaja dibuat supaya **M2 sudah memberi nilai** (inti kebutuhan Pak Anto) dan bisa diperlihatkan lebih awal untuk umpan balik.

---

## 14. Pengujian & kriteria akseptansi

**Otomatis (Go)**
- Unit test `store` dengan SQLite in-memory: pinjam, pinjam alat yang sudah dipinjam (harus gagal), pinjam banyak alat dalam satu transaksi, kembali baik/rusak, alat rusak tak bisa dipinjam, batal pinjam/kembali, hitungan "Perlu dicek" di sekitar batas waktu, penentuan shift.
- Test handler dengan `httptest`: respons fragmen vs halaman penuh, header `HX-Trigger`, penolakan POST lintas-origin.

**Manual (uji pakai, wajib)**
| Skenario | Lulus jika |
|----------|-----------|
| Orang yang tidak biasa komputer meminjamkan 1 alat | ≤ 4 klik, tanpa mengetik, tanpa bantuan |
| Mengembalikan alat | ≤ 2 klik |
| Salah memilih nama lalu membatalkan | Berhasil tanpa panik; riwayat tetap rapi |
| Alat melewati batas waktu | Muncul di bagian "Perlu dicek" dengan teks jelas |
| Alat dikembalikan rusak | Muncul di "Rusak" dengan peminjam terakhir |
| Dua orang mencatat alat yang sama hampir bersamaan | Yang kedua mendapat pesan ramah, data tidak ganda |
| Listrik padam saat berjalan, lalu nyala lagi | Data tidak rusak, aplikasi bisa dijalankan ulang |
| Dibuka di layar kecil/resolusi rendah | Tetap terbaca, tanpa scroll horizontal |

---

## 15. Deployment & operasional

- **Paket:** satu `alat-gudang.exe` (cross-compile `GOOS=windows GOARCH=amd64`) + folder `data/`.
- **Menjalankan:** otomatis saat Windows menyala (Startup folder atau Task Scheduler). Pintasan desktop **"Alat Gudang"** membuka `http://localhost:8080` dalam mode aplikasi (`msedge/chrome --app=...`) sehingga terasa seperti program biasa.
- **Konfigurasi (env):** `ADDR` (default `127.0.0.1:8080`), `DB_PATH` (default `data/gudang.db`), `TZ` (default `Asia/Jakarta`), `BACKUP_DIR`.
- **Backup:** saat start dan tiap 24 jam, salinan konsisten dengan `VACUUM INTO 'backup/gudang-YYYYMMDD.db'`, simpan 14 terakhir. IT disarankan menyalin folder `backup/` ke tempat lain sesekali.
- **Pembaruan:** ganti `.exe`; migrasi DB otomatis saat start.
- **Data awal:** IT mengisi dari daftar alat dan teknisi (CSV atau menu Kelola) sebelum hari pertama, dan memasukkan alat yang sedang dipinjam saat ini.
- **Masa transisi:** 2 minggu berdampingan dengan buku besar; setelah itu buku besar dipensiunkan jika kriteria bagian 1.3 terpenuhi.

---

## 16. Risiko & mitigasi

| Risiko | Dampak | Mitigasi |
|--------|--------|----------|
| Alat keluar tanpa lewat pos / teknisi tetap tidak mencatat | Aplikasi tidak akurat; masalah awal kembali | Ini masalah proses, bukan software. Buat pencatatan lebih cepat dari buku; beri tahu Pak Anto bahwa alat sebaiknya hanya keluar lewat pos; Beranda membuat "lupa" langsung terlihat. |
| Pak Anto/petugas malam kesulitan memakai | Tidak dipakai | UI besar, mouse saja, uji pakai dengan orang awam sebelum rilis, panduan 1 halaman bergambar ditempel di pos. |
| Teknisi merasa diawasi | Resistensi | Bahasa netral, tanpa ranking/skor, layar hanya di pos, dijelaskan bahwa tujuannya mencari alat, bukan menilai orang. |
| Browser komputer lama tidak mendukung CSS modern | Tampilan rusak | Tailwind v3.4, htmx yang kompatibel; cek browser lebih awal (Q7); instal Edge/Chrome/Firefox terbaru bila perlu. |
| Data awal tidak lengkap / alat tanpa kode | Daftar tidak berguna | IT + Pak Anto menyiapkan daftar dan stiker kode sebelum hari pertama. |
| Salah klik menghasilkan data keliru | Kepercayaan turun | Konfirmasi ringkas + tombol Batalkan + riwayat tidak pernah dihapus. |
| File database korup / komputer rusak | Data hilang | WAL + backup harian otomatis + anjuran salin ke lokasi lain. |
| Nama teknisi berubah / orang keluar-masuk | Daftar usang | Nonaktifkan (bukan hapus) agar riwayat aman; Pak Anto/IT bisa mengelola daftar. |
| Alat kembar tanpa penanda | "Siapa pegang yang mana" kabur | Wajib kode unik per alat. |

---

## 17. Fase berikutnya (setelah MVP terbukti dipakai)

1. Scan QR/barcode dengan stiker kode alat.
2. PIN sederhana untuk menu Kelola, dan nama petugas jaga malam bila diperlukan.
3. Tampilan di TV/monitor tambahan untuk melihat status dari jauh.
4. Ekspor CSV/Excel dan ringkasan mingguan untuk atasan.
5. Pengingat kalibrasi/perawatan alat ukur.
6. Modul stok sparepart (proyek terpisah).
7. Pengingat lembut ke teknisi (jika Pak Anto dan teknisi sepakat).

---

## Lampiran A: Checklist kerja

- [ ] M0 Setup proyek dan layout
- [ ] M1 Skema DB, `store`, dan unit test
- [ ] M2 Beranda, polling, pencarian, Dikembalikan
- [ ] M3 Pinjamkan alat, Semua Alat, Batalkan
- [ ] M4 Detail alat, riwayat, Rusak/Diperbaiki
- [ ] M5 Kelola + impor CSV
- [ ] M6 Polesan, backup, build `.exe`, uji pakai, panduan pasang
