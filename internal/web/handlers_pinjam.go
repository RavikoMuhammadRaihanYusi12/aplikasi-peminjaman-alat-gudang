package web

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"alat-gudang/internal/store"
)

type PinjamPageVM struct {
	ActiveNav   string             `json:"active_nav"`
	Technicians []store.Technician `json:"technicians"`
	Tools       []store.ToolRowVM  `json:"tools"`
	PreToolID   int64              `json:"pre_tool_id"`
	Toast       *store.ToastVM     `json:"toast"`
}

func (s *Server) formPinjam(w http.ResponseWriter, r *http.Request) {
	techs, err := s.store.ActiveTechnicians(r.Context())
	if err != nil {
		http.Error(w, "Gagal memuat teknisi", http.StatusInternalServerError)
		return
	}

	tools, err := s.store.LoanableTools(r.Context())
	if err != nil {
		http.Error(w, "Gagal memuat alat", http.StatusInternalServerError)
		return
	}

	preToolID, _ := strconv.ParseInt(r.URL.Query().Get("alat"), 10, 64)

	vm := PinjamPageVM{
		ActiveNav:   "pinjam",
		Technicians: techs,
		Tools:       tools,
		PreToolID:   preToolID,
		Toast:       GetFlashToast(w, r),
	}

	_ = s.renderer.RenderPage(w, r, "pinjam.html", vm)
}

func (s *Server) postPinjam(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.pinjamError(w, r, http.StatusBadRequest, "Form tidak valid. Coba muat ulang halaman.")
		return
	}

	techID, err := strconv.ParseInt(r.FormValue("technician_id"), 10, 64)
	if err != nil || techID <= 0 {
		s.pinjamError(w, r, http.StatusUnprocessableEntity, "Pilih teknisi terlebih dahulu.")
		return
	}

	toolIDs := make([]int64, 0, len(r.Form["tool_id"]))
	for _, rawID := range r.Form["tool_id"] {
		if id, parseErr := strconv.ParseInt(rawID, 10, 64); parseErr == nil && id > 0 {
			toolIDs = append(toolIDs, id)
		}
	}
	if len(toolIDs) == 0 {
		s.pinjamError(w, r, http.StatusUnprocessableEntity, "Pilih minimal satu alat yang tersedia.")
		return
	}

	loanIDs, err := s.store.Borrow(r.Context(), techID, toolIDs)
	if err != nil {
		s.pinjamError(w, r, http.StatusUnprocessableEntity, borrowErrorMessage(err))
		return
	}

	techName := s.technicianName(r, techID)

	toast := &store.ToastVM{
		Kind:        "success",
		Message:     fmt.Sprintf("Tersimpan: %s meminjam %d alat.", techName, len(loanIDs)),
		Subtext:     "Catatan otomatis disimpan • Bisa dibatalkan jika salah klik",
		UndoLoanIDs: loanIDs,
		UndoKind:    "pinjam",
	}
	SetFlashToast(w, toast)

	if isHTMX(r) {
		w.Header().Set("HX-Trigger", "loan-changed")
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// pinjamError mengirim error dengan cara yang sesuai: form Pinjam adalah form
// biasa (non-htmx), jadi error ditampilkan sebagai flash toast setelah redirect.
func (s *Server) pinjamError(w http.ResponseWriter, r *http.Request, status int, message string) {
	toast := &store.ToastVM{Kind: "danger", Message: message}
	if isHTMX(r) {
		writeErrorToast(w, r, status, message)
		return
	}
	flashRedirect(w, r, toast, "/pinjam")
}

func (s *Server) technicianName(r *http.Request, techID int64) string {
	techs, err := s.store.ActiveTechnicians(r.Context())
	if err != nil {
		return "Teknisi"
	}
	for _, tech := range techs {
		if tech.ID == techID {
			return tech.Name
		}
	}
	return "Teknisi"
}

func borrowErrorMessage(err error) string {
	var unavailable *store.UnavailableError
	switch {
	case errors.As(err, &unavailable):
		return unavailable.Error()
	case errors.Is(err, store.ErrInvalid):
		return "Pilih teknisi dan minimal satu alat dulu."
	case errors.Is(err, store.ErrNotFound):
		return "Data teknisi atau alat tidak ditemukan. Muat ulang halaman."
	default:
		return "Ada gangguan saat menyimpan peminjaman. Coba lagi atau hubungi IT."
	}
}
