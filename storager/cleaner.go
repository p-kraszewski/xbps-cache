package storager

import (
	"os"
	"path"
	"path/filepath"
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
			err := reloadCache(req.repo, req.arch)
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
	files := map[string]struct{}{}

	for n, r := range cache {
		rr := path.Dir(n)
		if rr == repo {
			// Protecting repo descriptor
			files[path.Base(n)] = struct{}{}
			log.Debugf("Scanning %s", rr)

			for f := range r.db {
				// log.Debugln(f)
				files[f+".xbps"] = struct{}{}
				files[f+".xbps.sig"] = struct{}{}
			}
		}
	}

	if len(files) > 1 {
		err := filepath.Walk(
			GetRepoPath(repo),
			func(path string, info os.FileInfo, err error) error {
				if info.Mode().IsRegular() {
					spath := filepath.Base(path)
					if _, found := files[spath]; !found {
						log.Infof("Removing %s", path)
						deleted += 1
						return os.Remove(path)
					} else {
						log.Tracef("Preserving %s", path)
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
