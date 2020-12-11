package cache

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

func (c *Cache) StNew() error {
	if err := c.StFillTsCache(); err != nil {
		return err
	}
	return nil
}

func (c *Cache) StFillTsCache() error {
	// Wipe old cache
	c.stTsCache = map[string]time.Time{}

	return filepath.Walk(c.StoreDir, func(path string, info os.FileInfo, err error) error {
		path, _ = filepath.Rel(c.StoreDir, path)

		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.Mode().IsRegular() {
			path = `/` + path
			c.stTsCache[path] = info.ModTime()
			if c.debug {
				log.Printf("Cached %s", path)
			}

		}
		return nil
	})
}

func (c *Cache) StGetPolicy(f string) time.Duration {
	fn := path.Base(f)
	ext := path.Ext(fn)

	pol := c.DefaultCacheTime

	if policy, found := c.CachePolicyMap[ext]; found {
		pol = policy
	}

	return pol
}

func (c *Cache) StExpireCache() {
	now := time.Now()
	for filePath, fileModStamp := range c.stTsCache {
		fileMaxAge := c.StGetPolicy(filePath)
		if fileModStamp.Add(fileMaxAge).Before(now) {
			delete(c.stTsCache, filePath)
			c.StFileDelete(filePath)
		}

	}
}

func (c *Cache) StCheckFile(filePath string) bool {
	if fileModStamp, found := c.stTsCache[filePath]; found {
		fileMaxAge := c.StGetPolicy(filePath)
		if fileModStamp.Add(fileMaxAge).Before(time.Now()) {
			c.StFileDelete(filePath)
			if c.debug {
				log.Printf("%s expired", filePath)
			}
			return false
		}
		if c.debug {
			log.Printf("%s valid", filePath)
		}
		return true
	}
	if c.debug {
		log.Printf("%s not present", filePath)
	}
	return false
}

func (c *Cache) StFileReader(filePath string) (io.ReadCloser, error) {
	realPath := path.Join(c.StoreDir, filePath)
	return os.Open(realPath)
}

func (c *Cache) StFileDelete(filePath string) error {
	delete(c.stTsCache, filePath)
	realPath := path.Join(c.StoreDir, filePath)
	err := os.Remove(realPath)
	if err != nil {
		log.Printf("Error purging file '%s': %v", filePath, err)
		return err
	} else {
		if c.debug {
			log.Printf("Purged '%s'", filePath)
		}
	}
	return nil
}

func (c *Cache) StFileWriter(filePath string) (io.WriteCloser, error) {
	realPath := path.Join(c.StoreDir, filePath)
	realDir := path.Dir(realPath)
	err := os.MkdirAll(realDir, 0700)
	if err != nil {
		return nil, err
	}
	return os.Create(realPath)
}
