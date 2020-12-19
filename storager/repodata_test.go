package storager

import (
	"testing"

	"github.com/p-kraszewski/xbps-cache/logging"
)

func TestReadRepofile(t *testing.T) {
	logCfg := &logging.Config{
		Debug:   true,
		Console: true,
	}

	log = logging.New(logCfg)

	// a := assert.New(t)
	//
	// rf, err := LoadDB("../.test_data/x86_64-repodata")
	// a.Nil(err)
	// a.NotNil(rf)
	//
	// t.Logf("%+v", rf)
}
