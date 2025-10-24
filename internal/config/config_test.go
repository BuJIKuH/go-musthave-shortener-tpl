package config_test

import (
	"flag"
	"os"
	"testing"

	"github.com/BuJIKuH/go-musthave-shortener-tpl/internal/config"
	"github.com/stretchr/testify/assert"
)

func resetEnvAndFlags() {
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestInitConfig(t *testing.T) {
	tests := []struct {
		name            string
		envAddress      string
		envBaseURL      string
		args            []string
		wantAddress     string
		wantShortenAddr string
	}{
		{
			name:            "default values",
			wantAddress:     "localhost:8080",
			wantShortenAddr: "http://localhost:8080",
		},
		{
			name:            "flags only",
			args:            []string{"-a", "localhost:9999", "-b", "http://short.io"},
			wantAddress:     "localhost:9999",
			wantShortenAddr: "http://short.io",
		},
		{
			name:            "env only",
			envAddress:      "localhost:7000",
			envBaseURL:      "http://env-base",
			wantAddress:     "localhost:7000",
			wantShortenAddr: "http://env-base",
		},
		{
			name:            "env has priority over flags",
			envAddress:      "localhost:6000",
			envBaseURL:      "http://env-priority",
			args:            []string{"-a", "localhost:9999", "-b", "http://short.io"},
			wantAddress:     "localhost:6000",
			wantShortenAddr: "http://env-priority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetEnvAndFlags()

			if tt.envAddress != "" {
				os.Setenv("SERVER_ADDRESS", tt.envAddress)
			}
			if tt.envBaseURL != "" {
				os.Setenv("BASE_URL", tt.envBaseURL)
			}

			if len(tt.args) > 0 {
				os.Args = append([]string{"cmd"}, tt.args...)
			} else {
				os.Args = []string{"cmd"}
			}

			cfg := config.InitConfig()

			assert.Equal(t, tt.wantAddress, cfg.Address)
			assert.Equal(t, tt.wantShortenAddr, cfg.ShortenAddress)
		})
	}
}
