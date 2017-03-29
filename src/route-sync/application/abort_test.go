package application_test

import (
	"context"
	. "route-sync/application"
	"sync"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Abort", func() {
	var (
		logger lager.Logger
	)
	BeforeEach(func() {
		logger = lagertest.NewTestLogger("")
	})
	Context("InterruptWaitFunc", func() {
		It("exits via context", func() {
			ctx, cancelFunc := context.WithCancel(context.Background())

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				InterruptWaitFunc(ctx, nil, logger)
				wg.Done()
			}()
			cancelFunc()
			wg.Wait()
		})
	})
})
