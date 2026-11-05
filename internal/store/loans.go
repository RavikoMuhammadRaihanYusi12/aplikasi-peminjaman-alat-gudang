package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"alat-gudang/internal/timefmt"
)

func (s *Store) Borrow(ctx context.Context, techID int64, toolIDs []int64) ([]int64, error) {
	if techID <= 0 || len(toolIDs) == 0 {
		return nil, ErrInvalid
	}

	// Hilangkan duplikat tanpa mengubah urutan. Kalau tool_id yang sama dikirim
	// dua kali, alat tetap hanya dipinjamkan satu kali.
	toolIDs = uniqueIDs(toolIDs)

	now := s.clock.Now().UTC()
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	shift := timefmt.ShiftAt(now, settings.ShiftPagiStart, settings.ShiftMalamStart, s.loc)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Cek teknisi aktif
	var techActive int
	err = tx.QueryRowContext(ctx, "SELECT active FROM technicians WHERE id = ?", techID).Scan(&techActive)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if techActive != 1 {
		return nil, &UnavailableError{Reason: "nonaktif"}
	}

	// 2. Validasi kelayakan tiap alat & insert loan
	nowStr := timefmt.FormatUTC(now)
	loanIDs := make([]int64, 0, len(toolIDs))

	for _, toolID := range toolIDs {
		var (
			code      string
			condition string
			active    int
		)
		err := tx.QueryRowContext(ctx, "SELECT code, condition, active FROM tools WHERE id = ?", toolID).Scan(&code, &condition, &active)
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, err
		}

		if active != 1 {
			return nil, &UnavailableError{ToolID: toolID, Code: code, Reason: "nonaktif"}
		}
		if condition == "rusak" {
			return nil, &UnavailableError{ToolID: toolID, Code: code, Reason: "rusak"}
		}

		// Cek apakah sedang ada loan terbuka
		var holderName string
		err = tx.QueryRowContext(ctx, `
			SELECT p.name 
			FROM loans l
			JOIN technicians p ON p.id = l.technician_id
			WHERE l.tool_id = ? AND l.returned_at IS NULL AND l.voided_at IS NULL
			LIMIT 1`, toolID).Scan(&holderName)
		if err == nil {
			return nil, &UnavailableError{ToolID: toolID, Code: code, Reason: "dipinjam", HolderName: holderName}
		} else if err != sql.ErrNoRows {
			return nil, err
		}

		// Insert loan baru
		res, err := tx.ExecContext(ctx, `
			INSERT INTO loans (tool_id, technician_id, borrowed_at, borrow_shift)
			VALUES (?, ?, ?, ?)`, toolID, techID, nowStr, shift)
		if err != nil {
			// Jaring pengaman terakhir: partial unique index menolak loan kedua.
			if isUniqueViolation(err) {
				return nil, &UnavailableError{ToolID: toolID, Code: code, Reason: "dipinjam", HolderName: holderName}
			}
			return nil, err
		}

		id, _ := res.LastInsertId()
		loanIDs = append(loanIDs, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return loanIDs, nil
}

func (s *Store) Return(ctx context.Context, loanID int64, c Condition, note string) error {
	if c != Baik && c != Rusak {
		return ErrInvalid
	}

	now := s.clock.Now().UTC()
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return err
	}
	shift := timefmt.ShiftAt(now, settings.ShiftPagiStart, settings.ShiftMalamStart, s.loc)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Compare-and-set: hanya berhasil jika loan masih terbuka
	res, err := tx.ExecContext(ctx, `
		UPDATE loans 
		SET returned_at = ?, return_shift = ?, return_condition = ?, return_note = ?
		WHERE id = ? AND returned_at IS NULL AND voided_at IS NULL`,
		timefmt.FormatUTC(now), shift, string(c), note, loanID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return ErrAlreadyReturned
	}

	// Jika rusak, perbarui status master alat
	if c == Rusak {
		_, err = tx.ExecContext(ctx, `
			UPDATE tools 
			SET condition = 'rusak', condition_note = ?
			WHERE id = (SELECT tool_id FROM loans WHERE id = ?)`, note, loanID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// uniqueIDs membuang ID duplikat sambil mempertahankan urutan kemunculan.
func uniqueIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// LoanForReturn mengambil data yang dibutuhkan dialog pengembalian.
// Hanya loan yang masih terbuka dan tidak dibatalkan yang boleh dikembalikan.
func (s *Store) LoanForReturn(ctx context.Context, loanID int64) (ReturnDialogVM, error) {
	var (
		vm         ReturnDialogVM
		returnedAt sql.NullString
		borrowedAt string
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT l.id, t.id, t.code, t.name, p.name, l.borrowed_at, l.returned_at
		FROM loans l
		JOIN tools t ON t.id = l.tool_id
		JOIN technicians p ON p.id = l.technician_id
		WHERE l.id = ? AND l.voided_at IS NULL`, loanID).Scan(
		&vm.LoanID, &vm.ToolID, &vm.ToolCode, &vm.ToolName,
		&vm.TechnicianName, &borrowedAt, &returnedAt,
	)
	if err == sql.ErrNoRows {
		return ReturnDialogVM{}, ErrNotFound
	}
	if err != nil {
		return ReturnDialogVM{}, err
	}
	if returnedAt.Valid {
		return ReturnDialogVM{}, ErrAlreadyReturned
	}

	vm.BorrowedAtFormatted = borrowedAt
	if t, err := timefmt.ParseUTC(borrowedAt); err == nil {
		vm.BorrowedAtFormatted = timefmt.FormatClock(t, s.loc)
	}
	return vm, nil
}

func (s *Store) VoidBorrow(ctx context.Context, loanIDs []int64) error {
	loanIDs = uniqueIDs(loanIDs)
	if len(loanIDs) == 0 {
		return ErrInvalid
	}

	now := s.clock.Now().UTC()
	nowStr := timefmt.FormatUTC(now)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, loanID := range loanIDs {
		// Pastikan belum pernah dikembalikan
		var returnedAt sql.NullString
		err := tx.QueryRowContext(ctx, "SELECT returned_at FROM loans WHERE id = ? AND voided_at IS NULL", loanID).Scan(&returnedAt)
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if returnedAt.Valid {
			return ErrCannotUndo
		}

		_, err = tx.ExecContext(ctx, "UPDATE loans SET voided_at = ? WHERE id = ?", nowStr, loanID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) VoidReturn(ctx context.Context, loanID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		toolID          int64
		returnCondition sql.NullString
		returnedAt      sql.NullString
	)

	err = tx.QueryRowContext(ctx, `
		SELECT tool_id, return_condition, returned_at 
		FROM loans 
		WHERE id = ? AND voided_at IS NULL`, loanID).Scan(&toolID, &returnCondition, &returnedAt)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if !returnedAt.Valid {
		return ErrInvalid
	}

	// Cek apakah loan ini adalah loan non-void terakhir alat itu
	var latestLoanID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM loans 
		WHERE tool_id = ? AND voided_at IS NULL 
		ORDER BY borrowed_at DESC, id DESC LIMIT 1`, toolID).Scan(&latestLoanID)
	if err != nil {
		return err
	}
	if latestLoanID != loanID {
		return ErrCannotUndo
	}

	// Batalkan pengembalian: kosongkan returned_at, return_shift, return_condition, return_note
	_, err = tx.ExecContext(ctx, `
		UPDATE loans 
		SET returned_at = NULL, return_shift = NULL, return_condition = NULL, return_note = ''
		WHERE id = ?`, loanID)
	if err != nil {
		return err
	}

	// Jika saat kembali ditandai rusak, kembalikan kondisi alat ke baik
	if returnCondition.Valid && returnCondition.String == string(Rusak) {
		_, err = tx.ExecContext(ctx, "UPDATE tools SET condition = 'baik', condition_note = '' WHERE id = ?", toolID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) Dashboard(ctx context.Context, q string) (DashboardVM, error) {
	now := s.clock.Now()
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return DashboardVM{}, err
	}

	defaultHours := settings.DefaultMaxHours
	if defaultHours <= 0 {
		defaultHours = 24
	}

	// 1. Ambil semua loan yang sedang terbuka
	rows, err := s.db.QueryContext(ctx, `
		SELECT l.id, t.id, t.code, t.name, t.max_hours,
		       p.id, p.name, p.section, l.borrowed_at, l.borrow_shift
		FROM loans l
		JOIN tools t ON t.id = l.tool_id
		JOIN technicians p ON p.id = l.technician_id
		WHERE l.returned_at IS NULL AND l.voided_at IS NULL
		ORDER BY l.borrowed_at ASC, l.id ASC`)
	if err != nil {
		return DashboardVM{}, fmt.Errorf("query open loans gagal: %w", err)
	}
	defer rows.Close()

	var overdueItems []DashboardItemVM
	var activeItems []DashboardItemVM
	totalBorrowed := 0

	q = strings.ToLower(strings.TrimSpace(q))

	for rows.Next() {
		var (
			loanID      int64
			toolID      int64
			toolCode    string
			toolName    string
			maxHours    sql.NullInt64
			techID      int64
			techName    string
			techSection string
			borrowedAt  string
			borrowShift string
		)
		if err := rows.Scan(&loanID, &toolID, &toolCode, &toolName, &maxHours, &techID, &techName, &techSection, &borrowedAt, &borrowShift); err != nil {
			return DashboardVM{}, err
		}

		totalBorrowed++

		borrowedTime, _ := timefmt.ParseUTC(borrowedAt)
		limitHours := defaultHours
		if maxHours.Valid && maxHours.Int64 > 0 {
			limitHours = int(maxHours.Int64)
		}

		hoursElapsed := now.Sub(borrowedTime).Hours()
		isOverdue := hoursElapsed > float64(limitHours)

		// Filter pencarian
		if q != "" {
			match := strings.Contains(strings.ToLower(toolName), q) ||
				strings.Contains(strings.ToLower(toolCode), q) ||
				strings.Contains(strings.ToLower(techName), q) ||
				strings.Contains(strings.ToLower(techSection), q)
			if !match {
				continue
			}
		}

		item := DashboardItemVM{
			LoanID:              loanID,
			ToolID:              toolID,
			ToolCode:            toolCode,
			ToolName:            toolName,
			TechnicianID:        techID,
			TechnicianName:      techName,
			TechnicianSection:   techSection,
			BorrowedAt:          borrowedAt,
			BorrowShift:         borrowShift,
			BorrowedAtFormatted: timefmt.FormatClock(borrowedTime, s.loc),
			DurationRelative:    timefmt.RelativeID(borrowedTime, now),
			IsOverdue:           isOverdue,
		}

		if isOverdue {
			overdueItems = append(overdueItems, item)
		} else {
			activeItems = append(activeItems, item)
		}
	}

	// 2. Ambil alat rusak beserta peminjam terakhir
	damagedRows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.code, t.name, t.condition_note,
		       COALESCE(p.name, ''), COALESCE(l.returned_at, '')
		FROM tools t
		LEFT JOIN loans l ON l.id = (
			SELECT id FROM loans
			WHERE tool_id = t.id AND voided_at IS NULL
			ORDER BY borrowed_at DESC, id DESC LIMIT 1)
		LEFT JOIN technicians p ON p.id = l.technician_id
		WHERE t.condition = 'rusak' AND t.active = 1
		ORDER BY t.code ASC`)
	if err != nil {
		return DashboardVM{}, fmt.Errorf("query damaged tools gagal: %w", err)
	}
	defer damagedRows.Close()

	var damagedItems []DamagedToolVM
	totalDamaged := 0

	for damagedRows.Next() {
		var (
			toolID       int64
			code         string
			name         string
			note         string
			lastBorrower string
			returnedAt   string
		)
		if err := damagedRows.Scan(&toolID, &code, &name, &note, &lastBorrower, &returnedAt); err != nil {
			return DashboardVM{}, err
		}
		totalDamaged++

		if q != "" {
			match := strings.Contains(strings.ToLower(name), q) ||
				strings.Contains(strings.ToLower(code), q) ||
				strings.Contains(strings.ToLower(lastBorrower), q)
			if !match {
				continue
			}
		}

		formattedReturn := ""
		if returnedAt != "" {
			if rt, err := timefmt.ParseUTC(returnedAt); err == nil {
				formattedReturn = timefmt.FormatClock(rt, s.loc)
			}
		}

		damagedItems = append(damagedItems, DamagedToolVM{
			ID:                  toolID,
			Code:                code,
			Name:                name,
			ConditionNote:       note,
			LastBorrowerName:    lastBorrower,
			ReturnedAt:          returnedAt,
			ReturnedAtFormatted: formattedReturn,
		})
	}

	return DashboardVM{
		ActiveNav: "beranda",
		Stats: DashboardStatsVM{
			BorrowedCount: totalBorrowed,
			OverdueCount:  len(overdueItems),
			DamagedCount:  totalDamaged,
		},
		OverdueLoans:    overdueItems,
		ActiveLoans:     activeItems,
		DamagedTools:    damagedItems,
		Query:           q,
		DefaultMaxHours: defaultHours,
	}, nil
}
