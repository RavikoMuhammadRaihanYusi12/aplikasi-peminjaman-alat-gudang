package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"alat-gudang/internal/timefmt"
)

func (s *Store) ListTools(ctx context.Context, filter string) ([]ToolRowVM, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.code, t.name, t.category, t.max_hours, t.condition, t.condition_note, t.active,
		       COALESCE(p_curr.name, '') AS current_borrower,
		       COALESCE(p_last.name, '') AS last_borrower
		FROM tools t
		LEFT JOIN loans l_curr ON l_curr.tool_id = t.id AND l_curr.returned_at IS NULL AND l_curr.voided_at IS NULL
		LEFT JOIN technicians p_curr ON p_curr.id = l_curr.technician_id
		LEFT JOIN loans l_last ON l_last.id = (
			SELECT id FROM loans
			WHERE tool_id = t.id AND voided_at IS NULL
			ORDER BY borrowed_at DESC, id DESC LIMIT 1)
		LEFT JOIN technicians p_last ON p_last.id = l_last.technician_id
		WHERE t.active = 1
		ORDER BY t.code ASC`)
	if err != nil {
		return nil, fmt.Errorf("list tools error: %w", err)
	}
	defer rows.Close()

	var result []ToolRowVM
	filter = strings.ToLower(strings.TrimSpace(filter))

	for rows.Next() {
		var (
			id              int64
			code            string
			name            string
			category        string
			maxHours        sql.NullInt64
			condition       string
			conditionNote   string
			active          int
			currentBorrower string
			lastBorrower    string
		)
		if err := rows.Scan(&id, &code, &name, &category, &maxHours, &condition, &conditionNote, &active, &currentBorrower, &lastBorrower); err != nil {
			return nil, err
		}

		status := "tersedia"
		if condition == "rusak" {
			status = "rusak"
		} else if currentBorrower != "" {
			status = "dipinjam"
		}

		if filter != "" && filter != "all" {
			if filter == "tersedia" && status != "tersedia" {
				continue
			}
			if filter == "dipinjam" && status != "dipinjam" {
				continue
			}
			if filter == "rusak" && status != "rusak" {
				continue
			}
		}

		var mHours *int
		if maxHours.Valid && maxHours.Int64 > 0 {
			val := int(maxHours.Int64)
			mHours = &val
		}

		result = append(result, ToolRowVM{
			ID:               id,
			Code:             code,
			Name:             name,
			Category:         category,
			MaxHours:         mHours,
			Condition:        Condition(condition),
			ConditionNote:    conditionNote,
			Active:           active == 1,
			Status:           status,
			CurrentBorrower:  currentBorrower,
			LastBorrowerName: lastBorrower,
		})
	}

	return result, nil
}

// ListAllTools mengembalikan semua alat termasuk yang nonaktif, dipakai di halaman Kelola.
func (s *Store) ListAllTools(ctx context.Context) ([]ToolRowVM, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.code, t.name, t.category, t.max_hours, t.condition, t.condition_note, t.active,
		       COALESCE(p_curr.name, '') AS current_borrower,
		       COALESCE(p_last.name, '') AS last_borrower
		FROM tools t
		LEFT JOIN loans l_curr ON l_curr.tool_id = t.id AND l_curr.returned_at IS NULL AND l_curr.voided_at IS NULL
		LEFT JOIN technicians p_curr ON p_curr.id = l_curr.technician_id
		LEFT JOIN loans l_last ON l_last.id = (
			SELECT id FROM loans
			WHERE tool_id = t.id AND voided_at IS NULL
			ORDER BY borrowed_at DESC, id DESC LIMIT 1)
		LEFT JOIN technicians p_last ON p_last.id = l_last.technician_id
		ORDER BY t.active DESC, t.code ASC`)
	if err != nil {
		return nil, fmt.Errorf("list all tools error: %w", err)
	}
	defer rows.Close()

	var result []ToolRowVM
	for rows.Next() {
		var (
			id              int64
			code            string
			name            string
			category        string
			maxHours        sql.NullInt64
			condition       string
			conditionNote   string
			active          int
			currentBorrower string
			lastBorrower    string
		)
		if err := rows.Scan(&id, &code, &name, &category, &maxHours, &condition, &conditionNote, &active, &currentBorrower, &lastBorrower); err != nil {
			return nil, err
		}

		status := "tersedia"
		if condition == "rusak" {
			status = "rusak"
		} else if currentBorrower != "" {
			status = "dipinjam"
		}

		var mHours *int
		if maxHours.Valid && maxHours.Int64 > 0 {
			val := int(maxHours.Int64)
			mHours = &val
		}

		result = append(result, ToolRowVM{
			ID:               id,
			Code:             code,
			Name:             name,
			Category:         category,
			MaxHours:         mHours,
			Condition:        Condition(condition),
			ConditionNote:    conditionNote,
			Active:           active == 1,
			Status:           status,
			CurrentBorrower:  currentBorrower,
			LastBorrowerName: lastBorrower,
		})
	}
	return result, nil
}

func (s *Store) LoanableTools(ctx context.Context) ([]ToolRowVM, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.code, t.name, t.category, t.max_hours, t.condition, t.condition_note, t.active,
		       COALESCE(p.name, '') AS current_borrower
		FROM tools t
		LEFT JOIN loans l ON l.tool_id = t.id AND l.returned_at IS NULL AND l.voided_at IS NULL
		LEFT JOIN technicians p ON p.id = l.technician_id
		WHERE t.active = 1
		ORDER BY t.code ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ToolRowVM
	for rows.Next() {
		var (
			id              int64
			code            string
			name            string
			category        string
			maxHours        sql.NullInt64
			condition       string
			conditionNote   string
			active          int
			currentBorrower string
		)
		if err := rows.Scan(&id, &code, &name, &category, &maxHours, &condition, &conditionNote, &active, &currentBorrower); err != nil {
			return nil, err
		}

		status := "tersedia"
		if condition == "rusak" {
			status = "rusak"
		} else if currentBorrower != "" {
			status = "dipinjam"
		}

		var mHours *int
		if maxHours.Valid && maxHours.Int64 > 0 {
			val := int(maxHours.Int64)
			mHours = &val
		}

		result = append(result, ToolRowVM{
			ID:              id,
			Code:            code,
			Name:            name,
			Category:        category,
			MaxHours:        mHours,
			Condition:       Condition(condition),
			ConditionNote:   conditionNote,
			Active:          active == 1,
			Status:          status,
			CurrentBorrower: currentBorrower,
		})
	}

	return result, nil
}

func (s *Store) ToolDetail(ctx context.Context, id int64) (ToolDetailVM, error) {
	var (
		code          string
		name          string
		category      string
		maxHours      sql.NullInt64
		condition     string
		conditionNote string
		active        int
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT code, name, category, max_hours, condition, condition_note, active
		FROM tools WHERE id = ?`, id).Scan(&code, &name, &category, &maxHours, &condition, &conditionNote, &active)
	if err == sql.ErrNoRows {
		return ToolDetailVM{}, ErrNotFound
	}
	if err != nil {
		return ToolDetailVM{}, err
	}

	// Cek open loan
	var currentBorrower string
	err = s.db.QueryRowContext(ctx, `
		SELECT p.name 
		FROM loans l
		JOIN technicians p ON p.id = l.technician_id
		WHERE l.tool_id = ? AND l.returned_at IS NULL AND l.voided_at IS NULL LIMIT 1`, id).Scan(&currentBorrower)
	if err != nil && err != sql.ErrNoRows {
		return ToolDetailVM{}, err
	}

	status := "tersedia"
	if condition == "rusak" {
		status = "rusak"
	} else if currentBorrower != "" {
		status = "dipinjam"
	}

	var mHours *int
	if maxHours.Valid && maxHours.Int64 > 0 {
		val := int(maxHours.Int64)
		mHours = &val
	}

	toolRow := ToolRowVM{
		ID:              id,
		Code:            code,
		Name:            name,
		Category:        category,
		MaxHours:        mHours,
		Condition:       Condition(condition),
		ConditionNote:   conditionNote,
		Active:          active == 1,
		Status:          status,
		CurrentBorrower: currentBorrower,
	}

	// Ambil riwayat peminjaman
	rows, err := s.db.QueryContext(ctx, `
		SELECT l.id, p.name, p.section, l.borrowed_at, l.borrow_shift,
		       l.returned_at, l.return_shift, l.return_condition, l.return_note
		FROM loans l
		JOIN technicians p ON p.id = l.technician_id
		WHERE l.tool_id = ? AND l.voided_at IS NULL
		ORDER BY l.borrowed_at DESC, l.id DESC
		LIMIT 50`, id)
	if err != nil {
		return ToolDetailVM{}, err
	}
	defer rows.Close()

	var history []ToolHistoryItemVM
	for rows.Next() {
		var (
			lID        int64
			pName      string
			pSection   string
			bAt        string
			bShift     string
			rAt        sql.NullString
			rShift     sql.NullString
			rCondition sql.NullString
			rNote      string
		)
		if err := rows.Scan(&lID, &pName, &pSection, &bAt, &bShift, &rAt, &rShift, &rCondition, &rNote); err != nil {
			return ToolDetailVM{}, err
		}

		bTime, _ := timefmt.ParseUTC(bAt)
		rTimeFormatted := ""
		if rAt.Valid && rAt.String != "" {
			if rt, err := timefmt.ParseUTC(rAt.String); err == nil {
				rTimeFormatted = timefmt.FormatDateTime(rt, s.loc)
			}
		}

		history = append(history, ToolHistoryItemVM{
			ID:                  lID,
			TechnicianName:      pName,
			TechnicianSection:   pSection,
			BorrowedAtFormatted: timefmt.FormatDateTime(bTime, s.loc),
			BorrowShift:         bShift,
			ReturnedAtFormatted: rTimeFormatted,
			ReturnShift:         rShift.String,
			ReturnCondition:     rCondition.String,
			ReturnNote:          rNote,
		})
	}

	var lastLoan *ToolHistoryItemVM
	if len(history) > 0 {
		lastLoan = &history[0]
		toolRow.LastBorrowerName = history[0].TechnicianName
	}

	return ToolDetailVM{
		ActiveNav: "alat",
		Tool:      toolRow,
		LastLoan:  lastLoan,
		History:   history,
	}, nil
}

func (s *Store) UpsertTool(ctx context.Context, tool *Tool) error {
	if tool == nil {
		return ErrInvalid
	}

	tool.Code = strings.ToUpper(strings.TrimSpace(tool.Code))
	tool.Name = strings.TrimSpace(tool.Name)
	tool.Category = strings.TrimSpace(tool.Category)

	if tool.Code == "" || tool.Name == "" {
		return ErrInvalid
	}
	if tool.MaxHours != nil && (*tool.MaxHours <= 0 || *tool.MaxHours > 8760) {
		return ErrInvalid
	}

	nowStr := timefmt.FormatUTC(s.clock.Now().UTC())

	if tool.ID == 0 {
		res, err := s.db.ExecContext(ctx, `
			INSERT INTO tools (code, name, category, max_hours, condition, condition_note, active, created_at)
			VALUES (?, ?, ?, ?, 'baik', '', 1, ?)`,
			tool.Code, tool.Name, tool.Category, tool.MaxHours, nowStr)
		if isUniqueViolation(err) {
			return ErrDuplicateCode
		}
		if err != nil {
			return err
		}
		tool.ID, _ = res.LastInsertId()
		slog.Info("Master alat dibuat", "id", tool.ID, "code", tool.Code, "name", tool.Name)
		return nil
	}

	res, err := s.db.ExecContext(ctx, `
		UPDATE tools 
		SET code = ?, name = ?, category = ?, max_hours = ?
		WHERE id = ?`,
		tool.Code, tool.Name, tool.Category, tool.MaxHours, tool.ID)
	if isUniqueViolation(err) {
		return ErrDuplicateCode
	}
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	slog.Info("Master alat diubah", "id", tool.ID, "code", tool.Code, "name", tool.Name)
	return nil
}

func (s *Store) SetToolActive(ctx context.Context, id int64, active bool) error {
	actVal := 0
	if active {
		actVal = 1
	}
	res, err := s.db.ExecContext(ctx, "UPDATE tools SET active = ? WHERE id = ?", actVal, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	slog.Info("Status aktif alat diubah", "id", id, "active", active)
	return nil
}

func (s *Store) MarkRepaired(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, "UPDATE tools SET condition = 'baik', condition_note = '' WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	slog.Info("Alat ditandai sudah diperbaiki", "id", id)
	return nil
}

// MarkDamaged menandai alat rusak secara manual (P1): alat ditemukan rusak
// setelah dikembalikan. Alat yang sedang dipinjam tidak boleh ditandai rusak
// karena pemegangnya harus tercatat lewat alur pengembalian.
func (s *Store) MarkDamaged(ctx context.Context, toolID int64, note string) error {
	if toolID <= 0 {
		return ErrInvalid
	}
	note = strings.TrimSpace(note)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var code string
	if err := tx.QueryRowContext(ctx, "SELECT code FROM tools WHERE id = ?", toolID).Scan(&code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	var holder string
	err = tx.QueryRowContext(ctx, `
		SELECT p.name
		FROM loans l
		JOIN technicians p ON p.id = l.technician_id
		WHERE l.tool_id = ? AND l.returned_at IS NULL AND l.voided_at IS NULL
		LIMIT 1`, toolID).Scan(&holder)
	if err == nil {
		return &UnavailableError{ToolID: toolID, Code: code, Reason: "dipinjam", HolderName: holder}
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE tools SET condition = 'rusak', condition_note = ? WHERE id = ?`,
		note, toolID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	slog.Info("Alat ditandai rusak", "id", toolID, "code", code, "note", note)
	return nil
}
