package main

import (
	"context"
	"os"
	"route-sync/application"
	"route-sync/config"
	"route-sync/pooler"
	"time"

	"code.cloudfoundry.org/lager"
	"route-sync/kubernetes"
	"route-sync/cloudfoundry"
)

func main() {
	logger := lager.NewLogger("route-sync")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	cfg := loadConfig(logger)

	applicationPooler := pooler.ByTime(time.Duration(30*time.Second), logger)
	src := kubernetes.DefaultSourceBuilder().CreateSource(cfg, logger)

	router := cloudfoundry.DefaultRouterBuilder().CreateRouter(cfg, logger)

	app := application.NewApplication(logger, applicationPooler, src, router)
	ctx := context.Background()
	app.Run(ctx, application.InterruptWaitFunc, cfg)
}

func loadConfig(logger lager.Logger) *config.Config {
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("parsing config", err)
	}

	return cfg
}
