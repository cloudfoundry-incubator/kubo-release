package application

import (
	"context"
	"route-sync/config"
	"route-sync/pooler"
	"route-sync/route"

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
// Application will run until the pooler exits or it is cancelled by the ctx.
func (app *Application) Run(ctx context.Context, config *config.Config) {
	app.router.Connect(config.NatsServers, app.logger)

	app.pooler.Run(ctx, app.src, app.router)
	app.logger.Info("exiting")
}
