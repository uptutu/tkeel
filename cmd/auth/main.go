package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/tkeel-io/tkeel"
	"github.com/tkeel-io/tkeel/logger"
	"github.com/tkeel-io/tkeel/service/auth"
	"github.com/tkeel-io/tkeel/service/auth/api"
	"github.com/tkeel-io/tkeel/version"
)

var (
	_log = logger.NewLogger("tKeel.auth")
)

func main() {
	logger.SetPluginVersion(version.Version())

	_log.Infof("[tKeel] starting auth -- version %s -- commit %s", version.Version(), version.Commit())
	plugin, err := tkeel.NewPluginFromFlags()
	if err != nil {
		_log.Fatalf("error init plugin: %s", err)
		return
	}

	pluginAuth := auth.NewPluginAuth(plugin)
	api.InitEntityIdp("./id_rsa", "./id_rsa.pem")
	pluginAuth.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop
}
