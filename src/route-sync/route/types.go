package route

import (
	"code.cloudfoundry.org/lager"
	cfconfig "code.cloudfoundry.org/route-registrar/config"
)

type Port int

// Endpoint defines a given TCP Endpoint
type Endpoint struct {
	IP   string
	Port Port
}

// TCP is a route definition for TCP traffic
type TCP struct {
	Frontend Port
	Backends []Endpoint
}

// HTTP is a route definition for HTTP traffic
type HTTP struct {
	Name     string
	Backends []Endpoint
}

//go:generate counterfeiter . Source

// Source provides routes
type Source interface {
	TCP() ([]*TCP, error)
	HTTP() ([]*HTTP, error)
}

//go:generate counterfeiter . Router

// Router consumes routes
type Router interface {
	Connect(natsServers []cfconfig.MessageBusServer, logger lager.Logger)
	TCP(routes []*TCP) error
	HTTP(routes []*HTTP) error
}
