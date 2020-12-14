package repodata

import (
	"archive/tar"
	"bytes"
	"encoding/hex"
	"io"
	"os"

	"github.com/klauspost/compress/zstd"
	"howett.net/plist"

	"github.com/p-kraszewski/xbps-cache/logging"
)

const (
	PKG_LIST = "index.plist"
)

var log = logging.Get()

type Pkg struct {
	Sha [32]byte
	Len uint64
}

type DB map[string]Pkg

func LoadDB(file string) (DB, error) {

	ib, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer ib.Close()
	id, err := zstd.NewReader(ib)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		if hdr.Name != PKG_LIST {
			log.Debugf("Skipping %s", hdr.Name)
			continue
		}

		log.Infof("Decoding %s", hdr.Name)
		var buf bytes.Buffer

		buf.ReadFrom(tr)

		_, err = plist.Unmarshal(buf.Bytes(), &db)
		if err != nil {
			return nil, err
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
				return nil, err
			}

			fsiz := vv["filename-size"].(uint64)
			p := Pkg{
				Len: fsiz,
			}
			copy(p.Sha[:], fshad)
			ans[fname] = p
		}

	}

	return ans, nil
}
