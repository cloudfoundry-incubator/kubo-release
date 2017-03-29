package application

import (
	"route-sync/cloudfoundry"
	"route-sync/cloudfoundry/tcp"
	"route-sync/config"
	"route-sync/kubernetes"
	"route-sync/route"

	uaa "code.cloudfoundry.org/uaa-go-client"
	k8sclient "k8s.io/client-go/kubernetes"
	k8sclientcmd "k8s.io/client-go/tools/clientcmd"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/route-registrar/messagebus"
)

func GetKubernetesSource(cfg *config.Config, logger lager.Logger) route.Source {
	srcBuilder := kubernetes.NewSourceBuilder(logger, k8sclientcmd.BuildConfigFromFlags, k8sclient.NewForConfig, kubernetes.NewSource)
	return srcBuilder.GetSource(cfg)
}

func GetCloudFoundryRouter(cfg *config.Config, logger lager.Logger) route.Router {
	routerBuilder := cloudfoundry.NewCloudFoundryRoutingBuilder(cfg, logger)
	return cloudfoundry.NewRouter(routerBuilder.CreateHTTPRouter(messagebus.NewMessageBus), routerBuilder.CreateTCPRouter(uaa.NewClient, tcp.NewRoutingApi))
}
