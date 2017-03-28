package main

import (
	"os"
	"route-sync/application"
	"route-sync/config"
	"route-sync/pooler"
	"time"

	"code.cloudfoundry.org/lager"
)

func main() {
	logger := lager.NewLogger("route-sync")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	cfg := loadConfig(logger)

	pooler := pooler.ByTime(time.Duration(30*time.Second), logger)
	src := application.GetKubernetesSource(cfg, logger)
	router := application.GetCloudFoundryRouter(cfg, logger)

	app := application.NewApplication(logger, pooler, src, router)
	app.Run(1, cfg)
}

func loadConfig(logger lager.Logger) *config.Config {
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("parsing config", err)
	}

	return cfg
}
