package cloudfoundry

import (
	"fmt"
	"route-sync/cloudfoundry/tcp"
	"route-sync/route"

	"code.cloudfoundry.org/route-registrar/config"

	"github.com/cloudfoundry/route-registrar/messagebus"
)

const privateInstanceId = "kubo-route-sync"

type sink struct {
	tcpRouter tcp.Router
	bus       messagebus.MessageBus
}

func NewSink(bus messagebus.MessageBus, tcpRouter tcp.Router) route.Sink {
	return &sink{bus: bus, tcpRouter: tcpRouter}
}

func (ts *sink) TCP(routes []*route.TCP) error {
	routerGroup, err := ts.tcpRouterGroup()
	if err != nil {
		return err
	}

	return ts.tcpRouter.CreateRoutes(routerGroup, routes)
}

func (ts *sink) tcpRouterGroup() (tcp.RouterGroup, error) {
	routerGroups, err := ts.tcpRouter.RouterGroups()

	if err != nil {
		return tcp.RouterGroup{}, err
	}

	if len(routerGroups) != 1 {
		return tcp.RouterGroup{}, fmt.Errorf("NYI: Multiple router groups not supported")
	}

	return routerGroups[0], nil
}

func (ts *sink) HTTP(routes []*route.HTTP) error {
	for _, httpRoute := range routes {
		for _, endpoint := range httpRoute.Backend {
			ts.bus.SendMessage("router.register", endpoint.IP, config.Route{
				Port: endpoint.Port,
				URIs: []string{httpRoute.Name},
			}, privateInstanceId)
		}
	}
	return nil
}
