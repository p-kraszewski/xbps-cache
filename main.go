package main

import (
	"flag"

	"github.com/p-kraszewski/xbps-cache/config"
	"github.com/p-kraszewski/xbps-cache/logging"
)

var (
	dbgMode = flag.Bool("v", true, "Verbose logging")
	cfgFile = flag.String("cfg", "xbps-cache.conf", "Cache file")
	conLog  = flag.Bool("stderr", true, "Log to stderr")
	jsonLog = flag.Bool("json", false, "Log to files in JSON")
)

func main() {
	flag.Parse()

	cfg, err := config.LoadConfig(*cfgFile, *dbgMode, *conLog)
	if err != nil {
		panic(err)
	}

	logCfg := &logging.Config{
		Dir:     cfg.LogDir,
		Debug:   *dbgMode,
		Console: *conLog,
		Json:    *jsonLog,
	}

	log := logging.Setup(logCfg)
	log.Error("Hello")
}
