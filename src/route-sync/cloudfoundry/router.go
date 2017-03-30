package cloudfoundry

import (
	"route-sync/cloudfoundry/tcp"
	"route-sync/route"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/route-registrar/config"
	"code.cloudfoundry.org/route-registrar/messagebus"
	cfConfig "code.cloudfoundry.org/route-registrar/config"

	"errors"
)

const privateInstanceID = "kubo-route-sync"

type router struct {
	tcpRouter tcp.Router
	bus       messagebus.MessageBus
}

// NewRouter creates a new route.Router for CloudFoundry
//
// This object wraps the CloudFoundry HTTP (gorouter) and TCP (routing-api) routers
func NewRouter(bus messagebus.MessageBus, tcpRouter tcp.Router) route.Router {
	return &router{bus: bus, tcpRouter: tcpRouter}
}

func (ts *router) Connect(natsServers []cfConfig.MessageBusServer, logger lager.Logger) {
	err := ts.bus.Connect(natsServers)
	if err != nil {
		logger.Fatal("connecting to NATS", err)
	}
}

func (ts *router) TCP(routes []*route.TCP) error {
	routerGroup, err := ts.tcpRouterGroup()
	if err != nil {
		return err
	}

	return ts.tcpRouter.CreateRoutes(routerGroup, routes)
}

func (ts *router) tcpRouterGroup() (tcp.RouterGroup, error) {
	routerGroups, err := ts.tcpRouter.RouterGroups()

	if err != nil {
		return tcp.RouterGroup{}, err
	}

	if len(routerGroups) != 1 {
		return tcp.RouterGroup{}, errors.New("NYI: Multiple router groups not supported")
	}

	return routerGroups[0], nil
}

func (ts *router) HTTP(routes []*route.HTTP) error {
	for _, httpRoute := range routes {
		for _, endpoint := range httpRoute.Backends {
			err := ts.bus.SendMessage("router.register", endpoint.IP, config.Route{
				Port: int(endpoint.Port),
				URIs: []string{httpRoute.Name},
			}, privateInstanceID)

			if err != nil {
				return err
			}
		}
	}
	return nil
}
