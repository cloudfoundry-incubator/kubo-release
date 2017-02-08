package fixedsource

import "route-sync/route"

type fixedSource struct {
	tcpRoutes  []*route.TCP
	httpRoutes []*route.HTTP
}

func (fs *fixedSource) TCP() ([]*route.TCP, error) {
	return fs.tcpRoutes, nil
}

func (fs *fixedSource) HTTP() ([]*route.HTTP, error) {
	return fs.httpRoutes, nil
}

// New returns a route.Source that always returns the given tcpRoutes, httpRoutes
func New(tcpRoutes []*route.TCP, httpRoutes []*route.HTTP) route.Source {
	return &fixedSource{tcpRoutes: tcpRoutes, httpRoutes: httpRoutes}
}
