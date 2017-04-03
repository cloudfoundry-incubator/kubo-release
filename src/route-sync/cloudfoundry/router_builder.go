package cloudfoundry

import (
	"route-sync/cloudfoundry/tcp"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/route-registrar/messagebus"
	uaa "code.cloudfoundry.org/uaa-go-client"
	uaaConfig "code.cloudfoundry.org/uaa-go-client/config"
	"route-sync/config"
	"route-sync/route"
)

type RouterBuilder struct {
	newUAAClient  func(lager.Logger, *uaaConfig.Config, clock.Clock) (uaa.Client, error)
	newTCPRouter  func(uaa.Client, string, bool) (tcp.Router, error)
	newMessageBus func(lager.Logger) messagebus.MessageBus
	newRouter     func(messagebus.MessageBus, tcp.Router) route.Router
}

func DefaultRouterBuilder() *RouterBuilder {
	return &RouterBuilder{
		newUAAClient:  uaa.NewClient,
		newTCPRouter:  tcp.NewRoutingApi,
		newMessageBus: messagebus.NewMessageBus,
		newRouter:     NewRouter,
	}
}

func NewRouterBuilder(
	newUAAClient func(lager.Logger, *uaaConfig.Config, clock.Clock) (uaa.Client, error),
	newTCPRouter func(uaa.Client, string, bool) (tcp.Router, error),
	newMessageBus func(lager.Logger) messagebus.MessageBus,
	newRouter func(messagebus.MessageBus, tcp.Router) route.Router) *RouterBuilder {
	return &RouterBuilder{
		newUAAClient:  newUAAClient,
		newTCPRouter:  newTCPRouter,
		newMessageBus: newMessageBus,
		newRouter:     newRouter,
	}
}

func (builder *RouterBuilder) CreateRouter(cfg *config.Config, logger lager.Logger) route.Router {
	return builder.newRouter(builder.createHTTPRouter(logger), builder.createTCPRouter(cfg, logger))
}

func (builder *RouterBuilder) createTCPRouter(cfg *config.Config, logger lager.Logger) tcp.Router {

	uaaClient, err := builder.newUAAClient(logger, cfg.UAAConfig(), clock.NewClock())
	if err != nil {
		logger.Fatal("creating UAA client", err)
	}
	tcpRouter, err := builder.newTCPRouter(uaaClient, cfg.RoutingAPIURL, cfg.SkipTLSVerification)
	if err != nil {
		logger.Fatal("creating TCP router", err)
	}
	return tcpRouter
}

func (builder *RouterBuilder) createHTTPRouter(logger lager.Logger) messagebus.MessageBus {
	return builder.newMessageBus(logger)
}
