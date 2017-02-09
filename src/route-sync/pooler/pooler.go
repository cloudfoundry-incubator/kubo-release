package pooler

import (
	"fmt"
	"route-sync/route"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
)

// Pooler is responsible for querying a route.Source and updating a route.Router
type Pooler interface {
	Start(route.Source, route.Router) (done chan<- struct{}, tick <-chan struct{})
	Running() bool
}

type timeBased struct {
	sync.Mutex
	duration time.Duration
	running  bool
	logger   lager.Logger
}

// ByTime returns a Pooler that refreshes every duration
func ByTime(duration time.Duration, logger lager.Logger) Pooler {
	return &timeBased{duration: duration, running: false, logger: logger}
}

func (tb *timeBased) tick(src route.Source, router route.Router) {
	tcpRoutes, err := src.TCP()
	if err != nil {
		tb.logger.Fatal("fetching TCP routes", err)
	}
	err = router.TCP(tcpRoutes)
	if err != nil {
		tb.logger.Fatal("posting TCP routes", err)
	}
	httpRoutes, err := src.HTTP()
	if err != nil {
		tb.logger.Fatal("fetching HTTP routes", err)
	}
	err = router.HTTP(httpRoutes)
	if err != nil {
		tb.logger.Fatal("posting HTTP routes", err)
	}
	tb.logger.Info("registered routes", lager.Data{
		"TCP":  fmt.Sprintf("%q", tcpRoutes),
		"HTTP": fmt.Sprintf("%q", httpRoutes),
	})
}

func (tb *timeBased) Start(src route.Source, router route.Router) (chan<- struct{}, <-chan struct{}) {
	tick := make(chan struct{})
	done := make(chan struct{})
	go func() {
		tb.setRunning(true)
		timer := time.Tick(1)

		for {
			select {
			case <-done:
				tb.setRunning(false)
				close(tick)
				return
			case <-timer:
				tb.tick(src, router)
				timer = time.Tick(tb.duration)
				go func() {
					tick <- struct{}{}
				}()
			}
		}
	}()

	return done, tick
}

func (tb *timeBased) Running() bool {
	tb.Lock()
	defer tb.Unlock()

	return tb.running
}

func (tb *timeBased) setRunning(to bool) {
	tb.Lock()
	defer tb.Unlock()

	tb.running = to
}
