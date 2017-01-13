package pooler

import (
	"route-sync/route"
	"time"
)

type Pooler interface {
	Start(route.Source, route.Sink) chan<- struct{}
	Running() bool
}

type time_based struct {
	duration time.Duration
	running  bool
}

func ByTime(duration time.Duration) Pooler {
	return &time_based{duration: duration, running: false}
}

func (t *time_based) tick(src route.Source, sink route.Sink) {
	routes, err := src.TCP()
	if err != nil {
		panic(err)
	}
	err = sink.TCP(routes)
	if err != nil {
		panic(err)
	}
}

func (tb *time_based) Start(src route.Source, sink route.Sink) chan<- struct{} {
	done := make(chan struct{})
	go func() {
		tb.running = true
		for {
			select {
			case <-done:
				tb.running = false
				return
			default:
				tb.tick(src, sink)
				time.Sleep(tb.duration)
			}
		}
	}()

	return done
}

func (tb *time_based) Running() bool {
	return tb.running
}
