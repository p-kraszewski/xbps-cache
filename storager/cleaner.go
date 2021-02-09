package storager

import (
	"os"
	"path/filepath"
	"strings"
)

type (
	updRepoBG struct {
		repo string
		arch string
	}
)

var (
	updRepoChn = make(chan updRepoBG)
)

func ReloadRepo(repo string, arch string) {
	updRepoChn <- updRepoBG{
		repo: repo,
		arch: arch,
	}
}

func StartUpdServer() {
	go func() {
		for req := range updRepoChn {
			err := reloadCache(req.repo)
			if err != nil {
				log.WithField("repo", req.repo).
					WithField("arch", req.arch).
					Errorln(err)
			} else {
				log.Infof("Reloaded %s:%s", req.repo, req.arch)
				del, err := cleanRepo(req.repo)
				if err != nil {
					log.WithField("repo", req.repo).
						WithField("arch", req.arch).
						Errorln(err)
				} else {
					if del > 0 {
						log.Infof("Cleaned %s:%s - %d file(s)", req.repo,
							req.arch, del)
					} else {
						log.Infof("No outdated files in %s:%s", req.repo,
							req.arch)
					}
				}

			}

		}
	}()
}

func cleanRepo(repo string) (int, error) {
	deleted := 0
	repo = flattenRepoName(repo)
	files := map[string]int64{}

	for n, r := range cache {
		if n == repo {

			for f, v := range r.db {
				files[f+".xbps"] = int64(v.Len)
				files[f+".xbps.sig"] = 512
			}
		}
	}

	if len(files) > 1 {
		err := filepath.Walk(
			GetRepoPath(repo),
			func(path string, info os.FileInfo, err error) error {
				if info.Mode().IsRegular() {
					spath := filepath.Base(path)
					if !strings.HasSuffix(spath, "-repodata") {
						if expLen, found := files[spath]; found {
							if info.Size() != expLen {
								log.Warnf("Removing mismatched size %s", path)
								return os.Remove(path)
							} else {
								log.Tracef("Preserving %s", path)
							}
						} else {
							log.Infof("Removing %s", path)
							deleted += 1
							return os.Remove(path)

						}
					}
				}

				return nil
			})
		if err != nil {
			log.Errorln(err)
		}
	} else {
		log.Warnf("No files in descriptor found, won't clean")
	}
	return deleted, nil
}
