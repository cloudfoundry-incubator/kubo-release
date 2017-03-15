package main

import (
	"os"
	"os/signal"
	"route-sync/cloudfoundry"
	"route-sync/cloudfoundry/tcp"
	"route-sync/config"
	"route-sync/pooler"
	"route-sync/route"
	"sync"
	"time"

	"route-sync/kubernetes"

	k8sclient "k8s.io/client-go/kubernetes"
	k8sclientcmd "k8s.io/client-go/tools/clientcmd"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/route-registrar/messagebus"
	uaa "code.cloudfoundry.org/uaa-go-client"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"
)

func main() {
	logger := lager.NewLogger("route-sync")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	cfg := loadConfig(logger)

	pooler := pooler.ByTime(time.Duration(30*time.Second), logger)
	poolerDone := pooler.Start(newKubernetesSource(logger, cfg), newCloudFoundrySink(logger, cfg))

	wg := sync.WaitGroup{}

	wg.Add(1)
	go gracefulExit(logger, &wg, poolerDone)
	wg.Wait()

	logger.Info("exiting")
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

func loadConfig(logger lager.Logger) *config.Config {
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("parsing config", err)
	}

	return cfg
}

func newKubernetesSource(logger lager.Logger, cfg *config.Config) route.Source {
	kubecfg, err := k8sclientcmd.BuildConfigFromFlags("", cfg.KubeConfigPath)
	if err != nil {
		logger.Fatal("building config from flags", err)
	}
	clientset, err := k8sclient.NewForConfig(kubecfg)
	if err != nil {
		logger.Fatal("creating clientset from kube config", err)
	}

	return kubernetes.New(clientset, cfg.CloudFoundryAPPDomainName)
}

func newCloudFoundrySink(logger lager.Logger, cfg *config.Config) route.Router {
	bus := messagebus.NewMessageBus(logger)
	err := bus.Connect(cfg.NatsServers)
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
	tcpRouter, err := tcp.NewRoutingApi(uaaClient, cfg.CloudFoundryAPIURL, cfg.SkipTLSVerification)
	if err != nil {
		logger.Fatal("creating TCP router", err)
	}

	return cloudfoundry.NewRouter(bus, tcpRouter)
}
