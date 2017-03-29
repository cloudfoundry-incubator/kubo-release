package application

import (
	"context"
	"os"
	"os/signal"

	"code.cloudfoundry.org/lager"
)

type AbortFunc func(ctx context.Context, cancelFunc context.CancelFunc, logger lager.Logger)

// Catch SIGINT (Ctrl+C) and tell pooler to quit
func InterruptWaitFunc(ctx context.Context, cancel context.CancelFunc, logger lager.Logger) {
	logger.Info("started, Ctrl+C to exit")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	for {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
			return
		}
	}
	logger.Info("recieved Ctrl+C, exiting")
}
