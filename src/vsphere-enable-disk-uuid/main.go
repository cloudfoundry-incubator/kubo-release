package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Options struct {
	Kubeconfig string `long:"kubeconfig" description:"Kubeconfig to connect to kuberenetes API server" required:"true"`
}

func main() {
	opts := Options{}
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	k8s, err := getK8sClient(opts)
	if err != nil {
		fmt.Printf("Failed to load kubeconfig: %s", err.Error())
		os.Exit(1)
	}

	ListWatch(k8s, nil)
}


func getK8sClient(opts Options) (kubernetes.Interface, error) {
	kubeconfig, err := clientcmd.LoadFromFile(opts.Kubeconfig)
	if err != nil {
		return nil, err
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*kubeconfig, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	k8s, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return k8s, nil
}
