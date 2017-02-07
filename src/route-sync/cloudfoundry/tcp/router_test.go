package router_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	. "route-sync/cloudfoundry/tcp"
	"route-sync/route"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("routing-api TCP router", func() {
	It("requires credentials and an API endpoint", func() {
		var invalidCreations = []struct {
			creds string
			api   string
		}{
			{"", ""},
			{"foo", ""},
			{"", "bar"},
		}

		for _, t := range invalidCreations {
			Context(fmt.Sprintf("creds: %s, api: %s", t.creds, t.api), func() {
				router, err := NewRoutingApi(t.creds, t.api)
				Expect(router).To(BeNil())
				Expect(err).NotTo(BeNil())
			})
		}

		router, err := NewRoutingApi("foo", "foo")
		Expect(router).NotTo(BeNil())
		Expect(err).To(BeNil())
	})

	It("posts the UAA token during a request", func() {
		requestChan := make(chan *http.Request)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			go func() {
				requestChan <- r
			}()
		}))
		defer ts.Close()

		router, _ := NewRoutingApi("foobar", ts.URL)

		router.RouterGroups()
		req := <-requestChan
		Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{"bearer foobar"}))

		router.CreateRoutes(RouterGroup{}, []route.TCP{route.TCP{}})
		req = <-requestChan
		Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{"bearer foobar"}))

	})
	It("handles routing groups", func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			w.Write([]byte(`[{"guid":"abc123","name":"default-tcp","reservable_ports":"1024-65535","type":"tcp"}]`))
		}))
		defer ts.Close()

		router, _ := NewRoutingApi("foobar", ts.URL)

		routerGroups, err := router.RouterGroups()
		Expect(err).To(BeNil())

		Expect(routerGroups).To(ConsistOf(RouterGroup{
			Guid:            "abc123",
			Name:            "default-tcp",
			ReservablePorts: "1024-65535",
			Type:            "tcp",
		}))
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

		router, _ := NewRoutingApi("foobar", ts.URL)

		routes := []route.TCP{route.TCP{
			Frontend: 1010,
			Backend: []route.Endpoint{
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
