package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"alat-gudang/internal/backup"
	"alat-gudang/internal/config"
	"alat-gudang/internal/store"
	"alat-gudang/internal/timefmt"
	"alat-gudang/internal/web"
)

var Version = "0.1.0-mvp"

func main() {
	cfg := config.Load()

	if cfg.Version {
		fmt.Printf("Alat Gudang version %s\n", Version)
		return
	}

	setupLogging(cfg.LogDir)

	slog.Info("Memulai aplikasi Alat Gudang", "version", Version, "addr", cfg.Addr, "db", cfg.DBPath)

	// 1. Buat folder data jika belum ada
	if dir := filepath.Dir(cfg.DBPath); dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	// 2. Buka Database
	db, err := store.Open(cfg.DBPath)
	if err != nil {
		slog.Error("Gagal membuka database SQLite", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// 3. Inisialisasi Store
	st, err := store.New(db, timefmt.RealClock{}, cfg.TimeZone)
	if err != nil {
		slog.Error("Gagal inisialisasi store", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 4. Periksa integritas database sebelum menyentuh data
	if err := st.QuickCheck(ctx); err != nil {
		slog.Error("Integritas database bermasalah; data tidak ditimpa", "err", err)
		os.Exit(1)
	}

	// 5. Jalankan Migrasi
	if err := st.Migrate(ctx); err != nil {
		slog.Error("Gagal menjalankan migrasi", "err", err)
		os.Exit(1)
	}
	slog.Info("Migrasi database berhasil diverifikasi")

	// 6. Impor CSV (opsional, dijalankan sebelum server menerima koneksi)
	if err := runImports(ctx, st, cfg); err != nil {
		slog.Error("Impor CSV gagal", "err", err)
		os.Exit(1)
	}

	// 7. Seed data contoh HANYA jika diminta secara eksplisit dengan -seed.
	// Data produksi tidak boleh terisi contoh secara tidak sengaja.
	if cfg.SeedSample {
		if err := st.SeedSampleData(ctx); err != nil {
			slog.Warn("Gagal seed sample data", "err", err)
		} else {
			slog.Info("Sample data berhasil di-seed")
		}
	} else if isEmptyDatabase(ctx, st) {
		slog.Info("Database masih kosong. Gunakan -seed untuk data contoh, atau isi lewat menu Kelola.")
	}

	// 8. Inisialisasi Backup Manager
	backupMgr := backup.NewManager(db, cfg.BackupDir, 14)
	if _, err := backupMgr.PerformBackup(ctx); err != nil {
		slog.Warn("Backup awal gagal", "err", err)
	}

	// Ticker backup berkala tiap 24 jam
	backupTicker := time.NewTicker(24 * time.Hour)
	stopBackup := make(chan struct{})
	defer func() {
		close(stopBackup)
		backupTicker.Stop()
	}()
	go func() {
		for {
			select {
			case <-backupTicker.C:
				bCtx, bCancel := context.WithTimeout(context.Background(), 30*time.Second)
				_, _ = backupMgr.PerformBackup(bCtx)
				bCancel()
			case <-stopBackup:
				return
			}
		}
	}()

	// 9. Setup Renderer & Server HTTP
	web.Version = Version
	renderer, err := web.NewRenderer(cfg.DevMode)
	if err != nil {
		slog.Error("Gagal memuat template HTML", "err", err)
		os.Exit(1)
	}

	srv := web.NewServer(st, renderer)
	router := srv.Routes(cfg.DevMode)

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 10. Graceful shutdown handler
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("Server HTTP siap menerima koneksi", "url", "http://"+cfg.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server HTTP berhenti tidak normal", "err", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("Menutup aplikasi Alat Gudang dengan aman...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Shutdown HTTP server error", "err", err)
	}

	// Backup terakhir sebelum keluar
	lastBackupCtx, lastBackupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer lastBackupCancel()
	if _, err := backupMgr.PerformBackup(lastBackupCtx); err != nil {
		slog.Warn("Backup terakhir gagal", "err", err)
	}

	slog.Info("Aplikasi Alat Gudang telah selesai dengan aman.")
}

// setupLogging menulis log ke stdout dan file harian di folder log.
// File yang lebih tua dari 30 hari dibersihkan saat start.
func setupLogging(logDir string) {
	writer := io.Writer(os.Stdout)

	if err := os.MkdirAll(logDir, 0755); err == nil {
		logFileName := fmt.Sprintf("app-%s.log", time.Now().Format("20060102"))
		logFilePath := filepath.Join(logDir, logFileName)
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			writer = io.MultiWriter(os.Stdout, logFile)
		}
		cleanOldLogs(logDir, 30)
	}

	logger := slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
}

func cleanOldLogs(logDir string, keepDays int) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -keepDays)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "app-") || !strings.HasSuffix(name, ".log") {
			continue
		}
		info, err := entry.Info()
		if err != nil || !info.ModTime().Before(cutoff) {
			continue
		}
		_ = os.Remove(filepath.Join(logDir, name))
	}
}

func isEmptyDatabase(ctx context.Context, st *store.Store) bool {
	var count int
	if err := st.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM technicians").Scan(&count); err != nil {
		return false
	}
	if count > 0 {
		return false
	}
	if err := st.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM tools").Scan(&count); err != nil {
		return false
	}
	return count == 0
}

func runImports(ctx context.Context, st *store.Store, cfg *config.Config) error {
	if cfg.ImportTools != "" {
		file, err := os.Open(cfg.ImportTools)
		if err != nil {
			return fmt.Errorf("buka %s: %w", cfg.ImportTools, err)
		}
		defer file.Close()

		result, err := st.ImportToolsCSV(ctx, file)
		if err != nil {
			return fmt.Errorf("impor alat: %w", err)
		}
		slog.Info("Impor alat selesai", "created", result.Created, "updated", result.Updated)
	}

	if cfg.ImportTechnicians != "" {
		file, err := os.Open(cfg.ImportTechnicians)
		if err != nil {
			return fmt.Errorf("buka %s: %w", cfg.ImportTechnicians, err)
		}
		defer file.Close()

		result, err := st.ImportTechniciansCSV(ctx, file)
		if err != nil {
			return fmt.Errorf("impor teknisi: %w", err)
		}
		slog.Info("Impor teknisi selesai", "created", result.Created, "updated", result.Updated)
	}

	return nil
}
