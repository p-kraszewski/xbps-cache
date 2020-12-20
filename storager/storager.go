package storager

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/p-kraszewski/xbps-cache/config"
	"github.com/p-kraszewski/xbps-cache/logging"
)

var (
	log *logrus.Logger

	baseDir string
	baseUrl string
	cache   = map[string]*Repository{}
)

func Config(c *config.Config, lc logging.Config) {
	lc.Suffix = "-storage"
	log = logging.New(&lc)
	baseDir = c.StoreDir
	baseUrl = c.UplinkURL

	go StartUpdServer()

	repos, err := filepath.Glob(baseDir + "/*/*-repodata")
	if err != nil {
		log.Errorln(err)
		return
	}
	for _, r := range repos {
		rr, _ := filepath.Rel(baseDir, r)
		rd := filepath.Dir(rr)
		arch := filepath.Base(rr)
		ReloadRepo(rd, arch)
	}

}

func flattenRepoName(repo string) string {
	repo = strings.Trim(repo, "/")
	return strings.Replace(repo, "/", "_", -1)
}

func mapRepoToDir(repo string, file string) (string, error) {
	var op string

	flat := flattenRepoName(repo)

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
