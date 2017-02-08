package fakes

import (
	"route-sync/route"
	"sync"
)

type Router struct {
	sync.Mutex
	HTTP_count  int
	HTTP_values [][]*route.HTTP
	TCP_count   int
	TCP_values  [][]*route.TCP
}

func (a *Router) TCP(val []*route.TCP) error {
	a.Lock()
	defer a.Unlock()

	a.TCP_count++
	a.TCP_values = append(a.TCP_values, val)

	return nil
}

func (a *Router) HTTP(val []*route.HTTP) error {
	a.Lock()
	defer a.Unlock()

	a.HTTP_count++
	a.HTTP_values = append(a.HTTP_values, val)

	return nil
}
