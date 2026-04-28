package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	DBPath    string
	BackupDir string
	ExportDir string
	Debug     bool
}

func Load() (Config, error) {
	var (
		cfg Config
		err error
	)
	cfg.DBPath = getEnv("LEDGER_DB", "ledger.db")
	cfg.BackupDir = getEnv("LEDGER_BACKUP_DIR", "backups")
	cfg.ExportDir = getEnv("LEDGER_EXPORT_DIR", "exports")
	cfg.Debug, err = parseBoolEnv("LEDGER_DEBUG", false)
	if err == nil && strings.TrimSpace(cfg.DBPath) == "" {
		err = errors.New("LEDGER_DB is required")
	}
	return cfg, err
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if strings.TrimSpace(val) == "" {
		val = fallback
	}
	return val
}

func parseBoolEnv(key string, fallback bool) (bool, error) {
	var (
		val bool
		err error
	)
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		val = fallback
	}
	if raw != "" {
		val, err = parseBool(raw)
	}
	return val, err
}

func parseBool(raw string) (bool, error) {
	var (
		val bool
		err error
	)
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "y", "on":
		val = true
	case "0", "false", "no", "n", "off":
		val = false
	default:
		err = errors.New("invalid boolean value")
	}
	return val, err
}
