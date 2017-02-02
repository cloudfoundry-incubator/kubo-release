package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"route-sync/cloudfoundry"
	"route-sync/fixed_source"
	"route-sync/pooler"
	"route-sync/route"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/route-registrar/config"
	"code.cloudfoundry.org/route-registrar/messagebus"
)

type serviceConfig struct {
	natsServers []config.MessageBusServer
}

func parseConfig(data []byte) (serviceConfig, error) {
	servers := serviceConfig{}
	err := json.Unmarshal(data, &servers)

	return servers, err
}

var (
	configPath = flag.String("configPath", "./config.json", "path to a route-sync config file")
)

func main() {
	logger := lager.NewLogger("route-sync")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	file, err := ioutil.ReadFile(*configPath)
	if err != nil {
		logger.Fatal("parsing config file", err)
	}
	config, err := parseConfig(file)

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
	bus.Connect(config.natsServers)
	sink := cloudfoundry.NewSink(bus)

	pooler := pooler.ByTime(time.Duration(10 * time.Second))
	done, tick := pooler.Start(src, sink)

	logger.Info("started, Ctrl+C to exit")
	for {
		<-tick
		logger.Info("announced!")
	}
	done <- struct{}{}
}
