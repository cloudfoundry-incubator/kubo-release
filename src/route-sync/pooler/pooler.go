package pooler

import (
	"route-sync/route"
	"time"
)

// Pooler is responsible for querying a route.Source and updating a route.Router
type Pooler interface {
	Start(route.Source, route.Router) (done chan<- struct{}, tick <-chan struct{})
	Running() bool
}

type timeBased struct {
	duration time.Duration
	running  bool
}

// ByTime returns a Pooler that refreshes every duration
func ByTime(duration time.Duration) Pooler {
	return &timeBased{duration: duration, running: false}
}

func (tb *timeBased) tick(src route.Source, router route.Router) {
	tcpRoutes, err := src.TCP()
	if err != nil {
		panic(err)
	}
	err = router.TCP(tcpRoutes)
	if err != nil {
		panic(err)
	}
	httpRoutes, err := src.HTTP()
	if err != nil {
		panic(err)
	}
	err = router.HTTP(httpRoutes)
	if err != nil {
		panic(err)
	}
}

func (tb *timeBased) Start(src route.Source, router route.Router) (chan<- struct{}, <-chan struct{}) {
	tick := make(chan struct{})
	done := make(chan struct{})
	go func() {
		tb.running = true
		for {
			select {
			case <-done:
				tb.running = false
				return
			default:
				tb.tick(src, router)
				go func() {
					tick <- struct{}{}
				}()
				time.Sleep(tb.duration)
			}
		}
	}()

	return done, tick
}

func (tb *timeBased) Running() bool {
	return tb.running
}
