package kubernetes

import (
	"route-sync/config"
	"route-sync/route"

	"code.cloudfoundry.org/lager"
	k8sclient "k8s.io/client-go/kubernetes"
	k8sclientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/rest"
)

type SourceBuilder struct {
	buildConfig            func(string, string) (*rest.Config, error)
	newKubernetesClientSet func(*rest.Config) (*k8sclient.Clientset, error)
	newKubernetesSource    func(k8sclient.Interface, string) route.Source
}

func DefaultSourceBuilder() *SourceBuilder {
	return NewSourceBuilder(k8sclientcmd.BuildConfigFromFlags, k8sclient.NewForConfig, NewSource)
}

func NewSourceBuilder(buildConfig func(string, string) (*rest.Config, error),
	newKubernetesClientSet func(*rest.Config) (*k8sclient.Clientset, error),
	newKubernetesSource func(k8sclient.Interface, string) route.Source) *SourceBuilder {
	return &SourceBuilder{
		buildConfig: buildConfig,
		newKubernetesClientSet: newKubernetesClientSet,
		newKubernetesSource: newKubernetesSource,
	}
}

func (builder *SourceBuilder) CreateSource(cfg *config.Config, logger lager.Logger) route.Source {
	kubecfg, err := builder.buildConfig("", cfg.KubeConfigPath)
	if err != nil {
		logger.Fatal("building config from flags", err)
	}
	clientset, err := builder.newKubernetesClientSet(kubecfg)
	if err != nil {
		logger.Fatal("creating clientset from kube config", err)
	}

	return builder.newKubernetesSource(clientset, cfg.CloudFoundryAppDomainName)
}
