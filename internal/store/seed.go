package store

import (
	"context"
	"time"

	"alat-gudang/internal/timefmt"
)

func (s *Store) SeedSampleData(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM technicians").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // Sudah ada data
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := s.clock.Now().UTC()
	nowStr := timefmt.FormatUTC(now)

	// 1. Seed Technicians
	techs := []struct {
		name, section string
	}{
		{"Budi", "Listrik"},
		{"Andi", "Mekanik"},
		{"Candra", "Listrik"},
		{"Dedi", "Mekanik"},
		{"Eko", "Utilitas"},
		{"Fajar", "Produksi"},
		{"Gunawan", "Listrik"},
		{"Hendra", "Mekanik"},
		{"Irwan", "Utilitas"},
		{"Joko", "Mekanik"},
		{"Krisna", "Listrik"},
		{"Lukman", "Produksi"},
	}

	techIDs := make(map[string]int64)
	for _, t := range techs {
		res, err := tx.ExecContext(ctx, "INSERT INTO technicians (name, section, active, created_at) VALUES (?, ?, 1, ?)", t.name, t.section, nowStr)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		techIDs[t.name] = id
	}

	// 2. Seed Tools
	tools := []struct {
		code, name, category, cond, note string
		maxHours                         int
	}{
		{"MM-01", "Multimeter Digital Fluke", "Alat ukur", "baik", "", 24},
		{"MM-02", "Multimeter Sanwa", "Alat ukur", "baik", "", 24},
		{"MM-03", "Multimeter Kyoritsu", "Alat ukur", "baik", "", 24},
		{"BR-01", "Bor Listrik Bosch", "Alat listrik", "baik", "", 24},
		{"BR-02", "Bor Baterai Makita", "Alat listrik", "baik", "", 24},
		{"JS-01", "Jangka Sorong 150mm", "Alat ukur", "baik", "", 24},
		{"JS-02", "Jangka Sorong Digital Mitutoyo", "Alat ukur", "baik", "", 24},
		{"JS-03", "Jangka Sorong 300mm", "Alat ukur", "baik", "", 24},
		{"OL-01", "Obeng Listrik Presisi", "Alat listrik", "baik", "", 24},
		{"OL-02", "Obeng Torsi Listrik", "Alat listrik", "rusak", "Motor tidak menyala & bau sangit", 24},
		{"KM-01", "Kunci Momen Tohnichi", "Kunci khusus", "baik", "", 24},
		{"KS-01", "Kunci Sok Set 24 pcs", "Kunci khusus", "baik", "", 24},
	}

	toolIDs := make(map[string]int64)
	for _, t := range tools {
		res, err := tx.ExecContext(ctx, `
			INSERT INTO tools (code, name, category, max_hours, condition, condition_note, active, created_at)
			VALUES (?, ?, ?, ?, ?, ?, 1, ?)`,
			t.code, t.name, t.category, t.maxHours, t.cond, t.note, nowStr)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		toolIDs[t.code] = id
	}

	// 3. Seed Sample Loans:
	// - MM-02 dipinjam Budi sejak 2 hari lalu (melewati batas -> PERLU DICEK)
	twoDaysAgo := now.Add(-48 * time.Hour)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO loans (tool_id, technician_id, borrowed_at, borrow_shift)
		VALUES (?, ?, ?, 'pagi')`, toolIDs["MM-02"], techIDs["Budi"], timefmt.FormatUTC(twoDaysAgo))
	if err != nil {
		return err
	}

	// - BR-01 dipinjam Andi sejak 28 jam lalu (melewati batas -> PERLU DICEK)
	oneDayAgo := now.Add(-28 * time.Hour)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO loans (tool_id, technician_id, borrowed_at, borrow_shift)
		VALUES (?, ?, ?, 'malam')`, toolIDs["BR-01"], techIDs["Andi"], timefmt.FormatUTC(oneDayAgo))
	if err != nil {
		return err
	}

	// - JS-02 dipinjam Dedi sejak 2 jam lalu (normal)
	twoHoursAgo := now.Add(-2 * time.Hour)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO loans (tool_id, technician_id, borrowed_at, borrow_shift)
		VALUES (?, ?, ?, 'pagi')`, toolIDs["JS-02"], techIDs["Dedi"], timefmt.FormatUTC(twoHoursAgo))
	if err != nil {
		return err
	}

	// - OL-02 dikembalikan rusak oleh Hendra kemarin (riwayat peminjam terakhir untuk OL-02)
	yesterday := now.Add(-24 * time.Hour)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO loans (tool_id, technician_id, borrowed_at, borrow_shift, returned_at, return_shift, return_condition, return_note)
		VALUES (?, ?, ?, 'pagi', ?, 'malam', 'rusak', 'Motor tidak menyala & bau sangit')`,
		toolIDs["OL-02"], techIDs["Hendra"], timefmt.FormatUTC(twoDaysAgo), timefmt.FormatUTC(yesterday))
	if err != nil {
		return err
	}

	return tx.Commit()
}
