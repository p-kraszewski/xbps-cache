package logging

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
)

type writer struct {
	dir       string
	lastStamp int64
	writer    io.WriteCloser
}

type Config struct {
	Dir     string
	Debug   bool
	Console bool
	Json    bool
}

var ll = logrus.New()

func (w *writer) Write(p []byte) (n int, err error) {
	now := time.Now()
	stamp := now.Unix() / 3600

	if stamp != w.lastStamp {
		w.lastStamp = stamp
		if w.writer != nil {
			w.writer.Close()
		}
		logname := now.Format("20060102-0304.log")
		filename := path.Join(w.dir, logname)
		w.writer, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.
			O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
	}
	return w.writer.Write(p)
}

func Get() *logrus.Logger { return ll }

func Setup(c *Config) *logrus.Logger {

	if c.Console {
		ll.Out = os.Stderr
	} else {
		if c.Json {
			ll.SetFormatter(&logrus.JSONFormatter{})
		} else {
			ll.SetFormatter(&logrus.TextFormatter{})
		}
		w := &writer{
			dir: c.Dir,
		}
		ll.Out = w
	}

	if c.Debug {
		ll.SetLevel(logrus.DebugLevel)
	} else {
		ll.SetLevel(logrus.InfoLevel)
	}

	return ll
}
