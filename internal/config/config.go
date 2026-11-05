package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr              string
	DBPath            string
	BackupDir         string
	LogDir            string
	TimeZone          string
	DevMode           bool
	SeedSample        bool
	Version           bool
	ImportTools       string
	ImportTechnicians string
	TursoURL          string
	TursoToken        string
}

func Load() *Config {
	port := getEnv("PORT", "8080")
	cfg := &Config{
		Addr:      getEnv("ADDR", "0.0.0.0:"+port),
		DBPath:    getEnv("DB_PATH", "data/gudang.db"),
		BackupDir: getEnv("BACKUP_DIR", "backup"),
		LogDir:    getEnv("LOG_DIR", "logs"),
		TimeZone:  getEnv("TZ", "Asia/Jakarta"),
		TursoURL:  getEnv("TURSO_DATABASE_URL", ""),
		TursoToken: getEnv("TURSO_AUTH_TOKEN", ""),
	}

	flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "HTTP listen address")
	flag.StringVar(&cfg.DBPath, "db", cfg.DBPath, "Path to SQLite database")
	flag.StringVar(&cfg.BackupDir, "backup-dir", cfg.BackupDir, "Directory for database backups")
	flag.StringVar(&cfg.LogDir, "log-dir", cfg.LogDir, "Directory for application logs")
	flag.StringVar(&cfg.TimeZone, "tz", cfg.TimeZone, "Timezone for display")
	flag.BoolVar(&cfg.DevMode, "dev", false, "Enable development mode (reload templates from disk)")
	flag.BoolVar(&cfg.SeedSample, "seed", false, "Seed sample tools and technicians if database is empty")
	flag.BoolVar(&cfg.Version, "version", false, "Print application version and exit")
	flag.StringVar(&cfg.ImportTools, "import-tools", "", "Path to CSV file with tools (code,name,category,max_hours)")
	flag.StringVar(&cfg.ImportTechnicians, "import-technicians", "", "Path to CSV file with technicians (name,section)")

	flag.Parse()
	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
