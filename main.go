package main

import (
	"game/src"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/acceptor"
	"github.com/topfreegames/pitaya/v2/config"
	"github.com/topfreegames/pitaya/v2/groups"
	"log"
	"net/http"
	"time"
)

func main() {
	app := BuildApp()
	defer app.Shutdown()
	StartGame(app)
}

func BuildApp() *pitaya.App {
	conf := config.NewDefaultBuilderConfig()
	conf.Pitaya.Buffer.Handler.LocalProcess = 15
	conf.Pitaya.Heartbeat.Interval = time.Duration(15 * time.Second)
	conf.Pitaya.Buffer.Agent.Messages = 32
	conf.Pitaya.Handler.Messages.Compression = false
	conf.Pitaya.Concurrency.Handler.Dispatch = 25
	builder := pitaya.NewDefaultBuilder(true, "game", pitaya.Cluster, map[string]string{}, *conf)
	builder.AddAcceptor(acceptor.NewWSAcceptor(":3850"))
	builder.Groups = groups.NewMemoryGroupService(*config.NewDefaultMemoryGroupConfig())
	app := builder.Build()
	customApp := app.(*pitaya.App)
	src.RegisterApp(customApp)
	return customApp
}

func StartGame(app *pitaya.App) {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))
	go http.ListenAndServe(":3851", nil)
	app.StartCustom()
}
