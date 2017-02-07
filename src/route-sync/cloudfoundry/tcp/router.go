package router

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"route-sync/route"
)

type RouterGroup struct {
	Guid            string
	Name            string
	ReservablePorts string `json:"reservable_ports"`
	Type            string
}

type TCPRouter interface {
	RouterGroups() ([]RouterGroup, error)
	CreateRoutes(RouterGroup, []route.TCP) error
}

type routing_api_router struct {
	uaaOathToken       string
	cloudFoundryApiUrl string
}

func NewRoutingApi(uaaOathToken string, cloudFoundryApiUrl string) (TCPRouter, error) {
	if uaaOathToken == "" {
		return nil, fmt.Errorf("uaaOathToken token requried")
	}
	if cloudFoundryApiUrl == "" {
		return nil, fmt.Errorf("cloudFoundryApiUrl required")
	}

	return &routing_api_router{uaaOathToken: uaaOathToken, cloudFoundryApiUrl: cloudFoundryApiUrl}, nil
}

func (r *routing_api_router) buildRequest(verb string, path string) (*http.Request, *http.Client, error) {
	client := &http.Client{}
	req, err := http.NewRequest(verb, fmt.Sprintf("%s/%s", r.cloudFoundryApiUrl, path), nil)

	if err != nil {
		return req, client, fmt.Errorf("routing_api_router: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", r.uaaOathToken))

	return req, client, nil
}

func (r *routing_api_router) RouterGroups() ([]RouterGroup, error) {
	var routerGroups []RouterGroup

	req, client, err := r.buildRequest("GET", "/routing/v1/router_groups")
	if err != nil {
		return routerGroups, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return routerGroups, fmt.Errorf("routing_api_router: performing request: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return routerGroups, fmt.Errorf("routing_api_router: reading response %v", err)
	}

	err = json.Unmarshal(body, &routerGroups)
	if err != nil {
		err = fmt.Errorf("routing_api_router: unmarshalling response: %v\n body: %s", err, body)
	}

	return routerGroups, err
}

func (r *routing_api_router) CreateRoutes(RouterGroup, []route.TCP) error {
	req, client, err := r.buildRequest("POST", "/routing/v1/tcp_routes/create")
	if err != nil {
		return err
	}

	_, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("routing_api_router: %v", err)
	}
	return fmt.Errorf("NYI")
}
