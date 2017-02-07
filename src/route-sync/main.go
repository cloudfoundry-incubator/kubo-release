package main

import (
	"os"
	"route-sync/cloudfoundry"
	"route-sync/cloudfoundry/tcp"
	"route-sync/config"
	"route-sync/pooler"
	"time"

	k8s "route-sync/kubernetes"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/route-registrar/messagebus"
	uaa "code.cloudfoundry.org/uaa-go-client"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"
)

func main() {
	logger := lager.NewLogger("route-sync")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("parsing config", err)
	}

	kubecfg, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfigPath)
	if err != nil {
		logger.Fatal("building config from flags", err)
	}
	clientset, err := kubernetes.NewForConfig(kubecfg)
	if err != nil {
		logger.Fatal("creating clientset from kube config", err)
	}

	src := k8s.New(clientset)
	bus := messagebus.NewMessageBus(logger)
	err = bus.Connect(cfg.NatsServers)
	if err != nil {
		logger.Fatal("connecting to NATS", err)
	}

	uaaCfg := &uaaconfig.Config{
		ClientName:       cfg.RoutingApiUsername,
		ClientSecret:     cfg.RoutingApiClientSecret,
		UaaEndpoint:      cfg.UAAApiUrl,
		SkipVerification: cfg.SkipTlsVerification,
	}

	uaaClient, err := uaa.NewClient(logger, uaaCfg, clock.NewClock())
	if err != nil {
		logger.Fatal("creating UAA client", err)
	}
	tcpRouter, err := tcp.NewRoutingApi(uaaClient, cfg.CloudFoundryApiUrl)
	if err != nil {
		logger.Fatal("creating TCP router", err)
	}

	sink := cloudfoundry.NewSink(bus, tcpRouter)

	pooler := pooler.ByTime(time.Duration(10 * time.Second))
	done, tick := pooler.Start(src, sink)

	logger.Info("started, Ctrl+C to exit")
	for {
		<-tick
		logger.Info("announced!")
	}
	done <- struct{}{}
}
