package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/tkeel-io/tkeel/logger"
	"github.com/tkeel-io/tkeel/module"
	"github.com/tkeel-io/tkeel/module/plugin"
	"github.com/tkeel-io/tkeel/version"
)

var (
	_log = logger.NewLogger("tKeel.plugins")
)

func main() {
	logger.SetPluginVersion(version.Version())
	_log.Infof("[tKeel] starting plugins -- version %s -- commit %s", version.Version(), version.Commit())
	p, err := module.NewPluginFromFlags()
	if err != nil {
		_log.Fatalf("error init plugin: %s", err)
		return
	}
	m, err := plugin.New(p)
	if err != nil {
		_log.Fatalf("error new plugins: %s", err)
		return
	}

	m.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop
}
