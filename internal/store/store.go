package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"
	"time"

	"alat-gudang/internal/timefmt"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type Store struct {
	db    *sql.DB
	clock timefmt.Clock
	loc   *time.Location
}

func Open(dbPath string) (*sql.DB, error) {
	// Connection string with PRAGMAs recommended in system-design.md
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=synchronous(NORMAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Limit to 1 open connection for clean sequential operations
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite gagal: %w", err)
	}

	return db, nil
}

func New(db *sql.DB, clock timefmt.Clock, tz string) (*Store, error) {
	if clock == nil {
		clock = timefmt.RealClock{}
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("zona waktu %q tidak dikenal: %w", tz, err)
	}

	return &Store{
		db:    db,
		clock: clock,
		loc:   loc,
	}, nil
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) Clock() timefmt.Clock {
	return s.clock
}

func (s *Store) Location() *time.Location {
	return s.loc
}

func (s *Store) QuickCheck(ctx context.Context) error {
	var result string
	row := s.db.QueryRowContext(ctx, "PRAGMA quick_check")
	if err := row.Scan(&result); err != nil {
		return fmt.Errorf("quick_check error: %w", err)
	}
	if result != "ok" {
		return fmt.Errorf("integritas database korup: %s", result)
	}
	return nil
}

// isUniqueViolation mendeteksi pelanggaran constraint UNIQUE dari driver SQLite.
// Dipakai untuk menerjemahkan error database menjadi pesan yang ramah pengguna.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint failed") ||
		strings.Contains(msg, "constraint failed: unique")
}

func (s *Store) Migrate(ctx context.Context) error {
	var userVersion int
	if err := s.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&userVersion); err != nil {
		return fmt.Errorf("baca user_version gagal: %w", err)
	}

	if userVersion < 1 {
		sqlBytes, err := migrationFS.ReadFile("migrations/001_init.sql")
		if err != nil {
			return fmt.Errorf("baca migration 001 gagal: %w", err)
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("eksekusi migration 001 gagal: %w", err)
		}

		if _, err := tx.ExecContext(ctx, "PRAGMA user_version = 1"); err != nil {
			return fmt.Errorf("update user_version gagal: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}
