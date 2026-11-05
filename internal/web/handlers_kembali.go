package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"alat-gudang/internal/store"
)

// damageChips adalah allowlist catatan cepat kerusakan. Server tidak menerima
// teks chip bebas; hanya nilai di daftar ini yang disimpan sebagai catatan.
var damageChips = map[string]bool{
	"Tidak menyala":   true,
	"Patah/aus":       true,
	"Aksesori hilang": true,
	"Lainnya":         true,
}

type DialogKembaliVM struct {
	Loan struct {
		ID             int64  `json:"id"`
		ToolName       string `json:"tool_name"`
		ToolCode       string `json:"tool_code"`
		TechnicianName string `json:"technician_name"`
	} `json:"loan"`
	BorrowedAtFormatted string
}

func (s *Server) dialogKembali(w http.ResponseWriter, r *http.Request) {
	if !isHTMX(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	loanID, ok := parsePositiveID(r.PathValue("loanID"))
	if !ok {
		writeErrorToast(w, r, http.StatusBadRequest, "ID peminjaman tidak valid.")
		return
	}

	loan, err := s.store.LoanForReturn(r.Context(), loanID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrAlreadyReturned):
			writeErrorToast(w, r, http.StatusConflict, "Alat ini sudah tercatat dikembalikan.")
		case errors.Is(err, store.ErrNotFound):
			writeErrorToast(w, r, http.StatusNotFound, "Data peminjaman tidak ditemukan. Muat ulang halaman.")
		default:
			writeErrorToast(w, r, http.StatusInternalServerError, "Ada gangguan saat memuat data. Coba lagi atau hubungi IT.")
		}
		return
	}

	vm := DialogKembaliVM{}
	vm.Loan.ID = loan.LoanID
	vm.Loan.ToolName = loan.ToolName
	vm.Loan.ToolCode = loan.ToolCode
	vm.Loan.TechnicianName = loan.TechnicianName
	vm.BorrowedAtFormatted = loan.BorrowedAtFormatted

	_ = s.renderer.RenderPartial(w, r, "dialog_kembali.html", vm)
}

func (s *Server) postKembali(w http.ResponseWriter, r *http.Request) {
	loanID, ok := parsePositiveID(r.PathValue("loanID"))
	if !ok {
		writeErrorToast(w, r, http.StatusBadRequest, "ID peminjaman tidak valid.")
		return
	}
	if err := r.ParseForm(); err != nil {
		writeErrorToast(w, r, http.StatusBadRequest, "Form tidak valid. Coba lagi.")
		return
	}

	kondisiStr := strings.TrimSpace(r.FormValue("kondisi"))
	var kondisi store.Condition
	switch kondisiStr {
	case string(store.Baik):
		kondisi = store.Baik
	case string(store.Rusak):
		kondisi = store.Rusak
	default:
		writeErrorToast(w, r, http.StatusUnprocessableEntity, "Pilih kondisi alat: Baik atau Ada kerusakan.")
		return
	}

	note := ""
	if kondisi == store.Rusak {
		note = buildDamageNote(r.FormValue("chip"), r.FormValue("catatan"))
	}

	if err := s.store.Return(r.Context(), loanID, kondisi, note); err != nil {
		switch {
		case errors.Is(err, store.ErrAlreadyReturned):
			// Klik ganda / data sudah basi: beri pesan ramah dan segarkan daftar.
			s.writeOOB(w, r, &store.ToastVM{
				Kind:    "warn",
				Message: "Alat ini sudah tercatat dikembalikan.",
				Subtext: "Daftar sudah diperbarui.",
			}, true)
		case errors.Is(err, store.ErrInvalid):
			writeErrorToast(w, r, http.StatusUnprocessableEntity, "Kondisi alat tidak valid.")
		case errors.Is(err, store.ErrNotFound):
			writeErrorToast(w, r, http.StatusNotFound, "Data peminjaman tidak ditemukan. Muat ulang halaman.")
		default:
			writeErrorToast(w, r, http.StatusInternalServerError, "Ada gangguan saat menyimpan pengembalian. Coba lagi atau hubungi IT.")
		}
		return
	}

	toast := &store.ToastVM{
		Kind:        "success",
		Message:     "Pengembalian berhasil dicatat.",
		Subtext:     "Status alat telah diperbarui • Bisa dibatalkan jika salah klik",
		UndoLoanIDs: []int64{loanID},
		UndoKind:    "kembali",
	}
	if kondisi == store.Rusak {
		toast.Kind = "warn"
		toast.Message = "Pengembalian dicatat dengan kerusakan (status alat menjadi Rusak)."
	}

	s.writeOOB(w, r, toast, true)
}

func (s *Server) postBatal(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeErrorToast(w, r, http.StatusBadRequest, "Form tidak valid. Coba lagi.")
		return
	}

	loanIDs := make([]int64, 0, len(r.Form["loan_id"]))
	for _, rawID := range r.Form["loan_id"] {
		if id, err := strconv.ParseInt(rawID, 10, 64); err == nil && id > 0 {
			loanIDs = append(loanIDs, id)
		}
	}
	if len(loanIDs) == 0 {
		writeErrorToast(w, r, http.StatusBadRequest, "Transaksi yang ingin dibatalkan tidak ditemukan.")
		return
	}

	jenis := strings.TrimSpace(r.FormValue("jenis"))

	var (
		toast *store.ToastVM
		err   error
	)
	switch jenis {
	case "pinjam":
		err = s.store.VoidBorrow(r.Context(), loanIDs)
		toast = &store.ToastVM{
			Kind:    "warn",
			Message: "Pencatatan peminjaman berhasil dibatalkan.",
			Subtext: "Alat kembali berstatus tersedia di rak gudang.",
		}
	case "kembali":
		if len(loanIDs) != 1 {
			writeErrorToast(w, r, http.StatusBadRequest, "Pembatalan pengembalian hanya untuk satu transaksi.")
			return
		}
		err = s.store.VoidReturn(r.Context(), loanIDs[0])
		toast = &store.ToastVM{
			Kind:    "warn",
			Message: "Pengembalian berhasil dibatalkan.",
			Subtext: "Alat kembali tercatat sedang dipinjam.",
		}
	default:
		writeErrorToast(w, r, http.StatusBadRequest, "Jenis pembatalan tidak dikenal.")
		return
	}

	if err != nil {
		switch {
		case errors.Is(err, store.ErrCannotUndo):
			writeErrorToast(w, r, http.StatusConflict, "Tidak bisa dibatalkan karena sudah ada pencatatan lain setelahnya.")
		case errors.Is(err, store.ErrNotFound):
			writeErrorToast(w, r, http.StatusNotFound, "Transaksi tidak ditemukan. Muat ulang halaman.")
		case errors.Is(err, store.ErrInvalid):
			writeErrorToast(w, r, http.StatusUnprocessableEntity, "Transaksi ini tidak bisa dibatalkan.")
		default:
			writeErrorToast(w, r, http.StatusInternalServerError, "Ada gangguan saat membatalkan. Coba lagi atau hubungi IT.")
		}
		return
	}

	s.writeOOB(w, r, toast, true)
}

// buildDamageNote menggabungkan chip dari allowlist dengan catatan bebas.
// Catatan dibatasi 500 karakter agar tidak memenuhi layar.
func buildDamageNote(chip, note string) string {
	chip = strings.TrimSpace(chip)
	note = strings.TrimSpace(note)

	if !damageChips[chip] || chip == "Lainnya" {
		chip = ""
	}

	var result string
	switch {
	case chip != "" && note != "":
		result = chip + " - " + note
	case chip != "":
		result = chip
	default:
		result = note
	}

	if utf8.RuneCountInString(result) > 500 {
		result = string([]rune(result)[:500])
	}
	return result
}
