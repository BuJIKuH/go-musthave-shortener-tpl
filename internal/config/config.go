package config

import (
	"flag"
	"fmt"
)

type Storage interface {
	Save(id, url string)
	Get(id string) (string, bool)
}

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
	flag.StringVar(&cfg.ShortenAddress, "b", "localhost:8080", "ShortAddress to listen on")

	flag.Parse()

	return &cfg

}
