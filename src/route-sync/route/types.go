package route

type Endpoint struct {
	IP   string
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
