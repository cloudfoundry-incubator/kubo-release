package router_test

import (
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
})
