package fixed_source

import "route-sync/route"

type fixed_source struct {
	tcpRoutes  []*route.TCP
	httpRoutes []*route.HTTP
}

func (fs *fixed_source) TCP() ([]*route.TCP, error) {
	return fs.tcpRoutes, nil
}

func (fs *fixed_source) HTTP() ([]*route.HTTP, error) {
	return fs.httpRoutes, nil
}

func New(tcpRoutes []*route.TCP, httpRoutes []*route.HTTP) route.Source {
	return &fixed_source{tcpRoutes: tcpRoutes, httpRoutes: httpRoutes}
}
