package kubernetes

import (
	"errors"
	"route-sync/config"
	"route-sync/route"
	"route-sync/route/routefakes"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var _ = Describe("Source", func() {
	Context("SourceBuilder", func() {
		var (
			logger                     lager.Logger
			fakeSrc                    route.Source
			fakeKubernetesSource       func(k8sclient.Interface, string) route.Source
			fakeNewKubernetesClientSet func(*rest.Config) (*k8sclient.Clientset, error)
			fakeBuildConfig            func(string, string) (*rest.Config, error)
			cfg                        = &config.Config{
				RawNatsServers:            "[{\"Host\": \"10.0.1.8:4222\",\"User\": \"nats\", \"Password\": \"natspass\"}]",
				RoutingApiUrl:             "https://api.cf.example.org",
				CloudFoundryAppDomainName: "apps.cf.example.org",
				UAAApiURL:                 "https://uaa.cf.example.org",
				RoutingAPIUsername:        "routeUser",
				RoutingAPIClientSecret:    "aabbcc",
				SkipTLSVerification:       true,
				KubeConfigPath:            "~/.config/kube",
			}
			fakeBuildConfigCallCount            = 0
			fakeNewKubernetesClientSetCallCount = 0
		)
		BeforeEach(func() {
			logger = lagertest.NewTestLogger("")
			fakeSrc = &routefakes.FakeSource{}
			fakeBuildConfigCallCount = 0
			fakeNewKubernetesClientSetCallCount = 0
			fakeKubernetesSource = func(k8sclient.Interface, string) route.Source {
				return fakeSrc
			}
			fakeNewKubernetesClientSet = func(*rest.Config) (*k8sclient.Clientset, error) {
				fakeNewKubernetesClientSetCallCount++
				return nil, nil
			}
			fakeBuildConfig = func(string, string) (*rest.Config, error) {
				fakeBuildConfigCallCount++
				return nil, nil
			}
		})

		It("builds a config and returns a Source", func() {
			srcBuilder := NewSourceBuilder(logger, fakeBuildConfig, fakeNewKubernetesClientSet, fakeKubernetesSource)
			src := srcBuilder.CreateSource(cfg)
			Expect(fakeBuildConfigCallCount).To(Equal(1))
			Expect(fakeNewKubernetesClientSetCallCount).To(Equal(1))
			Expect(src).To(Equal(fakeSrc))
		})

		It("panics when there are errors in build config", func() {
			fakeBuildConfig = func(string, string) (*rest.Config, error) {
				fakeBuildConfigCallCount++
				return nil, errors.New("")
			}
			srcBuilder := NewSourceBuilder(logger, fakeBuildConfig, fakeNewKubernetesClientSet, fakeKubernetesSource)
			defer func() {
				recover()
				Eventually(logger).Should(gbytes.Say("building config from flags"))
			}()
			srcBuilder.CreateSource(cfg)
		})

		It("panics when there are errors in kubernetes client", func() {
			fakeNewKubernetesClientSet = func(*rest.Config) (*k8sclient.Clientset, error) {
				return nil, errors.New("")
			}
			srcBuilder := NewSourceBuilder(logger, fakeBuildConfig, fakeNewKubernetesClientSet, fakeKubernetesSource)
			defer func() {
				recover()
				Eventually(logger).Should(gbytes.Say("creating clientset from kube config"))
			}()
			srcBuilder.CreateSource(cfg)
		})
	})
})
