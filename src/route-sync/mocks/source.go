package mocks

import "route-sync/route"

type Source struct {
	TCP_count int
	TCP_value []*route.TCP
}

func (s *Source) TCP() ([]*route.TCP, error) {
	s.TCP_count++
	return s.TCP_value, nil
}
