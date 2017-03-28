package kubernetes

import (
	"route-sync/config"
	"route-sync/route"

	"code.cloudfoundry.org/lager"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type SourceBuilder struct {
	logger                 lager.Logger
	buildConfig            func(string, string) (*rest.Config, error)
	newKubernetesClientSet func(*rest.Config) (*k8sclient.Clientset, error)
	newKubernetesSource    func(k8sclient.Interface, string) route.Source
}

func NewSourceBuilder(logger lager.Logger,
	buildConfig func(string, string) (*rest.Config, error),
	newKubernetesClientSet func(*rest.Config) (*k8sclient.Clientset, error),
	newKubernetesSource func(k8sclient.Interface, string) route.Source) *SourceBuilder {

	return &SourceBuilder{
		logger:                 logger,
		buildConfig:            buildConfig,
		newKubernetesClientSet: newKubernetesClientSet,
		newKubernetesSource:    newKubernetesSource,
	}
}

func (builder *SourceBuilder) GetSource(cfg *config.Config) route.Source {
	kubecfg, err := builder.buildConfig("", cfg.KubeConfigPath)
	if err != nil {
		builder.logger.Fatal("building config from flags", err)
	}
	clientset, err := builder.newKubernetesClientSet(kubecfg)
	if err != nil {
		builder.logger.Fatal("creating clientset from kube config", err)
	}

	return builder.newKubernetesSource(clientset, cfg.CloudFoundryAppDomainName)
}
