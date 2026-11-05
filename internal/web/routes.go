package web

import (
	"net/http"
)

func (s *Server) Routes(devMode bool) http.Handler {
	mux := http.NewServeMux()

	// Beranda
	mux.HandleFunc("GET /{$}", s.beranda)
	mux.HandleFunc("GET /fragmen/beranda", s.fragmenBeranda)

	// Pinjam
	mux.HandleFunc("GET /pinjam", s.formPinjam)
	mux.HandleFunc("POST /pinjam", s.postPinjam)

	// Kembali & Batal
	mux.HandleFunc("GET /kembali/{loanID}", s.dialogKembali)
	mux.HandleFunc("POST /kembali/{loanID}", s.postKembali)
	mux.HandleFunc("POST /batal", s.postBatal)

	// Alat
	mux.HandleFunc("GET /alat", s.daftarAlat)
	mux.HandleFunc("GET /alat/{id}", s.detailAlat)
	mux.HandleFunc("POST /alat/{id}/diperbaiki", s.postDiperbaiki)
	mux.HandleFunc("POST /alat/{id}/rusak", s.postRusak)

	// Kelola
	mux.HandleFunc("GET /kelola", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/kelola/alat", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /kelola/{bagian}", s.kelola)
	mux.HandleFunc("POST /kelola/alat", s.simpanAlat)
	mux.HandleFunc("POST /kelola/teknisi", s.simpanTeknisi)
	mux.HandleFunc("POST /kelola/{jenis}/{id}/aktif", s.ubahAktif)
	mux.HandleFunc("POST /kelola/pengaturan", s.simpanPengaturan)

	// Health check
	mux.HandleFunc("GET /healthz", s.healthz)

	// Static assets
	staticServer := http.StripPrefix("/static/", http.FileServer(StaticFS(devMode)))
	mux.Handle("GET /static/", staticServer)

	// Wrap middlewares: Recover -> Logging -> CrossOrigin -> NoCache -> Handler
	var handler http.Handler = mux
	handler = NoCacheMiddleware(handler)
	handler = CrossOriginMiddleware(handler)
	handler = LoggingMiddleware(handler)
	handler = RecoverMiddleware(handler)

	return handler
}
