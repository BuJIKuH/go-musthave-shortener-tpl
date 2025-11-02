package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Address         string `env:"SERVER_ADDRESS"`
	ShortenAddress  string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func (f *Config) String() string {
	return fmt.Sprintf("--a %s --b %s --f %s", f.Address, f.ShortenAddress, f.FileStoragePath)
}

func InitConfig() *Config {
	var cfg Config
	_ = godotenv.Load()

	defaultAddr := "localhost:8080"
	defaultBase := "http://localhost:8080"
	defaultStoragePath := "./storageJson.json"

	flag.StringVar(&cfg.Address, "a", "", "Address to listen on")
	flag.StringVar(&cfg.ShortenAddress, "b", "", "Base URL for shortened links")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "File storage path")
	flag.Parse()

	envAddress := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")
	envStoragePath := os.Getenv("FILE_STORAGE_PATH")

	if envAddress != "" {
		cfg.Address = envAddress
	} else if cfg.Address == "" {
		cfg.Address = defaultAddr
	}

	if envBaseURL != "" {
		cfg.ShortenAddress = envBaseURL
	} else if cfg.ShortenAddress == "" {
		cfg.ShortenAddress = defaultBase
	}

	if envStoragePath != "" {
		cfg.FileStoragePath = envStoragePath
	} else if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = defaultStoragePath
	}

	return &cfg
}
