package web

import (
	"fmt"
	"net/http"

	"alat-gudang/internal/store"
)

type Server struct {
	store    *store.Store
	renderer *Renderer
}

func NewServer(store *store.Store, renderer *Renderer) *Server {
	return &Server{
		store:    store,
		renderer: renderer,
	}
}

func (s *Server) beranda(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	vm, err := s.store.Dashboard(r.Context(), q)
	if err != nil {
		http.Error(w, "Gagal memuat beranda: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Cek flash toast
	vm.Toast = GetFlashToast(w, r)

	if err := s.renderer.RenderPage(w, r, "beranda.html", vm); err != nil {
		http.Error(w, "Gagal render beranda: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) fragmenBeranda(w http.ResponseWriter, r *http.Request) {
	// Fragmen hanya untuk htmx; akses langsung diarahkan ke Beranda penuh.
	if !isHTMX(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	q := r.URL.Query().Get("q")
	vm, err := s.store.Dashboard(r.Context(), q)
	if err != nil {
		http.Error(w, "Gagal memuat fragmen beranda", http.StatusInternalServerError)
		return
	}

	_ = s.renderer.RenderPartial(w, r, "beranda_content.html", vm)
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	if err := s.store.QuickCheck(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("unhealthy: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "ok version=%s", Version)
}
