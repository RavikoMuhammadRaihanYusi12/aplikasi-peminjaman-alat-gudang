package backup

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Manager struct {
	db        *sql.DB
	backupDir string
	keepCount int
}

func NewManager(db *sql.DB, backupDir string, keepCount int) *Manager {
	if keepCount <= 0 {
		keepCount = 14
	}
	return &Manager{
		db:        db,
		backupDir: backupDir,
		keepCount: keepCount,
	}
}

func (m *Manager) PerformBackup(ctx context.Context) (string, error) {
	if err := os.MkdirAll(m.backupDir, 0755); err != nil {
		return "", fmt.Errorf("buat folder backup gagal: %w", err)
	}

	// Nama memakai timestamp ber-milidetik agar urutan leksikografis sama dengan
	// urutan kronologis (dipakai saat membersihkan backup lama).
	now := time.Now()
	var targetPath string
	for i := 0; ; i++ {
		timestamp := now.Add(time.Duration(i) * time.Millisecond).Format("20060102-150405.000")
		targetPath = filepath.Join(m.backupDir, fmt.Sprintf("gudang-%s.db", timestamp))
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			break
		}
	}

	// VACUUM INTO membuat salinan yang konsisten secara transaksi.
	// Path di-bind sebagai parameter, bukan disusun menjadi SQL.
	if _, err := m.db.ExecContext(ctx, "VACUUM INTO ?", filepath.ToSlash(targetPath)); err != nil {
		return "", fmt.Errorf("eksekusi VACUUM INTO gagal: %w", err)
	}

	slog.Info("Backup database berhasil dibuat", "file", targetPath)

	// Clean up older backups exceeding keepCount
	m.cleanOldBackups()

	return targetPath, nil
}

func (m *Manager) cleanOldBackups() {
	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		return
	}

	var backupFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".db" {
			backupFiles = append(backupFiles, filepath.Join(m.backupDir, entry.Name()))
		}
	}

	if len(backupFiles) <= m.keepCount {
		return
	}

	// Sort oldest first
	sort.Strings(backupFiles)
	toDeleteCount := len(backupFiles) - m.keepCount
	for i := 0; i < toDeleteCount; i++ {
		_ = os.Remove(backupFiles[i])
		slog.Info("Menghapus backup lama", "file", backupFiles[i])
	}
}
