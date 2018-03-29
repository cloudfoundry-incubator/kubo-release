package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "vsphere-enable-disk-uuid"

	"sync"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	v1fakes "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	"k8s.io/client-go/testing"
)

type fakeWatchReactor struct {
	fakeWatch *watch.FakeWatcher
}

func (f fakeWatchReactor) Handles(action testing.Action) bool {
	return true
}

func (f fakeWatchReactor) React(action testing.Action) (bool, watch.Interface, error) {
	return true, f.fakeWatch, nil
}

type testState struct {
	wasCalled      bool
	callCountMutex sync.Mutex
	callCount      int
}

func (state *testState) fakeCallback() NodeAddedCallback {
	return func(n *v1.Node) {
		state.wasCalled = true
		state.callCountMutex.Lock()
		defer state.callCountMutex.Unlock()
		state.callCount = state.callCount + 1
	}
}

var _ = Describe("ListWatcher", func() {
	var (
		clientSet *fakeclientset.Clientset
		fakeNodes *v1fakes.FakeNodes
		state     *testState
		reactor   fakeWatchReactor
	)
	BeforeEach(func() {
		clientSet = fakeclientset.NewSimpleClientset()
		fakeNodes = clientSet.CoreV1().Nodes().(*v1fakes.FakeNodes)
		reactor = fakeWatchReactor{fakeWatch: watch.NewFake()}
		fakeNodes.Fake.Fake.WatchReactionChain = []testing.WatchReactor{reactor}
		state = &testState{}
	})

	It("should notify every time a node is added", func() {
		wg := listWatchWithFakeCallback(clientSet, state)

		reactor.fakeWatch.Add(&v1.Node{})
		reactor.fakeWatch.Add(&v1.Node{})
		reactor.fakeWatch.Stop()

		wg.Wait()

		Expect(state.wasCalled).To(Equal(true))
		Expect(state.callCount).To(Equal(2))
	})

	It("should not notify when a node is deleted", func() {
		wg := listWatchWithFakeCallback(clientSet, state)

		reactor.fakeWatch.Delete(&v1.Node{})
		reactor.fakeWatch.Stop()

		wg.Wait()

		Expect(state.wasCalled).To(Equal(false))
	})

})

func listWatchWithFakeCallback(clientSet kubernetes.Interface, state *testState) *sync.WaitGroup {
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		ListWatch(clientSet, state.fakeCallback())
	}()
	return &wg
}
