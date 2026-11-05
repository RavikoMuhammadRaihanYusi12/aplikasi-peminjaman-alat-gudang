package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"alat-gudang/internal/timefmt"

	_ "modernc.org/sqlite"
)

func setupTestStore(t *testing.T) (*Store, *timefmt.FakeClock) {
	t.Helper()

	// In-memory sqlite with WAL mode
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Open in-memory db error: %v", err)
	}
	// Satu koneksi agar seluruh test memakai database in-memory yang sama.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = db.Close() })

	clock := &timefmt.FakeClock{
		Current: time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC),
	}

	st, err := New(db, clock, "Asia/Jakarta")
	if err != nil {
		t.Fatalf("New store error: %v", err)
	}

	ctx := context.Background()
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("Migrate error: %v", err)
	}

	return st, clock
}

func seedTestEntities(t *testing.T, st *Store) (int64, int64, int64) {
	t.Helper()
	ctx := context.Background()

	tech := &Technician{Name: "Budi", Section: "Listrik"}
	if err := st.UpsertTechnician(ctx, tech); err != nil {
		t.Fatalf("Upsert tech error: %v", err)
	}

	tool1 := &Tool{Code: "MM-01", Name: "Multimeter", Category: "Alat ukur"}
	if err := st.UpsertTool(ctx, tool1); err != nil {
		t.Fatalf("Upsert tool1 error: %v", err)
	}

	tool2 := &Tool{Code: "BR-01", Name: "Bor Listrik", Category: "Alat listrik"}
	if err := st.UpsertTool(ctx, tool2); err != nil {
		t.Fatalf("Upsert tool2 error: %v", err)
	}

	// Query their IDs
	var techID, tool1ID, tool2ID int64
	_ = st.db.QueryRowContext(ctx, "SELECT id FROM technicians WHERE name = 'Budi'").Scan(&techID)
	_ = st.db.QueryRowContext(ctx, "SELECT id FROM tools WHERE code = 'MM-01'").Scan(&tool1ID)
	_ = st.db.QueryRowContext(ctx, "SELECT id FROM tools WHERE code = 'BR-01'").Scan(&tool2ID)

	return techID, tool1ID, tool2ID
}

// Test I1 & I3: Multi-tool atomic borrow and reject duplicate open loan
func TestBorrowAndDuplicateLoan(t *testing.T) {
	st, _ := setupTestStore(t)
	techID, tool1ID, tool2ID := seedTestEntities(t, st)
	ctx := context.Background()

	// 1. Borrow both tools in 1 transaction
	loanIDs, err := st.Borrow(ctx, techID, []int64{tool1ID, tool2ID})
	if err != nil {
		t.Fatalf("Borrow multi-tools error: %v", err)
	}
	if len(loanIDs) != 2 {
		t.Fatalf("Expected 2 loan IDs, got %d", len(loanIDs))
	}

	// 2. Attempt to borrow tool1 again while still open -> MUST fail with UnavailableError
	_, err = st.Borrow(ctx, techID, []int64{tool1ID})
	if err == nil {
		t.Fatalf("Expected error when borrowing already borrowed tool, got nil")
	}
	var unavail *UnavailableError
	if !errors.As(err, &unavail) || unavail.Reason != "dipinjam" {
		t.Fatalf("Expected UnavailableError with reason 'dipinjam', got %v", err)
	}
}

// Test I2: Damaged tool cannot be borrowed
func TestDamagedToolCannotBeBorrowed(t *testing.T) {
	st, _ := setupTestStore(t)
	techID, tool1ID, _ := seedTestEntities(t, st)
	ctx := context.Background()

	// Set tool condition to rusak
	_, _ = st.db.ExecContext(ctx, "UPDATE tools SET condition = 'rusak', condition_note = 'rusak' WHERE id = ?", tool1ID)

	_, err := st.Borrow(ctx, techID, []int64{tool1ID})
	if err == nil {
		t.Fatalf("Expected error when borrowing damaged tool, got nil")
	}
	var unavail *UnavailableError
	if !errors.As(err, &unavail) || unavail.Reason != "rusak" {
		t.Fatalf("Expected UnavailableError with reason 'rusak', got %v", err)
	}
}

// Test I5 & I7: Return Baik / Rusak and Prevent Double Return
func TestReturnAndDoubleReturn(t *testing.T) {
	st, _ := setupTestStore(t)
	techID, tool1ID, _ := seedTestEntities(t, st)
	ctx := context.Background()

	loanIDs, err := st.Borrow(ctx, techID, []int64{tool1ID})
	if err != nil {
		t.Fatalf("Borrow error: %v", err)
	}

	loanID := loanIDs[0]

	// 1. Return as damaged
	err = st.Return(ctx, loanID, Rusak, "Kabel putus")
	if err != nil {
		t.Fatalf("Return error: %v", err)
	}

	// Verify tool is now marked rusak in tools table (I7)
	var cond, note string
	_ = st.db.QueryRowContext(ctx, "SELECT condition, condition_note FROM tools WHERE id = ?", tool1ID).Scan(&cond, &note)
	if cond != string(Rusak) || note != "Kabel putus" {
		t.Fatalf("Expected tool condition 'rusak' and note 'Kabel putus', got %s, %s", cond, note)
	}

	// 2. Second return on same loan must fail with ErrAlreadyReturned (I5)
	err = st.Return(ctx, loanID, Baik, "")
	if !errors.Is(err, ErrAlreadyReturned) {
		t.Fatalf("Expected ErrAlreadyReturned on double click, got %v", err)
	}
}

// Test I6: Void Return (Undo)
func TestVoidReturn(t *testing.T) {
	st, _ := setupTestStore(t)
	techID, tool1ID, _ := seedTestEntities(t, st)
	ctx := context.Background()

	loanIDs, _ := st.Borrow(ctx, techID, []int64{tool1ID})
	loanID := loanIDs[0]

	_ = st.Return(ctx, loanID, Rusak, "Tombol macet")

	// Undo return
	err := st.VoidReturn(ctx, loanID)
	if err != nil {
		t.Fatalf("VoidReturn error: %v", err)
	}

	// Check tool condition is reset back to 'baik'
	var cond string
	_ = st.db.QueryRowContext(ctx, "SELECT condition FROM tools WHERE id = ?", tool1ID).Scan(&cond)
	if cond != string(Baik) {
		t.Fatalf("Expected tool condition reset to 'baik', got %s", cond)
	}
}
