package pooler

import (
	"context"
	"fmt"
	"route-sync/route"
	"sync"
	"time"

	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter . Pooler

// Pooler is responsible for querying a route.Source and updating a route.Router
type Pooler interface {
	Run(context.Context, route.Source, route.Router)
}

type timeBased struct {
	sync.Mutex
	interval time.Duration
	logger   lager.Logger
}

// ByTime returns a Pooler that refreshes every interval
func ByTime(interval time.Duration, logger lager.Logger) Pooler {
	return &timeBased{interval: interval, logger: logger}
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

func (tb *timeBased) Run(ctx context.Context, src route.Source, router route.Router) {
	timer := time.Tick(1)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer:
			tb.tick(src, router)
			timer = time.Tick(tb.interval)
		}
	}
}
