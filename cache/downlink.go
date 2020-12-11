package cache

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
)

func (c *Cache) DlNew() error {

	m := http.NewServeMux()
	m.HandleFunc(`/`, c.dlHandler)

	c.dlServer = &http.Server{
		Addr:           c.LocalEndpoint,
		Handler:        m,
		ReadTimeout:    1200 * time.Second,
		WriteTimeout:   1200 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		err := c.dlServer.ListenAndServe()
		if err != http.ErrServerClosed {
			panic(err)
		}
	}()
	return nil
}

func (c *Cache) DlClose(t time.Duration) {
	co, ca := context.WithTimeout(context.Background(), t)
	defer ca()
	_ = c.dlServer.Shutdown(co)
	_ = c.dlServer.Close()
}

func (c *Cache) dlHandler(w http.ResponseWriter, r *http.Request) {
	file := path.Clean(r.RequestURI)

	peer := strings.Split(r.RemoteAddr, ":")[0]

	pol := c.StGetPolicy(file)

	if pol == 0 { // Direct

		url := c.UplinksURL[0] + file
		resp, err := c.ulClient.Get(url)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Passthru error (Get) %v", err)
			log.Printf("Passthru error (Get) %v", err)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(200)
		l, err := io.Copy(w, resp.Body)
		if err != nil {
			log.Printf("Passthru error (Copy) %v", err)
		} else {
			log.Printf("[PASS] %s%s: %d bytes", peer, file, l)
		}
	} else {
		first := true
		for {
			if _, found := c.stLocks[file]; found {
				if first {
					log.Printf("[WAIT] %s%s", peer, file)
					first = false
				}
				time.Sleep(1 * time.Second)
				continue
			}
			first = false

			if c.StCheckFile(file) {
				delete(c.stLocks, file)
				rd, err := c.StFileReader(file)
				if err != nil {
					w.WriteHeader(500)
					fmt.Fprintf(w, "Read cache error (StFileReader) %v", err)
					log.Printf("Read cache error (StFileReader) %v", err)
					return
				}
				defer rd.Close()
				l, err := io.Copy(w, rd)
				if err != nil {
					log.Printf("Read cache error (copy) %v", err)
				} else {
					log.Printf("[CACH] %s%s: %d bytes", peer, file, l)
				}

			} else {
				c.stLocks[file] = struct{}{}
				defer delete(c.stLocks, file)
				url := c.UplinksURL[0] + file
				resp, err := c.ulClient.Get(url)
				if err != nil {
					w.WriteHeader(500)
					fmt.Fprintf(w, "Read+cache error (Get) %v", err)
					log.Printf("Read+cache error (Get) %v", err)
					return
				}
				defer resp.Body.Close()

				wrr, err := c.StFileWriter(file)
				if err != nil {
					w.WriteHeader(500)
					fmt.Fprintf(w, "Read+cache error (StFileWriter) %v", err)
					log.Printf("Read+cache error (StFileWriter) %v", err)
					c.StFileDelete(file)
					return
				}

				mw := io.MultiWriter(w, wrr)
				w.WriteHeader(200)

				buf := make([]byte, 1024*1024)

				l, err := io.CopyBuffer(mw, resp.Body, buf)
				wrr.Close()

				if err != nil {
					log.Printf("Read+cache error (copy) %v", err)
					c.StFileDelete(file)
				} else {
					log.Printf("[DOWN] %s%s: %d bytes", peer, file, l)
					c.stTsCache[file] = time.Now()
				}

			}
			return
		}
	}
}
