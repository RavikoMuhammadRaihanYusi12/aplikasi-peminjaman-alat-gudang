package store

import (
	"context"
	"log/slog"
	"strings"

	"alat-gudang/internal/timefmt"
)

func (s *Store) ListTechnicians(ctx context.Context) ([]Technician, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, section, active, created_at FROM technicians ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Technician
	for rows.Next() {
		var (
			id        int64
			name      string
			section   string
			active    int
			createdAt string
		)
		if err := rows.Scan(&id, &name, &section, &active, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, Technician{
			ID:        id,
			Name:      name,
			Section:   section,
			Active:    active == 1,
			CreatedAt: createdAt,
		})
	}
	return result, nil
}

func (s *Store) ActiveTechnicians(ctx context.Context) ([]Technician, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, section, active, created_at FROM technicians WHERE active = 1 ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Technician
	for rows.Next() {
		var (
			id        int64
			name      string
			section   string
			active    int
			createdAt string
		)
		if err := rows.Scan(&id, &name, &section, &active, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, Technician{
			ID:        id,
			Name:      name,
			Section:   section,
			Active:    active == 1,
			CreatedAt: createdAt,
		})
	}
	return result, nil
}

func (s *Store) UpsertTechnician(ctx context.Context, tech *Technician) error {
	if tech == nil {
		return ErrInvalid
	}

	tech.Name = strings.TrimSpace(tech.Name)
	tech.Section = strings.TrimSpace(tech.Section)

	if tech.Name == "" {
		return ErrInvalid
	}

	nowStr := timefmt.FormatUTC(s.clock.Now().UTC())

	if tech.ID == 0 {
		res, err := s.db.ExecContext(ctx, `
			INSERT INTO technicians (name, section, active, created_at)
			VALUES (?, ?, 1, ?)`, tech.Name, tech.Section, nowStr)
		if err != nil {
			return err
		}
		tech.ID, _ = res.LastInsertId()
		slog.Info("Master teknisi dibuat", "id", tech.ID, "name", tech.Name)
		return nil
	}

	res, err := s.db.ExecContext(ctx, `
		UPDATE technicians 
		SET name = ?, section = ?
		WHERE id = ?`, tech.Name, tech.Section, tech.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	slog.Info("Master teknisi diubah", "id", tech.ID, "name", tech.Name)
	return nil
}

func (s *Store) SetTechnicianActive(ctx context.Context, id int64, active bool) error {
	actVal := 0
	if active {
		actVal = 1
	}
	res, err := s.db.ExecContext(ctx, "UPDATE technicians SET active = ? WHERE id = ?", actVal, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	slog.Info("Status aktif teknisi diubah", "id", id, "active", active)
	return nil
}
