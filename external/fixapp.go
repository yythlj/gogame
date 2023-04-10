package pitaya

import (
	"github.com/topfreegames/pitaya/v2/cluster"
	"github.com/topfreegames/pitaya/v2/logger"
	mods "github.com/topfreegames/pitaya/v2/modules"
	"github.com/topfreegames/pitaya/v2/service"
	"github.com/topfreegames/pitaya/v2/session"
	"github.com/topfreegames/pitaya/v2/timer"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
)

func (app *App) StartCustom() {
	if !app.server.Frontend && len(app.acceptors) > 0 {
		logger.Log.Fatal("acceptors are not allowed on backend servers")
	}

	if app.server.Frontend && len(app.acceptors) == 0 {
		logger.Log.Fatal("frontend servers should have at least one configured acceptor")
	}

	if app.serverMode == Cluster {
		if reflect.TypeOf(app.rpcClient) == reflect.TypeOf(&cluster.GRPCClient{}) {
			app.serviceDiscovery.AddListener(app.rpcClient.(*cluster.GRPCClient))
		}

		if err := app.RegisterModuleBefore(app.rpcServer, "rpcServer"); err != nil {
			logger.Log.Fatal("failed to register rpc server module: %s", err.Error())
		}
		if err := app.RegisterModuleBefore(app.rpcClient, "rpcClient"); err != nil {
			logger.Log.Fatal("failed to register rpc client module: %s", err.Error())
		}
		// set the service discovery as the last module to be started to ensure
		// all modules have been properly initialized before the server starts
		// receiving requests from other pitaya servers
		if err := app.RegisterModuleAfter(app.serviceDiscovery, "serviceDiscovery"); err != nil {
			logger.Log.Fatal("failed to register service discovery module: %s", err.Error())
		}
	}

	app.periodicMetrics()

	app.listenCustom()

	defer func() {
		timer.GlobalTicker.Stop()
		app.running = false
	}()

	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	// stop server
	select {
	case <-app.dieChan:
		logger.Log.Warn("the app will shutdown in a few seconds")
	case s := <-sg:
		logger.Log.Warn("got signal: ", s, ", shutting down...")
		close(app.dieChan)
	}

	logger.Log.Warn("server is stopping...")

	app.sessionPool.CloseAll()
	app.shutdownModules()
	app.shutdownComponents()
}

func (app *App) listenCustom() {
	app.startupComponents()
	// create global ticker instance, timer precision could be customized
	// by SetTimerPrecision
	timer.GlobalTicker = time.NewTicker(timer.Precision)

	logger.Log.Infof("starting server %s:%s", app.server.Type, app.server.ID)
	logger.Log.Infof("开启定制化 Dispatch %d", app.config.Concurrency.Handler.Dispatch)
	app.handlerService.DispatchCustom(app.config.Concurrency.Handler.Dispatch)
	for _, acc := range app.acceptors {
		a := acc
		go func() {
			for conn := range a.GetConnChan() {
				go app.handlerService.HandleCustom(conn)
			}
		}()
		if app.config.Acceptor.ProxyProtocol {
			logger.Log.Info("Enabling PROXY protocol for inbond connections")
			a.EnableProxyProtocol()
		} else {
			logger.Log.Debug("PROXY protocol is disabled for inbound connections")
		}
		go func() {
			a.ListenAndServe()
		}()

		logger.Log.Infof("listening with acceptor %s on addr %s", reflect.TypeOf(a), a.GetAddr())
	}

	if app.serverMode == Cluster && app.server.Frontend && app.config.Session.Unique {
		unique := mods.NewUniqueSession(app.server, app.rpcServer, app.rpcClient, app.sessionPool)
		app.remoteService.AddRemoteBindingListener(unique)
		app.RegisterModule(unique, "uniqueSession")
	}

	app.startModules()

	logger.Log.Info("all modules started!")

	app.running = true
}

func (app *App) GetHandleService() *service.HandlerService {
	return app.handlerService
}

func (app *App) GetSessionPool() session.SessionPool {
	return app.sessionPool
}
