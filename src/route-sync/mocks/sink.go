package mocks

import "route-sync/route"

type Sink struct {
	HTTP_count  int
	HTTP_values [][]*route.HTTP
	TCP_count   int
	TCP_values  [][]*route.TCP
}

func (a *Sink) TCP(val []*route.TCP) error {
	a.TCP_count++
	a.TCP_values = append(a.TCP_values, val)

	return nil
}

func (a *Sink) HTTP(val []*route.HTTP) error {
	a.HTTP_count++
	a.HTTP_values = append(a.HTTP_values, val)

	return nil
}
