package cloudfoundry

import (
	"route-sync/cloudfoundry/tcp"
	"route-sync/config"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/route-registrar/messagebus"
	uaa "code.cloudfoundry.org/uaa-go-client"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"
)

type CloudFoundryRoutingBuilder struct {
	logger lager.Logger
	config *config.Config
}

func NewCloudFoundryRoutingBuilder(config *config.Config, logger lager.Logger) *CloudFoundryRoutingBuilder {
	return &CloudFoundryRoutingBuilder{
		logger: logger,
		config: config,
	}
}

func (builder *CloudFoundryRoutingBuilder) CreateTCPRouter(
	newUAAClient func(logger lager.Logger, cfg *uaaconfig.Config, clock clock.Clock) (uaa.Client, error),
	newTCPRouter func(uaaClient uaa.Client, routingApiUrl string, skipTlsVerification bool) (tcp.Router, error)) tcp.Router {

	uaaClient, err := newUAAClient(builder.logger, builder.config.UAAConfig(), clock.NewClock())
	if err != nil {
		builder.logger.Fatal("creating UAA client", err)
	}
	tcpRouter, err := newTCPRouter(uaaClient, builder.config.RoutingApiUrl, builder.config.SkipTLSVerification)
	if err != nil {
		builder.logger.Fatal("creating TCP router", err)
	}
	return tcpRouter
}

func (builder *CloudFoundryRoutingBuilder) CreateHTTPRouter(newMessageBus func(logger lager.Logger) messagebus.MessageBus) messagebus.MessageBus {
	return newMessageBus(builder.logger)
}
