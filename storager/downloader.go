package storager

import (
	"bytes"
	"crypto/sha256"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"
)

type (
	RepoFile struct {
		inStream  io.ReadCloser
		outStream io.WriteCloser
		sum       hash.Hash
		sumExp    []byte
		cachePath string
	}
)

var (
	RepoClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
		},
	}

	inFlight = map[string]bool{}
)

func GetRepoPath(repo string) string {
	return path.Join(baseDir, flattenRepoName(repo))
}

func GetRepoData(repo, file string) ([]byte, bool, error) {
	var (
		stamp    time.Time
		srvStamp time.Time
		hasFile  bool
	)

	dir, err := mapRepoToDir(repo, file)
	if err != nil {
		return nil, false, err
	}

	fullPath := path.Join(dir, file)

	fi, err := os.Stat(fullPath)
	if err == nil {
		stamp = fi.ModTime().UTC()
		hasFile = true
	}

	uri := baseUrl + repo + "/" + file

	respHdr, err := RepoClient.Head(uri)
	if err != nil {
		srvStamp = time.Now()
	} else {
		srvStamp, err = time.Parse(time.RFC1123, respHdr.Header.Get("Last-Modified"))
		if err != nil {
			srvStamp = time.Now()
		}
	}

	if srvStamp.After(stamp) {
		resp, err := RepoClient.Get(uri)

		if err != nil {
			if hasFile {
				data, err := ioutil.ReadFile(fullPath)
				return data, true, err
			} else {
				return nil, false, err
			}
		}

		defer resp.Body.Close()

		switch resp.StatusCode {
		case 404:
			return nil, false, http.ErrMissingFile
		case 200:
			var buf bytes.Buffer
			_, err := io.Copy(&buf, resp.Body)
			if err != nil {
				return nil, false, err
			}
			err = ioutil.WriteFile(fullPath, buf.Bytes(), 0644)
			if err != nil {
				return nil, false, err
			}
			return buf.Bytes(), false, nil
		}

		return nil, false, http.ErrServerClosed
	} else {
		log.Debugf("Using cached version of %s ", fullPath)
		data, err := ioutil.ReadFile(fullPath)
		return data, true, err
	}
}

func GetFile(repo string, file string, expSum []byte) (*RepoFile, bool, error) {
	var (
		stamp    time.Time
		srvStamp time.Time
		hasFile  bool
	)

	dir, err := mapRepoToDir(repo, file)
	if err != nil {
		return nil, false, err
	}

	fullPath := path.Join(dir, file)

	for {
		if _, found := inFlight[fullPath]; !found {
			break
		}
		time.Sleep(time.Second)
	}

	fi, err := os.Stat(fullPath)
	if err == nil {
		stamp = fi.ModTime().UTC()
		hasFile = true
	}

	uri := baseUrl + repo + "/" + file

	respHdr, err := RepoClient.Head(uri)
	if err != nil {
		srvStamp = time.Now()
	} else {
		srvStamp, err = time.Parse(time.RFC1123, respHdr.Header.Get("Last-Modified"))
		if err != nil {
			srvStamp = time.Now()
		}
	}

	if srvStamp.After(stamp) {
		resp, err := RepoClient.Get(uri)

		if err != nil {
			if hasFile {
				data, err := os.Open(fullPath)
				return &RepoFile{inStream: data}, true, err
			} else {
				return nil, false, err
			}
		}

		switch resp.StatusCode {
		case 404:
			return nil, false, http.ErrMissingFile
		case 304:
			data, err := os.Open(fullPath)
			return &RepoFile{inStream: data}, true, err
		case 200:

			cacheFile, err := os.Create(fullPath)
			if err != nil {
				return nil, false, err
			}
			inFlight[fullPath] = true
			if expSum == nil {
				return &RepoFile{
					inStream:  resp.Body,
					outStream: cacheFile,
					cachePath: fullPath,
				}, false, nil
			} else {
				return &RepoFile{
					inStream:  resp.Body,
					outStream: cacheFile,
					sum:       sha256.New(),
					sumExp:    expSum,
					cachePath: fullPath,
				}, false, nil
			}
		}

		return nil, false, http.ErrServerClosed
	} else {
		data, err := os.Open(fullPath)
		return &RepoFile{inStream: data}, true, err
	}
}

func (rf *RepoFile) Read(p []byte) (n int, err error) {
	ln, err := rf.inStream.Read(p)
	if err != nil {
		return ln, err
	}
	if rf.outStream != nil {
		if rf.sum != nil {
			rf.sum.Write(p[:ln])
		}
		return rf.outStream.Write(p[:ln])
	}
	return ln, err
}

func (rf *RepoFile) Close() error {
	if rf.inStream != nil {
		rf.inStream.Close()
	}
	if rf.outStream != nil {
		rf.outStream.Close()
	}
	delete(inFlight, rf.cachePath)

	if rf.sum != nil {
		calc := rf.sum.Sum(nil)
		if !bytes.Equal(calc, rf.sumExp) {
			log.Warnf("Downloaded package bad checksum %s, "+
				"removing from cache", rf.cachePath)
			os.Remove(rf.cachePath)
			os.Remove(rf.cachePath + ".sig")
		}
	}

	return nil
}
