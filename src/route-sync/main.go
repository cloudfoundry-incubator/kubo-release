package main

import (
	"os"
	"os/signal"
	"route-sync/cloudfoundry"
	"route-sync/cloudfoundry/tcp"
	"route-sync/config"
	"route-sync/pooler"
	"sync"
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
		ClientName:       cfg.RoutingAPIUsername,
		ClientSecret:     cfg.RoutingAPIClientSecret,
		UaaEndpoint:      cfg.UAAApiURL,
		SkipVerification: cfg.SkipTLSVerification,
	}

	uaaClient, err := uaa.NewClient(logger, uaaCfg, clock.NewClock())
	if err != nil {
		logger.Fatal("creating UAA client", err)
	}
	tcpRouter, err := tcp.NewRoutingApi(uaaClient, cfg.CloudFoundryAPIURL)
	if err != nil {
		logger.Fatal("creating TCP router", err)
	}

	sink := cloudfoundry.NewRouter(bus, tcpRouter)

	pooler := pooler.ByTime(time.Duration(10 * time.Second))
	poolerDone, tick := pooler.Start(src, sink)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		// Catch SIGINT (Ctrl+C) and tell pooler to quit
		logger.Info("started, Ctrl+C to exit")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		<-sigChan
		poolerDone <- struct{}{}
		wg.Done()
	}()

	go func() {
		for range tick {
			logger.Info("announced!")
		}
		wg.Done()
	}()

	wg.Wait()
	logger.Info("exiting")
}
