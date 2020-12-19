package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Debug   bool
	Console bool

	LocalEndpoint string
	StoreDir      string
	LogDir        string
	UplinkURL     string
}

func getDefaultConfig() *Config {

	// https://alpha.de.repo.voidlinux.org/current/debug/x86_64-repodata

	return &Config{
		LocalEndpoint: ":8081",
		StoreDir:      "data",
		LogDir:        "log",
		UplinkURL:     "https://alpha.de.repo.voidlinux.org",
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
