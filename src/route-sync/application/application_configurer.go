package application

import (
	"route-sync/cloudfoundry"
	"route-sync/cloudfoundry/tcp"
	"route-sync/config"
	"route-sync/route"

	uaa "code.cloudfoundry.org/uaa-go-client"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/route-registrar/messagebus"
)

func GetCloudFoundryRouter(cfg *config.Config, logger lager.Logger) route.Router {
	routerBuilder := cloudfoundry.NewCloudFoundryRoutingBuilder(cfg, logger)
	return cloudfoundry.NewRouter(routerBuilder.CreateHTTPRouter(messagebus.NewMessageBus), routerBuilder.CreateTCPRouter(uaa.NewClient, tcp.NewRoutingApi))
}
