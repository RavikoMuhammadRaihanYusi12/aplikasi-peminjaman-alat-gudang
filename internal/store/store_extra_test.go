package store

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestBorrowDeduplicatesToolIDs(t *testing.T) {
	st, _ := setupTestStore(t)
	techID, toolID, _ := seedTestEntities(t, st)
	ctx := context.Background()

	loanIDs, err := st.Borrow(ctx, techID, []int64{toolID, toolID, toolID})
	if err != nil {
		t.Fatalf("Borrow dengan ID duplikat error: %v", err)
	}
	if len(loanIDs) != 1 {
		t.Fatalf("ID duplikat harus menjadi 1 loan, got %d", len(loanIDs))
	}
}

func TestBorrowInactiveTechnician(t *testing.T) {
	st, _ := setupTestStore(t)
	techID, toolID, _ := seedTestEntities(t, st)
	ctx := context.Background()

	if err := st.SetTechnicianActive(ctx, techID, false); err != nil {
		t.Fatalf("SetTechnicianActive error: %v", err)
	}

	_, err := st.Borrow(ctx, techID, []int64{toolID})
	var unavailable *UnavailableError
	if !errors.As(err, &unavailable) || unavailable.Reason != "nonaktif" {
		t.Fatalf("Expected UnavailableError nonaktif, got %v", err)
	}
}

func TestVoidBorrowMakesToolAvailableAgain(t *testing.T) {
	st, _ := setupTestStore(t)
	techID, toolID, _ := seedTestEntities(t, st)
	ctx := context.Background()

	loanIDs, err := st.Borrow(ctx, techID, []int64{toolID})
	if err != nil {
		t.Fatalf("Borrow error: %v", err)
	}

	if err := st.VoidBorrow(ctx, loanIDs); err != nil {
		t.Fatalf("VoidBorrow error: %v", err)
	}

	var voidedAt string
	if err := st.db.QueryRowContext(ctx, "SELECT voided_at FROM loans WHERE id = ?", loanIDs[0]).Scan(&voidedAt); err != nil {
		t.Fatalf("query voided_at error: %v", err)
	}
	if voidedAt == "" {
		t.Fatal("voided_at harus terisi")
	}

	// Alat harus bisa dipinjam lagi setelah pembatalan.
	if _, err := st.Borrow(ctx, techID, []int64{toolID}); err != nil {
		t.Fatalf("Borrow setelah void harus berhasil: %v", err)
	}
}

func TestVoidReturnBlockedWhenNewerLoanExists(t *testing.T) {
	st, _ := setupTestStore(t)
	techID, toolID, _ := seedTestEntities(t, st)
	ctx := context.Background()

	loanIDs, _ := st.Borrow(ctx, techID, []int64{toolID})
	firstLoan := loanIDs[0]
	if err := st.Return(ctx, firstLoan, Baik, ""); err != nil {
		t.Fatalf("Return error: %v", err)
	}

	if _, err := st.Borrow(ctx, techID, []int64{toolID}); err != nil {
		t.Fatalf("Borrow ulang error: %v", err)
	}

	err := st.VoidReturn(ctx, firstLoan)
	if !errors.Is(err, ErrCannotUndo) {
		t.Fatalf("Expected ErrCannotUndo, got %v", err)
	}
}

func TestLoanForReturn(t *testing.T) {
	st, _ := setupTestStore(t)
	techID, toolID, _ := seedTestEntities(t, st)
	ctx := context.Background()

	loanIDs, _ := st.Borrow(ctx, techID, []int64{toolID})
	vm, err := st.LoanForReturn(ctx, loanIDs[0])
	if err != nil {
		t.Fatalf("LoanForReturn error: %v", err)
	}
	if vm.ToolCode != "MM-01" || vm.TechnicianName != "Budi" {
		t.Fatalf("Data dialog tidak sesuai: %+v", vm)
	}

	if err := st.Return(ctx, loanIDs[0], Baik, ""); err != nil {
		t.Fatalf("Return error: %v", err)
	}
	if _, err := st.LoanForReturn(ctx, loanIDs[0]); !errors.Is(err, ErrAlreadyReturned) {
		t.Fatalf("Expected ErrAlreadyReturned, got %v", err)
	}
	if _, err := st.LoanForReturn(ctx, 9999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

func TestUpsertToolDuplicateAndUpdate(t *testing.T) {
	st, _ := setupTestStore(t)
	techID, toolID, _ := seedTestEntities(t, st)
	ctx := context.Background()
	_ = techID

	if err := st.UpsertTool(ctx, &Tool{Code: "MM-01", Name: "Duplikat"}); !errors.Is(err, ErrDuplicateCode) {
		t.Fatalf("Expected ErrDuplicateCode, got %v", err)
	}

	if err := st.UpsertTool(ctx, &Tool{ID: toolID, Code: "MM-01", Name: "Multimeter Baru"}); err != nil {
		t.Fatalf("Update tool error: %v", err)
	}
	if err := st.UpsertTool(ctx, &Tool{ID: 9999, Code: "XX-99", Name: "Hilang"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expected ErrNotFound untuk update, got %v", err)
	}
	if err := st.UpsertTool(ctx, &Tool{Code: "", Name: ""}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Expected ErrInvalid, got %v", err)
	}
}

func TestSetActiveMissingEntity(t *testing.T) {
	st, _ := setupTestStore(t)
	ctx := context.Background()

	if err := st.SetToolActive(ctx, 9999, false); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expected ErrNotFound tool, got %v", err)
	}
	if err := st.SetTechnicianActive(ctx, 9999, false); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Expected ErrNotFound technician, got %v", err)
	}
}

func TestMarkDamaged(t *testing.T) {
	st, _ := setupTestStore(t)
	techID, toolID, tool2ID := seedTestEntities(t, st)
	ctx := context.Background()

	if err := st.MarkDamaged(ctx, toolID, "Ditemukan retak"); err != nil {
		t.Fatalf("MarkDamaged error: %v", err)
	}
	var cond, note string
	_ = st.db.QueryRowContext(ctx, "SELECT condition, condition_note FROM tools WHERE id = ?", toolID).Scan(&cond, &note)
	if cond != "rusak" || note != "Ditemukan retak" {
		t.Fatalf("Kondisi alat tidak sesuai: %s / %s", cond, note)
	}

	// Alat yang sedang dipinjam tidak boleh ditandai rusak manual.
	if _, err := st.Borrow(ctx, techID, []int64{tool2ID}); err != nil {
		t.Fatalf("Borrow error: %v", err)
	}
	err := st.MarkDamaged(ctx, tool2ID, "")
	var unavailable *UnavailableError
	if !errors.As(err, &unavailable) || unavailable.Reason != "dipinjam" {
		t.Fatalf("Expected UnavailableError dipinjam, got %v", err)
	}
}

func TestReturnInvalidCondition(t *testing.T) {
	st, _ := setupTestStore(t)
	techID, toolID, _ := seedTestEntities(t, st)
	ctx := context.Background()

	loanIDs, _ := st.Borrow(ctx, techID, []int64{toolID})
	if err := st.Return(ctx, loanIDs[0], Condition("entah"), ""); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Expected ErrInvalid, got %v", err)
	}
}

func TestSettingsValidation(t *testing.T) {
	st, _ := setupTestStore(t)
	ctx := context.Background()

	invalid := []Settings{
		{DefaultMaxHours: 0, ShiftPagiStart: "07:00", ShiftMalamStart: "19:00"},
		{DefaultMaxHours: 24, ShiftPagiStart: "7:00", ShiftMalamStart: "19:00"},
		{DefaultMaxHours: 24, ShiftPagiStart: "07:00", ShiftMalamStart: "07:00"},
	}
	for i, settings := range invalid {
		if err := st.UpdateSettings(ctx, settings); !errors.Is(err, ErrInvalid) {
			t.Fatalf("kasus %d: expected ErrInvalid, got %v", i, err)
		}
	}

	valid := Settings{DefaultMaxHours: 36, ShiftPagiStart: "06:30", ShiftMalamStart: "18:30"}
	if err := st.UpdateSettings(ctx, valid); err != nil {
		t.Fatalf("UpdateSettings valid error: %v", err)
	}
	got, err := st.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings error: %v", err)
	}
	if got.DefaultMaxHours != 36 || got.ShiftPagiStart != "06:30" {
		t.Fatalf("Settings tidak tersimpan: %+v", got)
	}
}

func TestDashboardOverdueUsesDefaultHours(t *testing.T) {
	st, clock := setupTestStore(t)
	techID, toolID, _ := seedTestEntities(t, st)
	ctx := context.Background()

	if _, err := st.Borrow(ctx, techID, []int64{toolID}); err != nil {
		t.Fatalf("Borrow error: %v", err)
	}

	clock.Current = clock.Current.Add(25 * time.Hour)
	vm, err := st.Dashboard(ctx, "")
	if err != nil {
		t.Fatalf("Dashboard error: %v", err)
	}
	if vm.Stats.OverdueCount != 1 {
		t.Fatalf("Expected 1 overdue loan, got %d", vm.Stats.OverdueCount)
	}
	if vm.DefaultMaxHours != 24 {
		t.Fatalf("Expected DefaultMaxHours 24, got %d", vm.DefaultMaxHours)
	}
	if len(vm.ActiveLoans) != 0 {
		t.Fatalf("Loan overdue tidak boleh masuk ActiveLoans")
	}
}

func TestImportCSV(t *testing.T) {
	st, _ := setupTestStore(t)
	ctx := context.Background()

	// Data awal untuk menguji update.
	if err := st.UpsertTool(ctx, &Tool{Code: "MM-01", Name: "Multimeter Lama"}); err != nil {
		t.Fatalf("seed tool error: %v", err)
	}
	if err := st.UpsertTechnician(ctx, &Technician{Name: "Budi", Section: "Listrik"}); err != nil {
		t.Fatalf("seed tech error: %v", err)
	}

	toolsCSV := "\ufeffcode,name,category,max_hours\nMM-01,Multimeter Baru,Alat ukur,48\nJS-01,Jangka Sorong Baru,Alat ukur,\n"
	toolResult, err := st.ImportToolsCSV(ctx, strings.NewReader(toolsCSV))
	if err != nil {
		t.Fatalf("ImportToolsCSV error: %v", err)
	}
	if toolResult.Created != 1 || toolResult.Updated != 1 {
		t.Fatalf("Hasil impor alat tidak sesuai: %+v", toolResult)
	}

	var name, category string
	var maxHours int
	if err := st.db.QueryRowContext(ctx,
		"SELECT name, category, max_hours FROM tools WHERE code = 'MM-01'").Scan(&name, &category, &maxHours); err != nil {
		t.Fatalf("query tool error: %v", err)
	}
	if name != "Multimeter Baru" || category != "Alat ukur" || maxHours != 48 {
		t.Fatalf("Update impor alat gagal: %s / %s / %d", name, category, maxHours)
	}

	techCSV := "\ufeffname,section\nbudi,Listrik Baru\nAndi,Mekanik\n"
	techResult, err := st.ImportTechniciansCSV(ctx, strings.NewReader(techCSV))
	if err != nil {
		t.Fatalf("ImportTechniciansCSV error: %v", err)
	}
	if techResult.Created != 1 || techResult.Updated != 1 {
		t.Fatalf("Hasil impor teknisi tidak sesuai: %+v", techResult)
	}

	var section string
	if err := st.db.QueryRowContext(ctx, "SELECT section FROM technicians WHERE LOWER(name) = 'budi'").Scan(&section); err != nil {
		t.Fatalf("query tech error: %v", err)
	}
	if section != "Listrik Baru" {
		t.Fatalf("Update impor teknisi gagal: %s", section)
	}
}
