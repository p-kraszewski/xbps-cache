package main

import (
	"flag"
	"log"

	"github.com/p-kraszewski/xbps-cache/cache"
)

var (
	dbgMode = flag.Bool("v", false, "Verbose logging")
	cfgFile = flag.String("cfg", "xbps-cache.conf", "Cache file")
)

func main() {
	flag.Parse()

	server, err := cache.LoadConfig(*cfgFile, *dbgMode)
	if err != nil {
		panic(err)
	}

	server.UlNew()

	err = server.StNew()
	if err != nil {
		panic(err)
	}

	err = server.DlNew()
	if err != nil {
		panic(err)
	}

	if *dbgMode {
		log.Printf("%+#v", server)
	}

	select {}
}
