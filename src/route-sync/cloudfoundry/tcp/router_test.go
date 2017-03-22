package tcp_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	. "route-sync/cloudfoundry/tcp"
	"route-sync/route"

	"code.cloudfoundry.org/uaa-go-client"
	uaafakes "code.cloudfoundry.org/uaa-go-client/fakes"
	uaaschema "code.cloudfoundry.org/uaa-go-client/schema"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("routing-api TCP router", func() {
	var (
		fooClient uaafakes.FakeClient
	)
	BeforeEach(func() {
		fooClient = uaafakes.FakeClient{}
		fooClient.FetchTokenReturns(&uaaschema.Token{
			AccessToken: "foo",
			ExpiresIn:   0,
		}, nil)
	})

	It("requires UAA client and an API endpoint", func() {
		var invalidCreations = []struct {
			uaaClient uaa_go_client.Client
			api       string
		}{
			{nil, ""},
			{&fooClient, ""},
			{nil, "bar"},
		}

		for _, t := range invalidCreations {
			Context(fmt.Sprintf("uaaClient: %q, api: %s", t.uaaClient, t.api), func() {
				router, err := NewRoutingApi(t.uaaClient, t.api, false)
				Expect(router).To(BeNil())
				Expect(err).NotTo(BeNil())
			})
		}
	})

	Context("With a valid UAA token", func() {
		var (
			requests chan *http.Request
			server   *httptest.Server
			router   Router
		)

		BeforeEach(func() {
			requests = make(chan *http.Request)
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				go func() {
					requests <- r
				}()
			}))
			router, _ = NewRoutingApi(&fooClient, server.URL, false)
		})

		AfterEach(func() {
			close(requests)
			server.Close()
		})

		It("posts the token with RouterGroups", func() {
			router.RouterGroups()
			req := <-requests
			Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{"bearer foo"}))
		})

		It("posts the token with CreateRoutes", func() {
			router.CreateRoutes(RouterGroup{}, []*route.TCP{&route.TCP{}})
			req := <-requests
			Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{"bearer foo"}))
		})
	})

	Context("RouterGroups", func() {
		It("parses a single router routing group", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				w.Write([]byte(`[{"guid":"abc123","name":"default-tcp","reservable_ports":"1024-65535","type":"tcp"}]`))
			}))
			defer ts.Close()

			router, _ := NewRoutingApi(&fooClient, ts.URL, false)

			routerGroups, err := router.RouterGroups()
			Expect(err).To(BeNil())

			Expect(routerGroups).To(ConsistOf(RouterGroup{
				Guid:            "abc123",
				Name:            "default-tcp",
				ReservablePorts: "1024-65535",
				Type:            "tcp",
			}))
		})
	})

	It("posts routes", func() {
		type expectedResponse struct {
			BackendPort     int    `json:"backend_port"`
			Port            int    `json:"port"`
			RouterGroupGuid string `json:"router_group_guid"`
			TTL             int    `json:"ttl"`
			BackendIp       string `json:"backend_ip"`
		}
		responseChan := make(chan []expectedResponse)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			decoder := json.NewDecoder(r.Body)

			var tcpRoutes []expectedResponse
			err := decoder.Decode(&tcpRoutes)
			Expect(err).To(BeNil())

			go func() {
				responseChan <- tcpRoutes
			}()
		}))
		defer ts.Close()

		router, _ := NewRoutingApi(&fooClient, ts.URL, false)

		routes := []*route.TCP{&route.TCP{
			Frontend: 1010,
			Backends: []route.Endpoint{
				route.Endpoint{
					IP:   "10.0.0.2",
					Port: 5050,
				},
				route.Endpoint{
					IP:   "10.0.0.3",
					Port: 2020,
				},
			},
		}}

		routerGroup := RouterGroup{}
		routerGroup.Guid = "foobar"
		err := router.CreateRoutes(routerGroup, routes)

		tcpRoutes := <-responseChan
		Expect(err).To(BeNil())

		Expect(tcpRoutes).To(HaveLen(2))
		Expect(tcpRoutes).To(ConsistOf(expectedResponse{
			BackendPort:     5050,
			Port:            1010,
			RouterGroupGuid: "foobar",
			TTL:             60,
			BackendIp:       "10.0.0.2",
		}, expectedResponse{
			BackendPort:     2020,
			Port:            1010,
			RouterGroupGuid: "foobar",
			TTL:             60,
			BackendIp:       "10.0.0.3",
		},
		))
	})
})
