# PRD: Alat Gudang

**Pencatatan peminjaman alat di gudang sparepart**

| | |
|---|---|
| **Versi / Status** | 0.1, Draft untuk ditinjau |
| **Tanggal** | 5 November 2026 |
| **Pemilik produk** | Pak Anto (Kepala Gudang Sparepart) |
| **Tim pembangun** | Divisi IT |
| **Dokumen terkait** | `planning.md` (scope, pertanyaan, asumsi), `system-design.md` (desain teknis) |

---

## 1. Ringkasan

**Alat Gudang** adalah aplikasi web sederhana di komputer pos gudang untuk mencatat peminjaman alat (kunci khusus, obeng listrik, bor, multimeter, jangka sorong, dll.) dan **menampilkan dengan cepat siapa yang sedang memegang alat apa, sejak kapan, serta alat mana yang sudah terlalu lama belum kembali**. Jika alat rusak saat dikembalikan, aplikasi menunjukkan peminjam terakhirnya.

Aplikasi ini menggantikan buku besar yang sering tidak terisi lengkap. Ia dirancang agar bisa dipakai oleh orang yang tidak terbiasa dengan komputer (hanya dengan mouse, hampir tanpa mengetik) dan agar tidak membuat teknisi merasa diawasi atau dihukum.

> **Permintaan asli Pak Anto (WhatsApp ke IT):** *"Tolong dibantu bikinkan yang gampang saja, supaya saya tahu siapa yang lagi pegang alat apa."*

---

## 2. Latar belakang dan masalah

### 2.1 Situasi saat ini

- Gudang sparepart sudah berjalan lebih dari 15 tahun; Pak Anto (58 th) memegangnya hampir 20 tahun dan hafal letak barang, tetapi tidak akrab dengan komputer.
- Ada sekitar **12 teknisi** dari beberapa bagian yang sering meminjam alat, pada **shift pagi dan malam**. Pada shift malam Pak Anto tidak ada; gudang dijaga petugas lain yang memakai komputer pos yang sama.
- Peminjaman dicatat di **buku besar**: nama, alat, tanda tangan, lalu dicoret saat kembali.

### 2.2 Masalah

| # | Masalah | Dampak |
|---|---------|--------|
| 1 | Teknisi lupa menulis, lupa tanda tangan, atau mengembalikan tanpa memberi tahu | Catatan di buku tidak bisa dipercaya |
| 2 | Pak Anto tidak tahu siapa memegang alat tertentu | Harus keliling pabrik mencari alat |
| 3 | Tidak ada tanda otomatis bahwa alat sudah terlalu lama keluar | Alat baru dicari saat sedang dibutuhkan |
| 4 | Saat alat rusak, sulit melacak siapa peminjam terakhir | Sulit menelusuri kejadian dan menindaklanjuti |

### 2.3 Kejadian yang memicu

Minggu lalu sebuah **multimeter yang dibutuhkan mendadak untuk perbaikan mesin tidak ditemukan selama setengah hari**. Ternyata ada di laci seorang teknisi yang sedang cuti.

### 2.4 Mengapa sekarang

Pencarian alat memakan waktu produktif Pak Anto dan teknisi, serta menunda perbaikan mesin. Perbaikannya murah dan cepat dibuat, tetapi hanya akan berhasil jika **lebih mudah daripada buku besar**.

---

## 3. Tujuan, non-tujuan, dan prinsip produk

### 3.1 Tujuan

1. **Visibilitas:** Pak Anto tahu siapa memegang alat apa dan sejak kapan dalam hitungan detik, tanpa keliling.
2. **Peringatan dini yang netral:** alat yang sudah lama belum kembali tampak jelas di layar.
3. **Keterlacakan kerusakan:** peminjam terakhir dari alat rusak selalu bisa dilihat.
4. **Mudah dipakai oleh siapa saja:** Pak Anto, petugas malam, dan teknisi bisa memakainya tanpa pelatihan panjang.
5. **Cepat jalan ("yang penting jalan dulu"):** versi pertama dipakai dalam hitungan hari, bukan bulan.

### 3.2 Non-tujuan (sengaja bukan target MVP)

Mengawasi atau menilai kinerja teknisi · memberi sanksi/denda · mengelola stok sparepart · laporan manajemen dan analitik · integrasi dengan sistem lain · aplikasi mobile. Daftar lengkap dan alasannya ada di `planning.md` bagian 6.

### 3.3 Prinsip produk (menjadi pegangan setiap keputusan)

1. **Lebih cepat dari buku besar**, kalau tidak, orang kembali ke buku.
2. **Nol mengetik** untuk pekerjaan harian; semua dengan klik.
3. **Hanya mouse**, tanpa pintasan keyboard atau gerakan rumit.
4. **Besar dan jelas**: tulisan besar, tombol besar, kontras tinggi, status selalu dengan teks (bukan hanya warna).
5. **Netral, tidak menghukum**: kata "Perlu dicek" (bukan "Terlambat"), "Peminjam terakhir" (bukan "Pelaku"); tidak ada ranking, skor, atau daftar "teknisi bermasalah".
6. **Mudah dibatalkan** bila salah klik.
7. **Satu layar utama** yang menjawab pertanyaan inti.

### 3.4 Pagar pengaman produk (guardrails)

- Aplikasi **tidak** mengirim pesan/penagihan otomatis ke teknisi.
- Aplikasi **tidak** menghasilkan laporan perbandingan antar teknisi.
- Layar hanya dibuka di komputer pos (kecuali diputuskan lain bersama Pak Anto).
- Riwayat **tidak dapat dihapus** (hanya dibatalkan dan tetap tercatat), agar data tetap bisa dipercaya.

---

## 4. Pengguna dan kebutuhan

### 4.1 Persona

**① Pak Anto, Kepala Gudang (pengguna utama, pemilik produk)**

| | |
|---|---|
| Profil | 58 tahun; ±20 tahun di gudang; hafal letak barang |
| Kemampuan komputer | Rendah: komputer hanya untuk surat dan email; mengetik dua jari; lebih nyaman dengan mouse; tidak memakai pintasan keyboard |
| Kebutuhan | Tahu siapa pegang apa dengan cepat; tahu alat mana yang sudah lama keluar; tahu peminjam terakhir bila alat rusak |
| Frustrasi | Keliling pabrik mencari alat; sistem yang rumit |
| Tidak diinginkan | Sistem rumit; menghukum siapa pun; teknisi merasa diawasi berlebihan |
| Ukuran sukses | "Saya tinggal lihat layar, langsung tahu." |

**② Petugas gudang shift malam**

| | |
|---|---|
| Profil | Menjaga gudang saat Pak Anto tidak ada; memakai komputer pos yang sama |
| Kebutuhan | Mencatat pinjam/kembali dengan alur yang sama persis tanpa perlu menelepon Pak Anto |
| Frustrasi | Harus mengurus buku besar dan menebak-nebak status alat |
| Ukuran sukses | Bisa memakai dengan penjelasan ≤ 5 menit |

**③ Teknisi (±12 orang, shift pagi dan malam, beberapa bagian)**

| | |
|---|---|
| Profil | Sering meminjam alat untuk perbaikan mesin; waktunya terbatas |
| Kebutuhan | Meminjam dan mengembalikan secepat mungkin; tidak merasa diawasi |
| Frustrasi | Menulis dan menandatangani buku besar |
| Ukuran sukses | Pinjam alat tanpa antre dan tanpa repot; tidak merasa dinilai |

### 4.2 Jobs-to-be-done

- *Ketika* ada alat yang dibutuhkan mendadak, *saya ingin* langsung tahu siapa pemegangnya, *agar* tidak membuang waktu mencari.
- *Ketika* alat sudah lama keluar, *saya ingin* melihatnya tanpa mencari sendiri, *agar* bisa menanyakannya dengan baik-baik.
- *Ketika* alat dikembalikan rusak, *saya ingin* tahu peminjam terakhirnya, *agar* bisa menanyakan kondisi dan menindaklanjuti.
- *Ketika* saya meminjam alat, *saya ingin* mencatatnya dalam beberapa detik, *agar* tidak menghambat pekerjaan.

---

## 5. Metrik keberhasilan

Metrik dihitung dari data aplikasi sendiri dan pengamatan langsung; **tidak ada pelacakan perilaku pengguna**. Angka target adalah usulan awal untuk disepakati bersama Pak Anto. *Baseline* diukur pada Minggu 0 (sebelum aplikasi dipakai).

### 5.1 Metrik utama

| Metrik | Definisi | Baseline | Target |
|--------|----------|----------|--------|
| **Waktu menemukan pemegang alat** | Waktu Pak Anto menjawab "alat X dipegang siapa?" (uji 5 alat acak) | Diukur Minggu 0 (keliling/membuka buku) | **< 10 detik** |
| **Kejadian "alat dicari"** | Jumlah kejadian Pak Anto harus mencari alat yang tidak jelas keberadaannya per minggu (dicatat Pak Anto) | Diukur Minggu 0 | **Turun ≥ 50%** pada Minggu 4 |
| **Cakupan pencatatan** | Peminjaman yang tercatat di aplikasi ÷ peminjaman yang tercatat di buku besar + yang ketahuan lewat cek langsung (masa dual-run) | - | **≥ 90%** (Minggu 2), **≥ 95%** (Minggu 4) |
| **Akurasi status** | Dari 5 alat acak yang dicek langsung seminggu sekali, jumlah yang statusnya di aplikasi sesuai kenyataan | - | **≥ 90%** |

### 5.2 Metrik kemudahan pakai

| Metrik | Target |
|--------|--------|
| Klik untuk meminjam 1 alat (tanpa mengetik) | ≤ 4 |
| Klik untuk mengembalikan alat | ≤ 2 |
| Waktu mencatat peminjaman 1 alat (median) | ≤ 15 detik |
| Petugas shift malam bisa memakai sendiri setelah penjelasan | ≤ 5 menit |

### 5.3 Metrik pagar pengaman (harus tetap terjaga)

| Metrik | Target |
|--------|--------|
| Teknisi yang menjawab "tidak merasa diawasi berlebihan" (tanya langsung, 3 pertanyaan singkat, Minggu 4) | ≥ 80% |
| Alat rusak yang tercatat dengan peminjam terakhir | 100% |
| Data hilang akibat gangguan (listrik padam, dll.) | 0 transaksi yang sudah tersimpan |

---

## 6. User story dan kriteria penerimaan

Prioritas: **Must** (wajib di MVP) · **Should** (dikerjakan bila waktu cukup) · **Could** (menyusul).
Istilah: *Loan* = satu catatan peminjaman satu alat oleh satu teknisi.

### Epik A: Melihat

**US-1: Beranda "siapa pegang apa"** (Must) — *Sebagai Pak Anto, saya ingin melihat semua alat yang sedang dipinjam beserta pemegang dan waktunya, agar tidak perlu keliling.*
- Saat Beranda dibuka, tampil jumlah *Dipinjam*, *Perlu dicek*, dan *Rusak*.
- Setiap alat yang dipinjam menampilkan kode + nama alat, nama teknisi (dan bagian), serta sejak kapan (relatif dan jam pasti).
- Daftar diurutkan dari yang paling lama dipinjam.
- Daftar menyegarkan diri otomatis tanpa pengguna melakukan apa pun.

**US-2: Penanda "Perlu dicek"** (Must) — *Sebagai Pak Anto, saya ingin alat yang sudah lama keluar terlihat jelas, agar bisa menanyakannya sebelum dibutuhkan.*
- Alat yang melewati batas waktu (default 24 jam, bisa diatur per alat) tampil di bagian teratas dengan label teks **"Perlu dicek"** (bukan hanya warna).
- Bahasa tidak menyalahkan (tanpa kata "terlambat", "pelanggaran").
- Alat kembali → langsung hilang dari bagian "Perlu dicek".

**US-3: Mencari** (Must) — *Sebagai Pak Anto, saya ingin mencari dengan mengetik sedikit, agar cepat menemukan alat atau orang tertentu.*
- Mengetik sebagian kode/nama alat atau nama teknisi memfilter daftar secara langsung.
- Pencarian bersifat opsional; semua fungsi inti tetap bisa dipakai tanpa mengetik.

**US-4: Alat rusak dan peminjam terakhir** (Must) — *Sebagai Pak Anto, saya ingin tahu siapa peminjam terakhir alat yang rusak.*
- Alat berstatus Rusak tampil di bagian "Rusak" di Beranda dengan catatan kerusakan, nama peminjam terakhir, dan waktu dikembalikan.
- Teks menyebut **"Peminjam terakhir"**, bukan "penyebab".
- Alat rusak tidak dapat dipilih pada layar peminjaman.

**US-5: Riwayat alat** (Must) — *Sebagai Pak Anto, saya ingin melihat riwayat peminjaman satu alat.*
- Halaman detail alat menampilkan status saat ini, **peminjam terakhir**, dan daftar riwayat (peminjam, dipinjam, dikembalikan, kondisi, catatan).
- Peminjaman yang dibatalkan tidak muncul sebagai riwayat peminjaman.

**US-6: Semua alat** (Must) — *Sebagai pengguna, saya ingin melihat daftar semua alat beserta statusnya.*
- Daftar menampilkan Tersedia / Dipinjam (oleh siapa) / Rusak.
- Alat yang tersedia punya tombol **Pinjamkan** langsung.

### Epik B: Mencatat

**US-7: Mencatat peminjaman** (Must) — *Sebagai petugas atau teknisi, saya ingin mencatat peminjaman dengan beberapa klik.*
- Alur: pilih nama teknisi → centang satu atau lebih alat → **Simpan**; tanpa mengetik.
- Tombol Simpan baru aktif setelah ada 1 teknisi dan ≥ 1 alat terpilih.
- Alat yang sedang dipinjam atau rusak tidak bisa dipilih dan jelas keterangannya.
- Jika beberapa alat dipinjam sekaligus dan salah satunya gagal, **tidak ada** yang tersimpan setengah jadi.
- Setelah tersimpan tampil pesan jelas (mis. "Tersimpan: Budi meminjam 2 alat") dengan tombol **Batalkan**.
- Jika dua orang mencatat alat yang sama bersamaan, yang kedua mendapat pesan ramah menyebutkan siapa pemegangnya, dan tidak ada data ganda.

**US-8: Mencatat pengembalian** (Must) — *Sebagai petugas atau teknisi, saya ingin mencatat pengembalian dengan sangat cepat.*
- Dari Beranda: klik **Dikembalikan** → klik **Baik** (total 2 klik).
- Klik ganda tidak menghasilkan catatan ganda atau pesan error yang membingungkan.

**US-9: Mencatat alat dikembalikan rusak** (Must) — *Sebagai petugas, saya ingin mencatat kerusakan tanpa banyak mengetik.*
- Pilihan **Ada kerusakan** menampilkan catatan cepat yang bisa diklik (mis. Tidak menyala, Patah/aus, Aksesori hilang, Lainnya) dan kolom catatan opsional.
- Setelah tersimpan, alat otomatis berstatus Rusak dan tidak bisa dipinjam.

**US-10: Alat sudah diperbaiki** (Must) — *Sebagai Pak Anto, saya ingin menandai alat sudah diperbaiki agar bisa dipinjam lagi.*
- Tombol **Sudah Diperbaiki** mengembalikan status alat menjadi Tersedia.
- Riwayat kerusakan sebelumnya tetap tersimpan.

**US-11: Membatalkan salah catat** (Must) — *Sebagai pengguna, saya ingin membetulkan salah klik tanpa panik.*
- Tersedia tombol **Batalkan** setelah mencatat peminjaman atau pengembalian, dan pada riwayat alat.
- Pembatalan hanya diizinkan untuk transaksi terakhir alat itu; jika sudah ada pencatatan lain setelahnya, pengguna mendapat penjelasan yang jelas.
- Data tidak dihapus, hanya ditandai batal.

### Epik C: Mengelola

**US-12: Mengelola daftar alat dan teknisi** (Must) — *Sebagai Pak Anto, saya ingin menambah atau mengubah alat dan teknisi.*
- Menambah/ubah alat (kode unik, nama, kategori, batas waktu khusus) dan teknisi (nama, bagian).
- Alat/teknisi dapat **dinonaktifkan** (bukan dihapus) sehingga tidak muncul di pilihan tetapi riwayatnya tetap ada.
- Kode alat harus unik.

**US-13: Mengatur batas waktu dan jam shift** (Should) — *Sebagai Pak Anto, saya ingin mengatur berapa lama alat dianggap "Perlu dicek".*
- Batas default (jam) dapat diubah; setiap alat boleh punya batas sendiri.
- Jam pergantian shift pagi/malam dapat diubah (hanya sebagai label otomatis pada catatan).

### Epik D: Operasional

**US-14: Dipakai di shift malam tanpa bantuan** (Must) — *Sebagai petugas malam, saya ingin memakai aplikasi yang sama tanpa harus menghubungi Pak Anto.*
- Seluruh alur harian (lihat, pinjam, kembali, batalkan) tersedia tanpa login dan tanpa langkah tambahan.
- Setiap catatan otomatis diberi label shift (pagi/malam) berdasarkan jam.

**US-15: Berjalan otomatis dan aman** (Must) — *Sebagai Pak Anto, saya ingin aplikasi selalu siap dan datanya aman.*
- Aplikasi menyala otomatis saat komputer pos menyala dan dapat dibuka dari pintasan di desktop bernama **"Alat Gudang"**.
- Data dicadangkan otomatis setiap hari (14 salinan terakhir disimpan).
- Setelah listrik padam, data yang sudah tersimpan tetap utuh.

### Cadangan (Should/Could)

| ID | Story | Prioritas |
|----|-------|-----------|
| US-16 | Tampilan **per teknisi** (mis. "Budi memegang 3 alat") | Should |
| US-17 | **Riwayat global** (100 transaksi terakhir) | Should |
| US-18 | **Impor CSV** alat dan teknisi untuk data awal | Should |
| US-19 | **Tandai rusak manual** (alat ditemukan rusak setelah dikembalikan) | Could |

---

## 7. Alur pengguna utama

**Alur 1: Meminjam** (petugas atau teknisi)
`Beranda` → **＋ Pinjamkan Alat** → pilih nama → centang alat → **Simpan** → kembali ke Beranda dengan pesan "Tersimpan… [Batalkan]".

**Alur 2: Mengembalikan**
`Beranda` → **Dikembalikan** pada baris alat → **Baik** *(selesai)*, atau **Ada kerusakan** → pilih catatan → **Simpan**.

**Alur 3: Mencari alat yang dipegang orang lain** (Pak Anto)
Buka layar → lihat bagian **Perlu dicek** dan **Sedang dipinjam**, atau ketik sebagian nama di kotak cari → nama pemegang dan waktunya terlihat.

**Alur 4: Menelusuri alat rusak**
`Beranda` bagian **Rusak** → lihat peminjam terakhir dan catatan → klik nama alat untuk riwayat lengkap → setelah dibetulkan, **Sudah Diperbaiki**.

Gambaran layar (wireframe) ada di `planning.md` bagian 7.

---

## 8. Kebutuhan non-fungsional

| Aspek | Kebutuhan |
|-------|-----------|
| **Kemudahan pakai** | Hanya mouse; teks ≥ 18px; tombol utama ≥ 56px; status dengan teks + ikon; semua dalam Bahasa Indonesia sederhana |
| **Kecepatan** | Layar terasa langsung (respons server < 100 ms; Beranda tampil < 1 detik di komputer lama) |
| **Skala** | ≤ 100 alat, ≤ 20 teknisi, ≈ 30 peminjaman/hari, ≤ 3 pengguna bersamaan |
| **Ketersediaan** | Menyala otomatis; pemulihan setelah gangguan < 15 menit |
| **Keandalan data** | Tidak ada data setengah jadi; riwayat tidak dapat dihapus; cadangan harian (kehilangan maksimal 24 jam bila komputer rusak total) |
| **Kompatibilitas** | Berjalan tanpa internet di Windows 10+ dengan browser Chrome/Edge/Firefox 2020+ |
| **Keamanan** | Hanya diakses dari komputer pos; tidak menyimpan data sensitif (hanya nama dan bagian) |
| **Kemudahan perawatan** | Satu file program; pembaruan cukup mengganti file; panduan pasang 1 halaman |

Rincian teknis ada di `system-design.md`.

---

## 9. Ruang lingkup MVP

**Termasuk (Must):** US-1 s/d US-12, US-14, US-15.
**Termasuk bila waktu cukup (Should):** US-13, US-16, US-17, US-18.
**Menyusul (Could):** US-19.

**Sengaja tidak dibuat:** login/akun, tanda tangan digital, notifikasi otomatis ke teknisi, ranking/skor/sanksi, scan barcode/QR, stok sparepart, reservasi alat, jadwal kalibrasi, integrasi HR/ERP, foto kerusakan, grafik/analitik/ekspor, aplikasi mobile, multi-gudang, approval berjenjang. Alasan masing-masing ada di `planning.md` bagian 6.

**Arah ke depan (Now / Next / Later)**

| Kapan | Isi |
|-------|-----|
| **Now (MVP)** | Epik A-D di atas |
| **Next** | QR/barcode pada stiker alat; PIN untuk menu Kelola; tampilan di TV/monitor tambahan; ekspor CSV |
| **Later** | Pengingat lembut ke teknisi (bila disepakati); jadwal kalibrasi alat ukur; modul stok sparepart (proyek terpisah) |

---

## 10. Asumsi, ketergantungan, dan batasan

### 10.1 Asumsi utama (daftar lengkap di `planning.md` bagian 4)

- Aplikasi dipakai di **satu komputer pos** Windows 10+, tanpa internet.
- Alat ±30-100 buah dan **bisa diberi kode unik** (stiker).
- Teknisi ±12 orang dipilih dari daftar nama; **tanpa login**.
- Pencatatan dilakukan petugas gudang atau teknisi di komputer pos; **alat tidak keluar tanpa melewati pos**.
- Batas "Perlu dicek" default **24 jam**; shift pagi 07:00-19:00, malam 19:00-07:00.
- Selama 2 minggu pertama, **buku besar tetap dipakai berdampingan**.

### 10.2 Ketergantungan

| Ketergantungan | Pemilik | Dibutuhkan sebelum |
|----------------|---------|--------------------|
| Daftar nama teknisi (nama + bagian) | Pak Anto | Pengujian internal |
| Daftar alat + stiker kode | Pak Anto + IT | Hari pertama pilot |
| Alat yang sedang dipinjam saat ini | Pak Anto | Hari pertama pilot |
| Spesifikasi komputer pos (versi Windows, browser) | IT | Mulai pengembangan |
| Pemasangan program, pintasan, dan auto-start di komputer pos | IT | Pilot |

### 10.3 Batasan

- Hanya satu komputer; bila mati, aplikasi tidak tersedia sampai dipulihkan (buku besar tetap menjadi cadangan selama masa transisi).
- Aplikasi tidak dapat memaksa teknisi untuk mencatat; keberhasilannya bergantung pada proses fisik (alat keluar lewat pos).

---

## 11. Pertanyaan terbuka

Daftar lengkap (11 pertanyaan, urut kepentingan) ada di `planning.md` bagian 3. Yang paling menentukan:

| # | Pertanyaan | Ditanyakan ke | Dibutuhkan sebelum |
|---|------------|---------------|--------------------|
| 1 | Siapa yang mencatat (petugas atau teknisi sendiri)? Apakah alat bisa diambil tanpa lewat pos? | Pak Anto | Desain alur final (M2) |
| 2 | Berapa lama alat dianggap "terlalu lama"? Sama untuk semua alat? | Pak Anto | M2 |
| 3 | Jumlah alat dan apakah punya/boleh diberi kode unik; ada alat kembar? | Pak Anto | M3 |
| 4 | Daftar teknisi dan siapa yang mengelola perubahannya | Pak Anto | M3 |
| 5 | Versi Windows dan browser di komputer pos; apakah Windows 7/8? | IT | Mulai M0 |
| 6 | Apakah teknisi/atasan boleh melihat layar, atau hanya di pos? | Pak Anto + atasan | Sebelum pilot |
| 7 | Berapa lama riwayat peminjaman disimpan? (asumsi: tanpa batas, datanya sangat kecil) | Pak Anto/atasan | Sebelum pilot |

**Cara bertanya yang disarankan:** datangi gudang ±15-20 menit, amati cara teknisi meminjam, tunjukkan sketsa layar di kertas, dan tanyakan 3 pertanyaan teratas. Hindari pertanyaan panjang lewat chat. Bila belum terjawab, pakai nilai default di `planning.md`.

---

## 12. Risiko dan mitigasi

| Risiko | Peluang | Dampak | Mitigasi |
|--------|---------|--------|----------|
| Alat tetap keluar tanpa dicatat | Sedang | Tinggi | Mencatat dibuat lebih cepat dari buku; Beranda membuat "lupa mencatat" terlihat; sepakati bahwa alat keluar lewat pos |
| Pak Anto atau petugas malam kesulitan memakai | Sedang | Tinggi | UI besar dan mouse-only; uji pakai dengan orang awam sebelum rilis; panduan 1 halaman bergambar ditempel di pos |
| Teknisi merasa diawasi dan menolak | Sedang | Tinggi | Bahasa netral; tanpa ranking/sanksi/notifikasi otomatis; pengumuman jelas bahwa tujuannya mencari alat; survei singkat Minggu 4 |
| Komputer lama/browser lama tidak cocok | Rendah-Sedang | Sedang | Cek spesifikasi sejak awal; pilih teknologi yang ramah browser lama; fallback server di PC lain |
| Data awal tidak lengkap (alat tanpa kode, nama keliru) | Sedang | Sedang | Persiapan Minggu 0: daftar alat, stiker, dan daftar teknisi disiapkan bersama Pak Anto |
| Data rusak atau komputer mati | Rendah | Tinggi | Cadangan harian otomatis; salinan berkala ke tempat lain; buku besar tetap ada selama transisi |
| Salah klik menghasilkan data keliru | Sedang | Rendah | Konfirmasi ringkas, tombol Batalkan, riwayat tidak pernah dihapus |
| Pak Anto sulit menjelaskan kebutuhan lebih jauh | Tinggi | Sedang | Observasi langsung, demo cepat, iterasi kecil, pakai nilai default yang tercatat |

---

## 13. Rencana rilis

### 13.1 Tahapan

| Tahap | Isi | Perkiraan |
|-------|-----|-----------|
| **0. Persiapan** | Observasi di gudang; jawaban 3 pertanyaan teratas; ukur *baseline* (catat kejadian "alat dicari"); kumpulkan daftar alat dan teknisi; tempel stiker kode | 1-2 hari |
| **1. Pengembangan MVP** | M0-M6 di `planning.md` bagian 13 (demo awal sudah bisa setelah Beranda + Dikembalikan) | ±3-4 hari kerja |
| **2. Uji pakai** | Dicoba oleh orang yang tidak biasa komputer; perbaikan tampilan | 1 hari |
| **3. Pilot (dual-run)** | Aplikasi dipakai **berdampingan dengan buku besar** di semua shift; penjelasan 10 menit untuk teknisi dan petugas malam | Minggu 1-2 |
| **4. Peralihan** | Jika kriteria terpenuhi, buku besar dipensiunkan; aplikasi menjadi sumber utama | Minggu 3-4 |
| **5. Evaluasi** | Tinjau metrik dan masukan; tetapkan isi fase berikutnya | Hari ke-14 dan ke-30 |

*Jadwal indikatif: pengembangan selesai kira-kira pertengahan November 2026, tergantung jawaban Pak Anto dan ketersediaan data awal.*

### 13.2 Kriteria siap rilis (go-live checklist)

- [ ] Seluruh story Must lulus kriteria penerimaan (bagian 6)
- [ ] Uji pakai dengan orang awam lulus: pinjam ≤ 4 klik, kembali ≤ 2 klik, tanpa bantuan
- [ ] Data awal terisi: alat, teknisi, dan alat yang sedang dipinjam saat ini
- [ ] Stiker kode terpasang pada semua alat
- [ ] Program terpasang, menyala otomatis, pintasan "Alat Gudang" ada di desktop
- [ ] Cadangan harian berjalan dan pemulihan sudah dicoba satu kali
- [ ] Panduan 1 halaman bergambar ditempel di pos
- [ ] Teknisi dan petugas malam sudah mendapat penjelasan 10 menit (tujuan: mencari alat, bukan menilai orang)

### 13.3 Kriteria menghentikan buku besar (akhir pilot)

- Cakupan pencatatan ≥ 95% selama minimal 1 minggu
- Akurasi status ≥ 90% pada pengecekan mingguan
- Pak Anto dan petugas malam menyatakan nyaman
- Tidak ada masalah data yang belum terselesaikan

Jika tidak terpenuhi, pilot diperpanjang dan penyebabnya ditelusuri (proses, pelatihan, atau aplikasi).

### 13.4 Manajemen perubahan

- Penjelasan singkat dan netral kepada teknisi: *"Ini supaya alat gampang dicari, bukan untuk menilai siapa pun."*
- Satu jalur umpan balik sederhana (Pak Anto menyampaikan ke IT lewat WhatsApp).
- Perbaikan kecil dirilis cepat selama pilot agar pengguna merasa didengar.

### 13.5 Dukungan dan pemeliharaan

- IT menjadi penanggung jawab teknis (pemasangan, pembaruan, pemulihan).
- Pembaruan cukup mengganti satu file program; cadangan otomatis tetap berjalan.
- Salinan cadangan disarankan disalin berkala ke flash disk/folder jaringan oleh IT.

---

## 14. Pemangku kepentingan

| Peran | Pihak | Tanggung jawab |
|-------|-------|----------------|
| Pemilik produk / pengguna utama | Pak Anto | Menentukan kebutuhan, menerima hasil, mengelola daftar alat dan teknisi |
| Pengguna | Petugas gudang shift malam | Memakai harian; memberi masukan |
| Pengguna | Teknisi (±12) | Meminjam dan mengembalikan; memberi masukan |
| Pembangun dan pemelihara | Divisi IT | Membangun, memasang, mencadangkan, mendukung |
| Sponsor / pihak yang perlu diberi tahu | Atasan gudang/pemeliharaan *(asumsi; perlu dikonfirmasi)* | Menyetujui penggunaan dan kebijakan layar/riwayat |

---

## 15. Glosarium

| Istilah | Arti |
|---------|------|
| **Loan / peminjaman** | Satu catatan satu alat dipinjam satu teknisi, dari dipinjam sampai dikembalikan |
| **Perlu dicek** | Alat yang sudah melewati batas waktu peminjaman; sengaja tidak disebut "terlambat" |
| **Peminjam terakhir** | Teknisi pada catatan peminjaman terbaru suatu alat |
| **Rusak** | Status alat yang ditandai rusak sampai ditandai "Sudah Diperbaiki" |
| **Batalkan (void)** | Membatalkan salah catat; data tetap tersimpan namun tidak dihitung |
| **Dual-run** | Masa transisi memakai aplikasi dan buku besar berdampingan |
| **Shift pagi / malam** | Label otomatis pada catatan berdasarkan jam (default 07:00-19:00 / 19:00-07:00) |

---

## 16. Riwayat revisi

| Versi | Tanggal | Perubahan |
|-------|---------|-----------|
| 0.1 | 5 Nov 2026 | Draf pertama PRD |
