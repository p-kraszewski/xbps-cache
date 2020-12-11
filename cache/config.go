package cache

import (
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Cache struct {
	debug bool

	LocalEndpoint    string
	StoreDir         string
	StoreMaxSizeMB   uint32
	UplinksURL       []string
	DefaultCacheTime time.Duration
	CachePolicyMap   map[string]time.Duration

	// Downlink
	dlServer *http.Server

	// Uplink
	ulClient *http.Client

	// Storage
	stTsCache map[string]time.Time
	stLocks   map[string]struct{}
}

func getDefaultConfig() *Cache {
	return &Cache{
		LocalEndpoint:    ":8081",
		StoreDir:         "data",
		StoreMaxSizeMB:   16384,
		UplinksURL:       []string{"https://alpha.de.repo.voidlinux.org"},
		DefaultCacheTime: 1 * time.Minute,
		CachePolicyMap: map[string]time.Duration{
			"":      5 * time.Minute,     // Don't cache
			".xbps": 30 * 24 * time.Hour, // ~1 month
			".sig":  30 * 24 * time.Hour, // ~1 month
		},
		stTsCache: map[string]time.Time{},
		stLocks:   map[string]struct{}{},
	}
}

func LoadConfig(f string, dbg bool) (*Cache, error) {
	c := getDefaultConfig()
	c.debug = dbg
	if _, err := toml.DecodeFile(f, c); err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	return c, nil
}
