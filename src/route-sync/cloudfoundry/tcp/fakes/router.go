package fakes

import (
	"route-sync/cloudfoundry/tcp"
	"route-sync/route"
)

type FakeRouter struct {
	RouterGroupsResult []tcp.RouterGroup
	RouterGroupsError  error

	CreateRoutesLastRouterGroup tcp.RouterGroup
	CreateRoutesLastRoutes      []*route.TCP
	CreateRoutesError           error
}

func (r *FakeRouter) RouterGroups() ([]tcp.RouterGroup, error) {
	return r.RouterGroupsResult, r.RouterGroupsError
}

func (r *FakeRouter) CreateRoutes(rg tcp.RouterGroup, routes []*route.TCP) error {
	r.CreateRoutesLastRouterGroup = rg
	r.CreateRoutesLastRoutes = routes

	return r.CreateRoutesError
}
