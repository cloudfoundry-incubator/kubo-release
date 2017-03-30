package tcp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"route-sync/route"

	"code.cloudfoundry.org/uaa-go-client"
	"errors"
)

type RouterGroup struct {
	Guid            string
	Name            string
	ReservablePorts string `json:"reservable_ports"`
	Type            string
}

type Router interface {
	RouterGroups() ([]RouterGroup, error)
	CreateRoutes(RouterGroup, []*route.TCP) error
}

const routeTTL = 60

type routing_api_router struct {
	uaaClient           uaa_go_client.Client
	routingApiUrl       string
	skipTlsVerification bool
}

func NewRoutingApi(uaaClient uaa_go_client.Client, routingApiUrl string, skipTlsVerification bool) (Router, error) {
	if uaaClient == nil {
		return nil, errors.New("uaaClient token requried")
	}
	if routingApiUrl == "" {
		return nil, errors.New("routingApiUrl required")
	}

	return &routing_api_router{uaaClient: uaaClient, routingApiUrl: routingApiUrl, skipTlsVerification: skipTlsVerification}, nil
}

func (router *routing_api_router) buildRequest(verb string, path string) (*http.Request, *http.Client, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: router.skipTlsVerification},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(verb, fmt.Sprintf("%s/%s", router.routingApiUrl, path), nil)

	if err != nil {
		return req, client, fmt.Errorf("routing_api_router: %v", err)
	}

	token, err := router.uaaClient.FetchToken(false)
	if err != nil {
		return req, client, fmt.Errorf("routing_api_router: %v", err)
	}

	if token == nil {
		return req, client, errors.New("routing_api_router: nil token from uaaClient")
	}

	key := token.AccessToken
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", key))

	return req, client, nil
}

func (router *routing_api_router) RouterGroups() ([]RouterGroup, error) {
	var routerGroups []RouterGroup

	req, client, err := router.buildRequest("GET", "routing/v1/router_groups")
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

func (router *routing_api_router) CreateRoutes(rg RouterGroup, inputRoutes []*route.TCP) error {
	req, client, err := router.buildRequest("POST", "routing/v1/tcp_routes/create")
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	outputRoutes := router.buildRoutes(rg, inputRoutes)
	outputRoutesJson, err := json.Marshal(outputRoutes)
	if err != nil {
		err = fmt.Errorf("routing_api_router: marshalling request: %v", err)
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(outputRoutesJson))

	_, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("routing_api_router: %v", err)
	}
	return nil
}

func (router *routing_api_router) buildRoutes(routerGroup RouterGroup, inputRoutes []*route.TCP) []map[string]interface{} {
	result := []map[string]interface{}{}

	for _, inputRoute := range inputRoutes {
		for _, backend := range inputRoute.Backends {
			outputRoute := map[string]interface{}{
				"router_group_guid": routerGroup.Guid,
				"port":              inputRoute.Frontend,
				"ttl":               routeTTL,
				"backend_ip":        backend.IP,
				"backend_port":      backend.Port,
			}

			result = append(result, outputRoute)
		}
	}

	return result
}
