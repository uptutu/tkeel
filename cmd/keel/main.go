package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/tkeel-io/tkeel"
	"github.com/tkeel-io/tkeel/logger"
	"github.com/tkeel-io/tkeel/service/keel"
	"github.com/tkeel-io/tkeel/version"
)

var (
	log = logger.NewLogger("tKeel.keel")
)

func main() {
	logger.SetPluginVersion(version.Version())
	log.Infof("[tKeel] starting keel -- version %s -- commit %s", version.Version(), version.Commit())
	p, err := tkeel.NewPluginFromFlags()
	if err != nil {
		log.Fatalf("error init plugin: %s", err)
	}

	g, err := keel.New(p)
	if err != nil {
		log.Fatalf("error new keel: %s", err)
	}

	g.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop
}
