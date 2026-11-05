package backup

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestPerformBackupAndRetention(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "source.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db error: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec("CREATE TABLE contoh (id INTEGER PRIMARY KEY, nama TEXT)"); err != nil {
		t.Fatalf("create table error: %v", err)
	}
	if _, err := db.Exec("INSERT INTO contoh (nama) VALUES ('multimeter')"); err != nil {
		t.Fatalf("insert error: %v", err)
	}

	backupDir := filepath.Join(dir, "backup")
	manager := NewManager(db, backupDir, 2)

	for i := 0; i < 3; i++ {
		path, err := manager.PerformBackup(context.Background())
		if err != nil {
			t.Fatalf("PerformBackup %d error: %v", i+1, err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("file backup tidak ada: %v", err)
		}
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("read backup dir error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("retensi harus menyisakan 2 backup, got %d", len(entries))
	}

	// Pastikan isi backup benar-benar bisa dibuka dan berisi data.
	backupPath := filepath.Join(backupDir, entries[len(entries)-1].Name())
	copyDB, err := sql.Open("sqlite", backupPath)
	if err != nil {
		t.Fatalf("open backup error: %v", err)
	}
	defer copyDB.Close()

	var count int
	if err := copyDB.QueryRow("SELECT COUNT(*) FROM contoh").Scan(&count); err != nil {
		t.Fatalf("query backup error: %v", err)
	}
	if count != 1 {
		t.Fatalf("data backup tidak sesuai, got %d", count)
	}
}
