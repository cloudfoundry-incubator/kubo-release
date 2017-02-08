package route

// Endpoint defines a given TCP Endpoint
type Endpoint struct {
	IP   string
	Port int
}

// TCP is a route definition for TCP traffic
type TCP struct {
	Frontend int // This is a port
	Backend  []Endpoint
}

// HTTP is a route definition for HTTP traffic
type HTTP struct {
	Name    string
	Backend []Endpoint
}

// Source provides routes
type Source interface {
	TCP() ([]*TCP, error)
	HTTP() ([]*HTTP, error)
}

// Router consumes routes
type Router interface {
	TCP(routes []*TCP) error
	HTTP(routes []*HTTP) error
}
