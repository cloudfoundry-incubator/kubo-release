package route

type Endpoint struct {
	IP   string
	Port int
}

type TCP struct {
	Frontend Endpoint
	Backend  []Endpoint
}

type HTTP struct {
	Name    string
	Backend []Endpoint
}

type Source interface {
	TCP() ([]*TCP, error)
	HTTP() ([]*HTTP, error)
}

type Sink interface {
	TCP(routes []*TCP) error
	HTTP(routes []*HTTP) error
}
