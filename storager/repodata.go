package storager

import (
	"archive/tar"
	"bytes"
	"encoding/hex"
	"io"
	"os"
	"path"

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

	RepoID string

	Repository struct {
		id   string
		db   map[string]Pkg
		dir  string
		file string
	}
)

func (r *Repository) LoadDB() error {

	file := path.Join(r.dir, r.file)

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

	db := map[string]interface{}{}

	ans := map[string]Pkg{}

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
			ans[fname] = p
		}

	}
	r.db = ans

	return nil
}

func ReloadCache(repo, arch string) error {
	repo = flattenRepoName(repo) + "/" + arch
	repoPath := path.Join(baseDir, repo)
	if _, found := cache[repo]; !found {
		cache[repo] = &Repository{
			id:   repo,
			db:   nil,
			dir:  path.Dir(repoPath),
			file: path.Base(repoPath),
		}
	}

	return cache[repo].LoadDB()
}
