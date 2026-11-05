package web

import (
	"errors"
	"fmt"
	"net/http"

	"alat-gudang/internal/store"
)

type AlatPageVM struct {
	ActiveNav string            `json:"active_nav"`
	Tools     []store.ToolRowVM `json:"tools"`
	Toast     *store.ToastVM    `json:"toast"`
}

func (s *Server) daftarAlat(w http.ResponseWriter, r *http.Request) {
	tools, err := s.store.ListTools(r.Context(), "")
	if err != nil {
		http.Error(w, "Gagal memuat daftar alat: "+err.Error(), http.StatusInternalServerError)
		return
	}

	vm := AlatPageVM{
		ActiveNav: "alat",
		Tools:     tools,
		Toast:     GetFlashToast(w, r),
	}

	_ = s.renderer.RenderPage(w, r, "alat_list.html", vm)
}

func (s *Server) detailAlat(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r.PathValue("id"))
	if !ok {
		http.Error(w, "ID alat tidak valid", http.StatusBadRequest)
		return
	}

	vm, err := s.store.ToolDetail(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, "Data alat tidak ditemukan", http.StatusNotFound)
			return
		}
		http.Error(w, "Gagal memuat detail alat", http.StatusInternalServerError)
		return
	}

	vm.ActiveNav = "alat"
	vm.Toast = GetFlashToast(w, r)

	_ = s.renderer.RenderPage(w, r, "alat_detail.html", vm)
}

func (s *Server) postDiperbaiki(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r.PathValue("id"))
	if !ok {
		writeErrorToast(w, r, http.StatusBadRequest, "ID alat tidak valid.")
		return
	}

	if err := s.store.MarkRepaired(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeErrorToast(w, r, http.StatusNotFound, "Data alat tidak ditemukan.")
			return
		}
		writeErrorToast(w, r, http.StatusInternalServerError, "Gagal menandai alat sudah diperbaiki.")
		return
	}

	toast := &store.ToastVM{
		Kind:    "success",
		Message: "Alat telah ditandai sudah diperbaiki.",
		Subtext: "Status alat kembali Tersedia untuk dipinjamkan.",
	}

	if isHTMX(r) {
		s.writeOOB(w, r, toast, true)
		return
	}

	SetFlashToast(w, toast)
	http.Redirect(w, r, fmt.Sprintf("/alat/%d", id), http.StatusSeeOther)
}

// postRusak menandai alat rusak secara manual (P1): alat ditemukan rusak
// setelah dikembalikan. Belum ada tombol di UI MVP, tetapi endpoint dan
// aturan bisnisnya sudah siap dipakai.
func (s *Server) postRusak(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r.PathValue("id"))
	if !ok {
		writeErrorToast(w, r, http.StatusBadRequest, "ID alat tidak valid.")
		return
	}
	if err := r.ParseForm(); err != nil {
		writeErrorToast(w, r, http.StatusBadRequest, "Form tidak valid. Coba lagi.")
		return
	}

	note := buildDamageNote(r.FormValue("chip"), r.FormValue("catatan"))
	if err := s.store.MarkDamaged(r.Context(), id, note); err != nil {
		var unavailable *store.UnavailableError
		switch {
		case errors.As(err, &unavailable):
			writeErrorToast(w, r, http.StatusUnprocessableEntity, unavailable.Error())
		case errors.Is(err, store.ErrNotFound):
			writeErrorToast(w, r, http.StatusNotFound, "Data alat tidak ditemukan.")
		default:
			writeErrorToast(w, r, http.StatusInternalServerError, "Gagal menandai alat rusak.")
		}
		return
	}

	toast := &store.ToastVM{
		Kind:    "warn",
		Message: "Alat telah ditandai rusak.",
		Subtext: "Alat tidak bisa dipinjam sampai ditandai sudah diperbaiki.",
	}

	if isHTMX(r) {
		s.writeOOB(w, r, toast, true)
		return
	}

	SetFlashToast(w, toast)
	http.Redirect(w, r, fmt.Sprintf("/alat/%d", id), http.StatusSeeOther)
}
