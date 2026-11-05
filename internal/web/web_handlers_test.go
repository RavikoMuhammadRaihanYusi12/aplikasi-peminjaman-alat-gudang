package web

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"alat-gudang/internal/store"
	"alat-gudang/internal/timefmt"

	_ "modernc.org/sqlite"
)

func newTestServer(t *testing.T) (*Server, *store.Store) {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db error: %v", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = db.Close() })

	st, err := store.New(db, timefmt.RealClock{}, "Asia/Jakarta")
	if err != nil {
		t.Fatalf("store.New error: %v", err)
	}
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate error: %v", err)
	}

	renderer, err := NewRenderer(false)
	if err != nil {
		t.Fatalf("NewRenderer error: %v", err)
	}

	return NewServer(st, renderer), st
}

func seedWebData(t *testing.T, st *store.Store) (int64, int64) {
	t.Helper()
	ctx := context.Background()

	tech := &store.Technician{Name: "Budi", Section: "Listrik"}
	if err := st.UpsertTechnician(ctx, tech); err != nil {
		t.Fatalf("seed technician error: %v", err)
	}
	tool := &store.Tool{Code: "MM-01", Name: "Multimeter"}
	if err := st.UpsertTool(ctx, tool); err != nil {
		t.Fatalf("seed tool error: %v", err)
	}
	return tech.ID, tool.ID
}

func doForm(handler http.Handler, method, target string, values url.Values, htmx bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if htmx {
		req.Header.Set("HX-Request", "true")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestHealthzIncludesVersion(t *testing.T) {
	Version = "test-version"
	srv, _ := newTestServer(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/healthz", nil)
	srv.Routes(false).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "ok version=test-version") {
		t.Fatalf("body tidak memuat versi: %q", rec.Body.String())
	}
}

func TestFragmentWithoutHTMXRedirects(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/fragmen/beranda", nil)
	srv.Routes(false).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/" {
		t.Fatalf("expected Location /, got %q", rec.Header().Get("Location"))
	}
}

// fragmenBeranda dipakai sebagai innerHTML swap ke #beranda-live, jadi respons
// tidak boleh memuat wrapper #beranda-live lagi (kalau tidak, wrapper bersarang
// setiap polling 30 detik).
func TestFragmentBerandaHasNoLiveWrapper(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest("GET", "/fragmen/beranda", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	srv.Routes(false).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if got := strings.Count(body, `id="beranda-live"`); got != 0 {
		t.Fatalf("fragmen tidak boleh memuat id=\"beranda-live\", ditemukan %d kali", got)
	}
	if strings.TrimSpace(body) == "" {
		t.Fatal("fragmen beranda kosong")
	}
}

func TestBerandaPageHasSingleLiveWrapper(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	srv.Routes(false).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if got := strings.Count(body, `id="beranda-live"`); got != 1 {
		t.Fatalf("halaman beranda harus punya tepat 1 id=\"beranda-live\", got %d", got)
	}
}

func TestOOBResponseHasSingleLiveWrapper(t *testing.T) {
	srv, st := newTestServer(t)
	techID, toolID := seedWebData(t, st)
	loanIDs, err := st.Borrow(context.Background(), techID, []int64{toolID})
	if err != nil {
		t.Fatalf("Borrow error: %v", err)
	}

	values := url.Values{}
	values.Set("kondisi", "baik")
	rec := doForm(srv.Routes(false), "POST", "/kembali/"+int64String(loanIDs[0]), values, true)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if got := strings.Count(body, `id="beranda-live"`); got != 1 {
		t.Fatalf("respons OOB harus punya tepat 1 id=\"beranda-live\", got %d", got)
	}
	if !strings.Contains(body, `hx-swap-oob="innerHTML"`) {
		t.Fatal("respons harus berisi atribut hx-swap-oob")
	}
	if !strings.Contains(body, `id="toast"`) {
		t.Fatal("respons harus memperbarui toast")
	}
}

func TestDialogWithoutHTMXRedirects(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/kembali/1", nil)
	srv.Routes(false).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rec.Code)
	}
}

func TestPostPinjamValidationRedirects(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := doForm(srv.Routes(false), "POST", "/pinjam", url.Values{}, false)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/pinjam" {
		t.Fatalf("expected Location /pinjam, got %q", rec.Header().Get("Location"))
	}
	if len(rec.Result().Cookies()) == 0 {
		t.Fatal("flash toast cookie harus diset")
	}
}

func TestPostPinjamDuplicateToolRejected(t *testing.T) {
	srv, st := newTestServer(t)
	techID, toolID := seedWebData(t, st)
	handler := srv.Routes(false)

	values := url.Values{}
	values.Set("technician_id", int64String(techID))
	values.Add("tool_id", int64String(toolID))

	if rec := doForm(handler, "POST", "/pinjam", values, false); rec.Code != http.StatusSeeOther {
		t.Fatalf("pinjam pertama expected 303, got %d", rec.Code)
	}

	// Peminjaman kedua harus ditolak dan diarahkan kembali ke form.
	rec := doForm(handler, "POST", "/pinjam", values, false)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("pinjam kedua expected 303, got %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/pinjam" {
		t.Fatalf("expected redirect ke /pinjam, got %q", rec.Header().Get("Location"))
	}

	var count int
	if err := st.DB().QueryRow("SELECT COUNT(*) FROM loans").Scan(&count); err != nil {
		t.Fatalf("query loans error: %v", err)
	}
	if count != 1 {
		t.Fatalf("loan harus tetap 1, got %d", count)
	}
}

func TestPostKembaliInvalidCondition(t *testing.T) {
	srv, st := newTestServer(t)
	techID, toolID := seedWebData(t, st)
	loanIDs, err := st.Borrow(context.Background(), techID, []int64{toolID})
	if err != nil {
		t.Fatalf("Borrow error: %v", err)
	}

	values := url.Values{}
	values.Set("kondisi", "entah")
	rec := doForm(srv.Routes(false), "POST", "/kembali/"+int64String(loanIDs[0]), values, true)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Pilih kondisi alat") {
		t.Fatalf("pesan error tidak sesuai: %q", rec.Body.String())
	}
}

func TestPostKembaliSuccessAndChipAllowlist(t *testing.T) {
	srv, st := newTestServer(t)
	techID, toolID := seedWebData(t, st)
	loanIDs, _ := st.Borrow(context.Background(), techID, []int64{toolID})

	values := url.Values{}
	values.Set("kondisi", "rusak")
	values.Set("chip", "Tidak menyala")
	values.Set("catatan", "Kabel putus")

	rec := doForm(srv.Routes(false), "POST", "/kembali/"+int64String(loanIDs[0]), values, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `hx-swap-oob`) {
		t.Fatalf("respons harus berisi OOB: %q", rec.Body.String())
	}

	var note, condition string
	if err := st.DB().QueryRow("SELECT return_note FROM loans WHERE id = ?", loanIDs[0]).Scan(&note); err != nil {
		t.Fatalf("query note error: %v", err)
	}
	if note != "Tidak menyala - Kabel putus" {
		t.Fatalf("catatan tidak sesuai: %q", note)
	}
	if err := st.DB().QueryRow("SELECT condition FROM tools WHERE id = ?", toolID).Scan(&condition); err != nil {
		t.Fatalf("query condition error: %v", err)
	}
	if condition != "rusak" {
		t.Fatalf("alat harus rusak, got %s", condition)
	}
}

func TestPostBatalInvalidJenis(t *testing.T) {
	srv, st := newTestServer(t)
	techID, toolID := seedWebData(t, st)
	loanIDs, _ := st.Borrow(context.Background(), techID, []int64{toolID})

	values := url.Values{}
	values.Set("loan_id", int64String(loanIDs[0]))
	values.Set("jenis", "entah")

	rec := doForm(srv.Routes(false), "POST", "/batal", values, true)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPostBatalVoidBorrow(t *testing.T) {
	srv, st := newTestServer(t)
	techID, toolID := seedWebData(t, st)
	loanIDs, _ := st.Borrow(context.Background(), techID, []int64{toolID})

	values := url.Values{}
	values.Set("loan_id", int64String(loanIDs[0]))
	values.Set("jenis", "pinjam")

	rec := doForm(srv.Routes(false), "POST", "/batal", values, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var voidedAt sql.NullString
	if err := st.DB().QueryRow("SELECT voided_at FROM loans WHERE id = ?", loanIDs[0]).Scan(&voidedAt); err != nil {
		t.Fatalf("query loan error: %v", err)
	}
	if !voidedAt.Valid {
		t.Fatal("loan harus sudah dibatalkan")
	}
}

func TestCrossOriginPostRejected(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest("POST", "/pinjam", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "http://evil.example")
	req.Host = "127.0.0.1:8080"

	rec := httptest.NewRecorder()
	srv.Routes(false).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAllPagesRender(t *testing.T) {
	srv, st := newTestServer(t)
	techID, toolID := seedWebData(t, st)
	if _, err := st.Borrow(context.Background(), techID, []int64{toolID}); err != nil {
		t.Fatalf("Borrow error: %v", err)
	}

	handler := srv.Routes(false)
	paths := []string{
		"/",
		"/pinjam",
		"/pinjam?alat=" + int64String(toolID),
		"/alat",
		"/alat/" + int64String(toolID),
		"/kelola/alat",
		"/kelola/teknisi",
		"/kelola/pengaturan",
	}

	for _, path := range paths {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("%s: expected 200, got %d (%s)", path, rec.Code, rec.Body.String())
			continue
		}
		if rec.Body.Len() == 0 {
			t.Errorf("%s: body kosong", path)
		}
	}
}

func TestKelolaEditPrefillsForm(t *testing.T) {
	srv, st := newTestServer(t)
	_, toolID := seedWebData(t, st)

	handler := srv.Routes(false)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/kelola/alat?edit="+int64String(toolID), nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Ubah Data Alat", `name="id"`, `value="MM-01"`, "Simpan Perubahan"} {
		if !strings.Contains(body, want) {
			t.Errorf("form ubah alat tidak memuat %q", want)
		}
	}
}

func TestSimpanAlatEdit(t *testing.T) {
	srv, st := newTestServer(t)
	_, toolID := seedWebData(t, st)

	values := url.Values{}
	values.Set("id", int64String(toolID))
	values.Set("code", "MM-01")
	values.Set("name", "Multimeter Digital")
	values.Set("category", "Alat ukur")
	values.Set("max_hours", "36")

	rec := doForm(srv.Routes(false), "POST", "/kelola/alat", values, false)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rec.Code)
	}

	var name, category string
	var maxHours int
	if err := st.DB().QueryRow("SELECT name, category, max_hours FROM tools WHERE id = ?", toolID).
		Scan(&name, &category, &maxHours); err != nil {
		t.Fatalf("query tool error: %v", err)
	}
	if name != "Multimeter Digital" || category != "Alat ukur" || maxHours != 36 {
		t.Fatalf("perubahan alat tidak tersimpan: %s / %s / %d", name, category, maxHours)
	}
}

func TestSimpanTeknisiEdit(t *testing.T) {
	srv, st := newTestServer(t)
	techID, _ := seedWebData(t, st)

	values := url.Values{}
	values.Set("id", int64String(techID))
	values.Set("name", "Budi Santoso")
	values.Set("section", "Mekanik")

	rec := doForm(srv.Routes(false), "POST", "/kelola/teknisi", values, false)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rec.Code)
	}

	var name, section string
	if err := st.DB().QueryRow("SELECT name, section FROM technicians WHERE id = ?", techID).
		Scan(&name, &section); err != nil {
		t.Fatalf("query technician error: %v", err)
	}
	if name != "Budi Santoso" || section != "Mekanik" {
		t.Fatalf("perubahan teknisi tidak tersimpan: %s / %s", name, section)
	}
}

func TestAlatDetailHasMarkDamagedForm(t *testing.T) {
	srv, st := newTestServer(t)
	_, toolID := seedWebData(t, st)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/alat/"+int64String(toolID), nil)
	srv.Routes(false).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Tandai Rusak") {
		t.Fatal("tombol Tandai Rusak tidak ada di detail alat")
	}
	if !strings.Contains(body, "/alat/"+int64String(toolID)+"/rusak") {
		t.Fatal("form tandai rusak tidak mengarah ke endpoint yang benar")
	}
}

func TestPostRusakRejectsBorrowedTool(t *testing.T) {
	srv, st := newTestServer(t)
	techID, toolID := seedWebData(t, st)
	if _, err := st.Borrow(context.Background(), techID, []int64{toolID}); err != nil {
		t.Fatalf("Borrow error: %v", err)
	}

	values := url.Values{}
	values.Set("chip", "Tidak menyala")
	rec := doForm(srv.Routes(false), "POST", "/alat/"+int64String(toolID)+"/rusak", values, true)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "sedang dipinjam") {
		t.Fatalf("pesan error tidak sesuai: %q", rec.Body.String())
	}
}

func TestKelolaInvalidTabReturns404(t *testing.T) {
	srv, _ := newTestServer(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/kelola/entah", nil)
	srv.Routes(false).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func int64String(value int64) string {
	return strconv.FormatInt(value, 10)
}
