package cloudfoundry

import (
	"route-sync/route"

	"code.cloudfoundry.org/route-registrar/config"

	"github.com/cloudfoundry/route-registrar/messagebus"
)

const privateInstanceId = "kubo-route-sync"

type sink struct {
	bus messagebus.MessageBus
}

func NewSink(bus messagebus.MessageBus) route.Sink {
	return &sink{bus: bus}
}

func (ts *sink) TCP(routes []*route.TCP) error {
	return nil
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
