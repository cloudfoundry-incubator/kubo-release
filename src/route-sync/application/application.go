package application

import (
	"context"
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

// Execute the Application on the current goroutine.
//
// Application will run until the pooler exits or it is cancelled.
// Cancellation is handled by the provided ctx or optional abortFunc.
func (app *Application) Run(ctx context.Context, abortFunc AbortFunc, config *config.Config) {
	app.router.Connect(config.NatsServers, app.logger)
	poolerCtx, cancelFunc := context.WithCancel(ctx)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		app.pooler.Run(poolerCtx, app.src, app.router)
		wg.Done()
	}()

	if abortFunc != nil {
		go func() {
			wg.Add(1)
			abortFunc(poolerCtx, cancelFunc, app.logger)
			wg.Done()
		}()
	}

	wg.Wait()
	app.logger.Info("exiting")
}
