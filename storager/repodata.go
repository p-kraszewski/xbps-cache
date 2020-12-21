package storager

import (
	"archive/tar"
	"bytes"
	"encoding/hex"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
	"howett.net/plist"
)

const (
	PKG_LIST = "index.plist"
)

type (
	Pkg struct {
		Sha [32]byte
		Len uint64
	}

	Repository struct {
		id  string
		db  map[string]Pkg
		dir string
	}
)

func (r *Repository) LoadDB() error {
	log.Debugf("Scanning %s", r.dir)
	files, err := filepath.Glob(r.dir + "/*-repodata")
	if err != nil {
		return err
	}

	ans := map[string]Pkg{}

	for _, file := range files {
		func() error {
			log.Infof("Parsing repo descriptor %s", file)
			db := map[string]interface{}{}

			// file := path.Join(r.dir, r.file)
			ans[path.Base(file)] = Pkg{}

			ib, err := os.Open(file)
			if err != nil {
				return err
			}

			defer ib.Close()
			id, err := zstd.NewReader(ib)
			if err != nil {
				return err
			}
			defer id.Close()

			tr := tar.NewReader(id)
			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break // End of archive
				}
				if err != nil {
					log.Error(err)
					return err
				}

				if hdr.Name != PKG_LIST {
					log.Debugf("Skipping %s", hdr.Name)
					continue
				}

				var buf bytes.Buffer

				buf.ReadFrom(tr)

				_, err = plist.Unmarshal(buf.Bytes(), &db)
				if err != nil {
					return err
				}
				break

			}

			for _, v := range db {
				switch vv := v.(type) {
				case map[string]interface{}:
					fname := vv["pkgver"].(string)
					farch := vv["architecture"].(string)
					fsha := vv["filename-sha256"].(string)

					fshad, err := hex.DecodeString(fsha)
					if err != nil {
						return err
					}

					fsiz := vv["filename-size"].(uint64)
					p := Pkg{
						Len: fsiz,
					}
					copy(p.Sha[:], fshad)
					ans[fname+"."+farch] = p
				}

			}
			return nil
		}()
	}

	r.db = ans
	log.Infof("Added %d files to cache", len(ans))

	return nil
}

func reloadCache(repo string) error {
	repo = flattenRepoName(repo)
	repoPath := path.Join(baseDir, repo)
	if _, found := cache[repo]; !found {
		cache[repo] = &Repository{
			id:  repo,
			db:  nil,
			dir: repoPath,
		}
	}

	return cache[repo].LoadDB()
}

func GetFileSha256(repo, file string) []byte {
	if path.Ext(file) != ".xbps" {
		return nil
	}

	file = file[:len(file)-5]

	repo = flattenRepoName(repo)

	if rec, found := cache[repo]; found {
		if fi, found := rec.db[file]; found {
			return fi.Sha[:]
		} else {
			log.Warnf("No file of %s:%s", repo, file)
			return nil
		}
	} else {
		log.Warnf("No repo of %s", repo)
		return nil
	}
}
