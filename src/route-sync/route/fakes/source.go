package fakes

import (
	"route-sync/route"
	"sync"
)

type Source struct {
	sync.Mutex

	TCP_count  int
	TCP_value  []*route.TCP
	HTTP_count int
	HTTP_value []*route.HTTP
}

func (s *Source) TCP() ([]*route.TCP, error) {
	s.Lock()
	defer s.Unlock()

	s.TCP_count++
	return s.TCP_value, nil
}

func (s *Source) HTTP() ([]*route.HTTP, error) {
	s.Lock()
	defer s.Unlock()

	s.HTTP_count++
	return s.HTTP_value, nil
}
