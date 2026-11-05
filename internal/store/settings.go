package store

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

func (s *Store) GetSettings(ctx context.Context) (Settings, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT key, value FROM settings")
	if err != nil {
		return Settings{DefaultMaxHours: 24, ShiftPagiStart: "07:00", ShiftMalamStart: "19:00"}, err
	}
	defer rows.Close()

	settings := Settings{
		DefaultMaxHours: 24,
		ShiftPagiStart:  "07:00",
		ShiftMalamStart: "19:00",
	}

	for rows.Next() {
		var key, val string
		if err := rows.Scan(&key, &val); err != nil {
			return settings, err
		}
		switch key {
		case "default_max_hours":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				settings.DefaultMaxHours = n
			}
		case "shift_pagi_start":
			if val != "" {
				settings.ShiftPagiStart = val
			}
		case "shift_malam_start":
			if val != "" {
				settings.ShiftMalamStart = val
			}
		}
	}

	return settings, nil
}

func (s *Store) UpdateSettings(ctx context.Context, settings Settings) error {
	settings.ShiftPagiStart = strings.TrimSpace(settings.ShiftPagiStart)
	settings.ShiftMalamStart = strings.TrimSpace(settings.ShiftMalamStart)

	if settings.DefaultMaxHours < 1 || settings.DefaultMaxHours > 8760 {
		return ErrInvalid
	}
	if !validClock(settings.ShiftPagiStart) || !validClock(settings.ShiftMalamStart) {
		return ErrInvalid
	}
	if settings.ShiftPagiStart == settings.ShiftMalamStart {
		return ErrInvalid
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	hoursStr := strconv.Itoa(settings.DefaultMaxHours)

	queries := []struct {
		key, val string
	}{
		{"default_max_hours", hoursStr},
		{"shift_pagi_start", settings.ShiftPagiStart},
		{"shift_malam_start", settings.ShiftMalamStart},
	}

	for _, q := range queries {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO settings (key, value) VALUES (?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value`, q.key, q.val)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	slog.Info("Pengaturan sistem diubah",
		"default_max_hours", settings.DefaultMaxHours,
		"shift_pagi_start", settings.ShiftPagiStart,
		"shift_malam_start", settings.ShiftMalamStart)
	return nil
}

func validClock(value string) bool {
	if len(value) != 5 {
		return false
	}
	_, err := time.Parse("15:04", value)
	return err == nil
}
