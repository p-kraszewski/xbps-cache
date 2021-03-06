package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/p-kraszewski/xbps-cache/config"
	"github.com/p-kraszewski/xbps-cache/logging"
	"github.com/p-kraszewski/xbps-cache/storager"
)

var (
	log *logrus.Logger
	srv *http.Server
)

func Start(c *config.Config, lc logging.Config) {

	lc.Suffix = "-server"
	log = logging.New(&lc)

	m := http.NewServeMux()
	m.HandleFunc(`/`, dlHandler)

	srv = &http.Server{
		Addr:           c.LocalEndpoint,
		Handler:        m,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   1200 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			panic(err)
		}
	}()
}

func dlHandler(w http.ResponseWriter, r *http.Request) {
	fName := r.RequestURI
	peer := strings.Split(r.RemoteAddr, ":")[0]

	repo := path.Dir(fName)
	elem := path.Base(fName)
	ext := path.Ext(fName)

	switch ext {
	case "":
		data, cached, err := storager.GetRepoData(repo, elem)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(404)
			} else {
				w.WriteHeader(500)
			}
			log.WithField("peer", peer).
				WithField("repo", repo).
				WithField("file", elem).
				Errorln(err)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		w.Header().Add(
			`content-length`,
			strconv.FormatUint(uint64(len(data)), 10),
		)
		w.WriteHeader(200)
		w.Write(data)
		log.WithField("peer", peer).
			WithField("repo", repo).
			WithField("file", elem).
			WithField("cache", cached).
			Infoln(len(data))
		if !cached {
			log.Infof("Repo %s:%s updated, refreshing and cleaning", repo, elem)
			storager.ReloadRepo(repo, elem)
		}
		return

	case ".xbps", ".sig":
		csum, flen := storager.GetFileSha256AndLen(repo, elem)
		data, cached, err := storager.GetFile(repo, elem, csum, int64(flen))
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(404)
			} else {
				w.WriteHeader(500)
			}
			log.WithField("peer", peer).
				WithField("repo", repo).
				WithField("file", elem).
				Errorln(err)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}
		defer data.Close()
		w.Header().Add(`content-length`, strconv.FormatUint(flen, 10))
		w.WriteHeader(200)
		tl, err := io.Copy(w, data)
		if err != nil {
			log.WithField("peer", peer).
				WithField("repo", repo).
				WithField("file", elem).
				Errorln(err)
			return
		}
		log.WithField("peer", peer).
			WithField("repo", repo).
			WithField("file", elem).
			WithField("cache", cached).
			Infoln(tl)
		return

	default:
		log.Errorf("[%s] Unsupported file type for %s", peer, elem)
		w.WriteHeader(404)
		w.Write([]byte("Error"))
	}

	log.Debugf("%s : %s : %s", repo, elem, ext)

}
