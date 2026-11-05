package web

import (
	"bytes"
	"html"
	"net/http"
	"strconv"

	"alat-gudang/internal/store"
)

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func parsePositiveID(raw string) (int64, bool) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// flashRedirect menyimpan pesan sekali-pakai lalu mengarahkan pengguna kembali.
// Dipakai oleh form non-htmx (Kelola, Pinjam) agar error tidak tampil sebagai teks mentah.
func flashRedirect(w http.ResponseWriter, r *http.Request, toast *store.ToastVM, target string) {
	SetFlashToast(w, toast)
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// writeOOB mengirim satu respons htmx yang memperbarui toast dan (opsional)
// seluruh area Beranda dalam satu round-trip.
func (s *Server) writeOOB(w http.ResponseWriter, r *http.Request, toast *store.ToastVM, includeDashboard bool) {
	var buf bytes.Buffer

	if includeDashboard {
		vm, err := s.store.Dashboard(r.Context(), "")
		if err == nil {
			vm.Toast = toast
			buf.WriteString(`<div id="beranda-live" hx-swap-oob="innerHTML">`)
			_ = s.renderer.RenderPartial(&buf, r, "beranda_content.html", vm)
			buf.WriteString(`</div>`)
		}
	}

	buf.WriteString(`<div id="toast" hx-swap-oob="innerHTML">`)
	_ = s.renderer.RenderPartial(&buf, r, "toast.html", map[string]any{"Toast": toast})
	buf.WriteString(`</div>`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "loan-changed")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

// writeErrorToast mengirim toast error untuk request htmx, atau fallback teks biasa.
func writeErrorToast(w http.ResponseWriter, r *http.Request, status int, message string) {
	if !isHTMX(r) {
		http.Error(w, message, status)
		return
	}

	toast := &store.ToastVM{Kind: "danger", Message: message}
	var buf bytes.Buffer
	buf.WriteString(`<div id="toast" hx-swap-oob="innerHTML">`)
	// Render manual agar tetap bekerja walau renderer tidak tersedia di helper ini.
	buf.WriteString(`<div class="rounded-2xl p-5 mb-4 shadow-lg border-2 flex items-center justify-between gap-4 bg-rose-50 border-rose-500 text-rose-950">`)
	buf.WriteString(`<div class="text-xl font-bold leading-tight">`)
	buf.WriteString(html.EscapeString(toast.Message))
	buf.WriteString(`</div></div></div>`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Reswap", "none")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}
