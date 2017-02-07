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
		requestChan := make(chan *http.Request)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			decoder := json.NewDecoder(r.Body)
			var tcpRoutes []map[string]interface{}
			err := decoder.Decode(&tcpRoutes)
			Expect(err).To(BeNil())

			Expect(tcpRoutes).To(HaveLen(2))

			tcpRoute := tcpRoutes[0]
			Expect(tcpRoute["backend_port"].(float64)).To(Equal(float64(5050)))
			Expect(tcpRoute["port"].(float64)).To(Equal(float64(1010)))
			Expect(tcpRoute["router_group_guid"].(string)).To(Equal("foobar"))
			Expect(tcpRoute["ttl"].(float64)).To(Equal(float64(60)))
			Expect(tcpRoute["backend_ip"].(string)).To(Equal("10.0.0.2"))

			go func() {
				requestChan <- r
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

		<-requestChan
		Expect(err).To(BeNil())
	})
})
