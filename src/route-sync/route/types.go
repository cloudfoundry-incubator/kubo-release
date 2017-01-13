package route

import "net"

type Endpoint struct {
	IP   net.IP
	Port int
}

type TCP struct {
	Frontend Endpoint
	Backend  []Endpoint
}

type Source interface {
	TCP() ([]*TCP, error)
}

type Sink interface {
	TCP(routes []*TCP) error
}
