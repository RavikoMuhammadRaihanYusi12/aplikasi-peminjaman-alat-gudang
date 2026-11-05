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
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type Store struct {
	db      *sql.DB
	clock   timefmt.Clock
	loc     *time.Location
	remote  bool
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

// OpenRemote membuka database Turso/libSQL lewat jaringan. Dipakai saat
// TURSO_DATABASE_URL diisi (mis. saat deploy di Render).
func OpenRemote(url, token string) (*sql.DB, error) {
	dsn := url
	if token != "" {
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		dsn += sep + "authToken=" + token
	}

	db, err := sql.Open("libsql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping turso gagal: %w", err)
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

// SetRemote menandai bahwa database berada di Turso (bukan file lokal).
func (s *Store) SetRemote(remote bool) {
	s.remote = remote
}

func (s *Store) IsRemote() bool {
	return s.remote
}

func (s *Store) Clock() timefmt.Clock {
	return s.clock
}

func (s *Store) Location() *time.Location {
	return s.loc
}

func (s *Store) QuickCheck(ctx context.Context) error {
	if s.remote {
		// Turso (libSQL) tidak selalu mengizinkan PRAGMA quick_check via HTTP.
		var one int
		if err := s.db.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil {
			return fmt.Errorf("koneksi turso gagal: %w", err)
		}
		return nil
	}

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
	// Tabel penanda migrasi portabel (SQLite lokal maupun Turso/libSQL).
	if _, err := s.db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY, applied_at TEXT NOT NULL)"); err != nil {
		return fmt.Errorf("buat tabel schema_migrations gagal: %w", err)
	}

	var applied int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE name = ?", "001_init").Scan(&applied); err != nil {
		return fmt.Errorf("baca schema_migrations gagal: %w", err)
	}
	if applied > 0 {
		return nil
	}

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

	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (name, applied_at) VALUES (?, ?)", "001_init", time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("catat migration 001 gagal: %w", err)
	}

	return tx.Commit()
}
