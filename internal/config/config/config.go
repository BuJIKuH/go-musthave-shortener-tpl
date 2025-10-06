package config

import (
	"flag"
	"fmt"
)

type Config struct {
	Address        string
	ShortenAddress string
}

func (f *Config) String() string {
	return fmt.Sprintf("--a %s --b %s", f.Address, f.ShortenAddress)
}

func InitConfig() *Config {
	var cfg Config

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Address to listen on")
	flag.StringVar(&cfg.ShortenAddress, "b", "localhost:8000", "ShortAddress to listen on")

	flag.Parse()

	return &cfg

}
