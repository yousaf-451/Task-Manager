// Package config centralizes application configuration.
//
// It reads values from OS environment variables, falling back to a local
// .env file (if present) for local development convenience, and finally to
// sane defaults. No third-party dependency is required.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds every configuration value the application needs.
type Config struct {
	ServerPort string
	ServerHost string

	CORSAllowedOrigins []string

	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBMaxOpenConns       int
	DBMaxIdleConns       int
	DBConnMaxLifetimeMin time.Duration
}

// Load reads the .env file (if present) into the process environment and
// then builds a Config from environment variables, applying defaults for
// anything that is missing.
func Load() (*Config, error) {
	// Best-effort: a missing .env file is not an error, since the app can
	// also be configured via real environment variables (e.g. in Docker).
	_ = loadDotEnv(".env")

	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),

		CORSAllowedOrigins: splitAndTrim(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),

		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "task_user"),
		DBPassword: getEnv("DB_PASSWORD", "task_password"),
		DBName:     getEnv("DB_NAME", "task_manager"),
	}

	maxOpen, err := getEnvInt("DB_MAX_OPEN_CONNS", 25)
	if err != nil {
		return nil, err
	}
	cfg.DBMaxOpenConns = maxOpen

	maxIdle, err := getEnvInt("DB_MAX_IDLE_CONNS", 10)
	if err != nil {
		return nil, err
	}
	cfg.DBMaxIdleConns = maxIdle

	lifetimeMin, err := getEnvInt("DB_CONN_MAX_LIFETIME_MIN", 5)
	if err != nil {
		return nil, err
	}
	cfg.DBConnMaxLifetimeMin = time.Duration(lifetimeMin) * time.Minute

	return cfg, nil
}

// DSN builds the MySQL Data Source Name used by database/sql.
func (c *Config) DSN() string {
	// parseTime=true lets the driver scan DATE/DATETIME columns directly
	// into time.Time. loc=Local keeps timestamps in the server's local zone.
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

// Addr returns the host:port the HTTP server should bind to.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.ServerHost, c.ServerPort)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("config: invalid integer for %s: %w", key, err)
	}
	return v, nil
}

func splitAndTrim(csv string) []string {
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// loadDotEnv parses a simple KEY=VALUE .env file and applies each entry to
// the process environment, without overriding variables that are already
// set (so real environment variables always take precedence).
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		// Not finding a .env file is fine; it's optional.
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}
