package main

import (
	"os"
	"route-sync/cloudfoundry"
	"route-sync/config"
	"route-sync/fixed_source"
	"route-sync/pooler"
	"route-sync/route"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/route-registrar/messagebus"
)

func main() {
	logger := lager.NewLogger("route-sync")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("parsing config", err)
	}

	httpRoutes := []*route.HTTP{
		&route.HTTP{
			Name: "foo.bar.com",
			Backend: []route.Endpoint{
				route.Endpoint{
					IP:   "10.10.10.10",
					Port: 8080,
				},
			},
		},
	}

	// TODO: replace this with a kubernetes source
	src := fixed_source.New(nil, httpRoutes)

	bus := messagebus.NewMessageBus(logger)
	err = bus.Connect(cfg.NatsServers)
	if err != nil {
		logger.Fatal("connecting to NATS", err)
	}

	sink := cloudfoundry.NewSink(bus, nil)

	pooler := pooler.ByTime(time.Duration(10 * time.Second))
	done, tick := pooler.Start(src, sink)

	logger.Info("started, Ctrl+C to exit")
	for {
		<-tick
		logger.Info("announced!")
	}
	done <- struct{}{}
}
