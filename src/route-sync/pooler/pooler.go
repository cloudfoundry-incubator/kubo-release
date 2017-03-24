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
	Start(route.Source, route.Router) (done chan<- struct{})
	Running() bool
}

type timeBased struct {
	sync.Mutex
	interval time.Duration
	running  bool
	logger   lager.Logger
}

// ByTime returns a Pooler that refreshes every interval
func ByTime(interval time.Duration, logger lager.Logger) Pooler {
	return &timeBased{interval: interval, running: false, logger: logger}
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

func (tb *timeBased) Start(src route.Source, router route.Router) chan<- struct{} {
	done := make(chan struct{})
	go func() {
		tb.setRunning(true)
		timer := time.Tick(1)

		for {
			select {
			case <-done:
				tb.setRunning(false)
				return
			case <-timer:
				tb.tick(src, router)
				timer = time.Tick(tb.interval)
			}
		}
	}()

	return done
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
