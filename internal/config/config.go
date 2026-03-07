// Package config предоставляет функционал для загрузки конфигурации
// сервиса сокращения URL из флагов, переменных окружения и .env файла.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the configuration settings for the URL shortener service.
// Fields include server address, base URL, storage paths, and auth secret.
type Config struct {
	Address         string `env:"SERVER_ADDRESS"`
	ShortenAddress  string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDNS     string `env:"DATABASE_DNS"`
	AuthSecret      string `env:"AUTH_SECRET"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`

	EnableHTTPS bool `env:"ENABLE_HTTPS"`
}

type jsonConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// String returns a string representation of the config for logging or debugging.
func (f *Config) String() string {
	return fmt.Sprintf(
		"--a %s --b %s --f %s --d %s --af %s --au %s --s %t",
		f.Address,
		f.ShortenAddress,
		f.FileStoragePath,
		f.DatabaseDNS,
		f.AuditFile,
		f.AuditURL,
		f.EnableHTTPS,
	)
}

// InitConfig initializes and returns the service configuration.
// It parses flags and environment variables, falling back to defaults.
func InitConfig() *Config {
	var cfg Config

	_ = godotenv.Load()

	defaultAddr := "localhost:8080"
	defaultBase := "http://localhost:8080"
	defaultStoragePath := "./storageJson.json"

	var (
		secure     bool
		configPath string
	)

	flag.StringVar(&cfg.Address, "a", "", "Address to listen on")
	flag.StringVar(&cfg.ShortenAddress, "b", "", "Base URL for shortened links")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "File storage path")
	flag.StringVar(&cfg.DatabaseDNS, "d", "", "Database DNS")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "audit log file path")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "audit http endpoint")
	flag.BoolVar(&secure, "s", false, "enable https")

	flag.StringVar(&configPath, "c", "", "config file path")
	flag.StringVar(&configPath, "config", "", "config file path")

	flag.Parse()

	// --- JSON CONFIG (самый низкий приоритет) ---
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	if configPath != "" {
		if jc, err := loadJSONConfig(configPath); err == nil {
			if jc.ServerAddress != "" {
				cfg.Address = jc.ServerAddress
			}
			if jc.BaseURL != "" {
				cfg.ShortenAddress = jc.BaseURL
			}
			if jc.FileStoragePath != "" {
				cfg.FileStoragePath = jc.FileStoragePath
			}
			if jc.DatabaseDSN != "" {
				cfg.DatabaseDNS = jc.DatabaseDSN
			}
			cfg.EnableHTTPS = jc.EnableHTTPS
		}
	}

	// --- ENV ---
	if v := os.Getenv("SERVER_ADDRESS"); v != "" {
		cfg.Address = v
	}
	if v := os.Getenv("BASE_URL"); v != "" {
		cfg.ShortenAddress = v
	}
	if v := os.Getenv("FILE_STORAGE_PATH"); v != "" {
		cfg.FileStoragePath = v
	}
	if v := os.Getenv("DATABASE_DNS"); v != "" {
		cfg.DatabaseDNS = v
	}
	if v := os.Getenv("AUTH_SECRET"); v != "" {
		cfg.AuthSecret = v
	}
	if v := os.Getenv("AUDIT_FILE"); v != "" {
		cfg.AuditFile = v
	}
	if v := os.Getenv("AUDIT_URL"); v != "" {
		cfg.AuditURL = v
	}
	if os.Getenv("ENABLE_HTTPS") == "true" {
		cfg.EnableHTTPS = true
	}

	// --- FLAGS (самый высокий приоритет) ---
	if secure {
		cfg.EnableHTTPS = true
	}

	// --- DEFAULTS ---
	if cfg.Address == "" {
		cfg.Address = defaultAddr
	}
	if cfg.ShortenAddress == "" {
		cfg.ShortenAddress = defaultBase
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = defaultStoragePath
	}

	return &cfg
}

func loadJSONConfig(path string) (*jsonConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg jsonConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
