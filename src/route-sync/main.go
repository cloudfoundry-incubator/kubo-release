package main

import (
	"fmt"
	"route-sync/cloudfoundry"
	"route-sync/fixed_source"
	"route-sync/pooler"
	"route-sync/route"
	"time"
)

func main() {
	fmt.Print("starting route-sync\n")
	httpRoutes := []*route.HTTP{
		&route.HTTP{
			Name: "foo.bar.com",
			Backend: []route.Endpoint{
				route.Endpoint{
					IP:   "10.10.10.10",
					Port: 8080,
				},
			},
		},
	}

	// TODO: replace this with a kubernetes source
	src := fixed_source.New(nil, httpRoutes)

	// TODO: pass in a valid MessageBus here
	sink := cloudfoundry.NewSink(nil)

	pooler := pooler.ByTime(time.Duration(10 * time.Second))
	done, tick := pooler.Start(src, sink)

	fmt.Print("started, Ctrl+C to exit\n")
	for {
		<-tick
		fmt.Print("announced!\n")
	}
	done <- struct{}{}
}
