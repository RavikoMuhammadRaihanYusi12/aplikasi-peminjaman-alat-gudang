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
}

func Load() *Config {
	cfg := &Config{
		Addr:      getEnv("ADDR", "127.0.0.1:8080"),
		DBPath:    getEnv("DB_PATH", "data/gudang.db"),
		BackupDir: getEnv("BACKUP_DIR", "backup"),
		LogDir:    getEnv("LOG_DIR", "logs"),
		TimeZone:  getEnv("TZ", "Asia/Jakarta"),
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
