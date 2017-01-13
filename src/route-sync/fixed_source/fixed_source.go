package fixed_source

import "route-sync/route"

type fixed_source struct {
	tcp_routes []*route.TCP
}

func (fs *fixed_source) TCP() ([]*route.TCP, error) {
	return fs.tcp_routes, nil
}

func New(tcp_routes []*route.TCP) route.Source {
	return &fixed_source{tcp_routes}
}
