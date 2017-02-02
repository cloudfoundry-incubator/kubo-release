package mocks

import "route-sync/route"

type Source struct {
	TCP_count  int
	TCP_value  []*route.TCP
	HTTP_count int
	HTTP_value []*route.HTTP
}

func (s *Source) TCP() ([]*route.TCP, error) {
	s.TCP_count++
	return s.TCP_value, nil
}

func (s *Source) HTTP() ([]*route.HTTP, error) {
	s.HTTP_count++
	return s.HTTP_value, nil
}
