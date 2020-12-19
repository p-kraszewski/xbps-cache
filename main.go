package main

import (
	"flag"

	"github.com/p-kraszewski/xbps-cache/config"
	"github.com/p-kraszewski/xbps-cache/logging"
	"github.com/p-kraszewski/xbps-cache/server"
	"github.com/p-kraszewski/xbps-cache/storager"
)

var (
	dbgMode = flag.Bool("v", false, "Verbose logging")
	cfgFile = flag.String("cfg", "/etc/xbps-cache.conf", "Cache file")
	conLog  = flag.Bool("stderr", false, "Log to stderr")
	jsonLog = flag.Bool("json", false, "Log to files in JSON")
)

func main() {
	flag.Parse()

	cfg, err := config.LoadConfig(*cfgFile, *dbgMode, *conLog)
	if err != nil {
		panic(err)
	}

	logCfg := logging.Config{
		Dir:     cfg.LogDir,
		Debug:   *dbgMode,
		Console: *conLog,
		Json:    *jsonLog,
	}

	log := logging.New(&logCfg)
	log.Debugf("%#+v", cfg)

	storager.Config(cfg, logCfg)

	server.Start(cfg, logCfg)

	select {}
}
