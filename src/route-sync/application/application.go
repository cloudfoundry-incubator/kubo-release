package application

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

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/route-registrar/messagebus"
)

//go:generate counterfeiter . Application

type Application struct {
}

func NewApp() Application {
	return Application{}
}

func (app *Application) Start(cfg *config.Config, logger lager.Logger) {
	mainPooler := pooler.ByTime(time.Duration(30*time.Second), logger)
	poolerDone := mainPooler.Start(NewKubernetesSource(logger, cfg), NewCloudFoundrySink(app, logger, cfg))

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

func NewKubernetesSource(logger lager.Logger, cfg *config.Config) route.Source {
	kubecfg, err := k8sclientcmd.BuildConfigFromFlags("", cfg.KubeConfigPath)
	if err != nil {
		logger.Fatal("building config from flags", err)
	}
	clientset, err := k8sclient.NewForConfig(kubecfg)
	if err != nil {
		logger.Fatal("creating clientset from kube config", err)
	}

	return kubernetes.NewSource(clientset, cfg.CloudFoundryAppDomainName)
}

func (app *Application) NewMessageBus(logger lager.Logger, cfg *config.Config) (messagebus.MessageBus, error) {
	bus := messagebus.NewMessageBus(logger)
	err := bus.Connect(cfg.NatsServers)
	if err != nil {
		logger.Fatal("connecting to NATS", err)
	}
	return bus, err
}

func (app *Application) NewTCPRouter(logger lager.Logger, cfg *config.Config) (tcp.Router, error) {
	//	uaaClient, err := uaa.NewClient(logger, cfg.UAAConfig(), clock.NewClock())
	//	if err != nil {
	//		logger.Fatal("creating UAA client", err)
	//	}
	//	tcpRouter, err := tcp.NewRoutingApi(uaaClient, cfg.RoutingApiUrl, cfg.SkipTLSVerification)
	//	if err != nil {
	//		logger.Fatal("creating TCP router", err)
	//	}
	//	return tcpRouter, err
}

func NewCloudFoundrySink(app *Application, logger lager.Logger, cfg *config.Config) route.Router {
	bus, err := app.NewMessageBus(logger, cfg)
	if err != nil {
		panic(err)
	}

	tcpRouter, err := app.NewTCPRouter(logger, cfg)
	if err != nil {
		panic(err)
	}

	return cloudfoundry.NewRouter(bus, tcpRouter)
}
