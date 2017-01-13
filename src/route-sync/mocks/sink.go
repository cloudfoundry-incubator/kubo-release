package mocks

import "route-sync/route"

type Sink struct {
	TCP_count  int
	TCP_values [][]*route.TCP
}

func (a *Sink) TCP(val []*route.TCP) error {
	a.TCP_count++
	a.TCP_values = append(a.TCP_values, val)

	return nil
}
