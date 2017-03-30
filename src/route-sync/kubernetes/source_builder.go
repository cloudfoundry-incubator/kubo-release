package kubernetes

import (
	"route-sync/config"
	"route-sync/route"

	"code.cloudfoundry.org/lager"
	k8sclient "k8s.io/client-go/kubernetes"
	k8sclientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/rest"
)

type SourceBuildStrategy struct {
	buildConfig            func(string, string) (*rest.Config, error)
	newKubernetesClientSet func(*rest.Config) (*k8sclient.Clientset, error)
	newKubernetesSource    func(k8sclient.Interface, string) route.Source
}

type SourceBuilder struct {
	logger   lager.Logger
	strategy SourceBuildStrategy
}

func DefaultBuildStrategy() SourceBuildStrategy {
	return SourceBuildStrategy{k8sclientcmd.BuildConfigFromFlags, k8sclient.NewForConfig, NewSource }
}

func NewSourceBuilder(logger lager.Logger, strategy SourceBuildStrategy) *SourceBuilder {

	return &SourceBuilder{
		logger:   logger,
		strategy: strategy,
	}
}

func (builder *SourceBuilder) CreateSource(cfg *config.Config) route.Source {
	kubecfg, err := builder.strategy.buildConfig("", cfg.KubeConfigPath)
	if err != nil {
		builder.logger.Fatal("building config from flags", err)
	}
	clientset, err := builder.strategy.newKubernetesClientSet(kubecfg)
	if err != nil {
		builder.logger.Fatal("creating clientset from kube config", err)
	}

	return builder.strategy.newKubernetesSource(clientset, cfg.CloudFoundryAppDomainName)
}
