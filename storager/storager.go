package storager

import (
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/p-kraszewski/xbps-cache/config"
	"github.com/p-kraszewski/xbps-cache/logging"
)

var (
	log *logrus.Logger

	baseDir string
	baseUrl string
	cache   = map[RepoID]Repository{}
)

func Config(c *config.Config, lc logging.Config) {
	lc.Suffix = "-storage"
	log = logging.New(&lc)
	baseDir = c.StoreDir
	baseUrl = c.UplinkURL
}

func mapRepoToDir(repo string, file string) (string, error) {
	var op string

	repo = strings.Trim(repo, "/")
	flat := strings.Replace(repo, "/", "_", -1)

	if strings.HasSuffix(file, "-repodata") {
		op = path.Join(baseDir, flat)
	} else {
		var fd string

		if strings.HasPrefix(file, "lib") {
			fd = file[0:4]
		} else {
			fd = file[0:1]
		}
		fd = strings.ToLower(fd)
		op = path.Join(baseDir, flat, fd)
	}

	if err := os.MkdirAll(op, 0700); err != nil {
		return "", err
	}
	return op, nil

}
