package config

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Debug   bool
	Console bool

	LocalEndpoint    string
	StoreDir         string
	StoreMaxSizeMB   uint32
	LogDir           string
	UplinksURL       []string
	DefaultCacheTime time.Duration
	CachePolicyMap   map[string]time.Duration
}

func getDefaultConfig() *Config {
	return &Config{
		LocalEndpoint:    ":8081",
		StoreDir:         "data",
		StoreMaxSizeMB:   16384,
		LogDir:           "log",
		UplinksURL:       []string{"https://alpha.de.repo.voidlinux.org"},
		DefaultCacheTime: 1 * time.Minute,
		CachePolicyMap: map[string]time.Duration{
			"":      5 * time.Minute,     // Don't cache
			".xbps": 30 * 24 * time.Hour, // ~1 month
			".sig":  30 * 24 * time.Hour, // ~1 month
		},
	}
}

func LoadConfig(f string, dbg, con bool) (*Config, error) {
	c := getDefaultConfig()
	if dbg {
		c.Debug = true
	}
	if con {
		c.Console = true
	}

	if _, err := toml.DecodeFile(f, c); err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	return c, nil
}
