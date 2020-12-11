package cache

import (
	"net/http"
	"time"
)

func (c *Cache) UlNew() {
	t := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 1200 * time.Second,
	}

	c.ulClient = &http.Client{Transport: t}
}
