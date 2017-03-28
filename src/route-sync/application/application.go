package application

import (
	"os"
	"os/signal"
	"route-sync/config"
	"route-sync/pooler"
	"route-sync/route"
	"sync"

	"code.cloudfoundry.org/lager"
)

type Application struct {
	logger lager.Logger
	pooler pooler.Pooler
	src    route.Source
	router route.Router
}

func NewApplication(logger lager.Logger, pooler pooler.Pooler, src route.Source, router route.Router) Application {
	return Application{
		logger: logger,
		pooler: pooler,
		src:    src,
		router: router,
	}
}

func (app *Application) Run(waitGroupDelta int, config *config.Config) {
	app.router.Connect(config.NatsServers, app.logger)
	poolerDone := app.pooler.Start(app.src, app.router)

	wg := sync.WaitGroup{}
	wg.Add(waitGroupDelta)
	go gracefulExit(app.logger, &wg, poolerDone)
	wg.Wait()

	app.logger.Info("exiting")
}

// Catch SIGINT (Ctrl+C) and tell pooler to quit
func gracefulExit(logger lager.Logger, wg *sync.WaitGroup, poolerDone chan<- struct{}) {
	logger.Info("started, Ctrl+C to exit")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	logger.Info("recieved Ctrl+C, exiting")
	poolerDone <- struct{}{}
	wg.Done()
}
