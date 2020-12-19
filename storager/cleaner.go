package storager

import (
	"os"
	"path"
	"path/filepath"
)

func CleanRepo(repo string) error {
	repo = flattenRepoName(repo)
	files := map[string]struct{}{}

	for n, r := range cache {
		rr := path.Dir(n)
		if rr == repo {
			// Protecting repo descriptor
			files[path.Base(n)] = struct{}{}
			log.Debugf("Scanning %s", rr)

			for f := range r.db {
				files[f+".xbps"] = struct{}{}
				files[f+".xbps.sig"] = struct{}{}
			}
		}
	}

	if len(files) > 0 {
		err := filepath.Walk(
			GetRepoPath(repo),
			func(path string, info os.FileInfo, err error) error {
				if info.Mode().IsRegular() {
					spath := filepath.Base(path)
					if _, found := files[spath]; !found {
						log.Infof("Removing %s", path)
						return os.Remove(path)
					} else {
						log.Debugf("Preserving %s", path)
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
	return nil
}
