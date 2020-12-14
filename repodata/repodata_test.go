package repodata

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/p-kraszewski/xbps-cache/logging"
)

func TestReadRepofile(t *testing.T) {
	logCfg := &logging.Config{
		Debug:   true,
		Console: true,
	}

	logging.Setup(logCfg)

	a := assert.New(t)

	rf, err := LoadDB("../.test_data/x86_64-repodata")
	a.Nil(err)
	a.NotNil(rf)

	t.Logf("%+v", rf)
}
