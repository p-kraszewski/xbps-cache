package puller

import (
	"sync"

	"github.com/p-kraszewski/xbps-cache/config"
)

type None struct{}

var Nothing = None{}

type Puller struct {
	feedLock  sync.Mutex
	requestor chan string
	inFlight  map[string]None
}

func New(c *config.Config) (*Puller, error) {
	puller := &Puller{
		requestor: make(chan string),
		inFlight:  map[string]None{},
	}
	return puller, nil
}

func (p *Puller) Request(file string) {
	p.feedLock.Lock()
	defer p.feedLock.Unlock()

	p.inFlight[file] = Nothing
	p.requestor <- file
}

func (p *Puller) Close() {
	close(p.requestor)
}
