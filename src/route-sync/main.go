package main

import (
	"context"
	"flag"
	"os"
	"route-sync/application"
	"route-sync/config"
	"route-sync/pooler"
	"time"

	"route-sync/cloudfoundry"
	"route-sync/kubernetes"

	"code.cloudfoundry.org/lager"
)

var configPathFlag = flag.String("configPath", "route-sync.yml", "path to configuration file with json encoded content")

func main() {
	logger := lager.NewLogger("route-sync")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	flag.Parse()
	cfg := loadConfig(logger, *configPathFlag)

	applicationPooler := pooler.ByTime(time.Duration(30*time.Second), logger)
	src := kubernetes.DefaultSourceBuilder().CreateSource(cfg, logger)

	router := cloudfoundry.DefaultRouterBuilder().CreateRouter(cfg, logger)

	app := application.NewApplication(logger, applicationPooler, src, router)
	ctx := context.Background()
	app.Run(ctx, cfg)
}

func loadConfig(logger lager.Logger, configPath string) *config.Config {
	cfgSchema, err := config.NewConfigSchemaFromFile(configPath)
	if err != nil {
		logger.Fatal("loading config", err)
	}

	cfg, err := cfgSchema.ToConfig()
	if err != nil {
		logger.Fatal("parsing config", err)
	}

	return cfg
}
