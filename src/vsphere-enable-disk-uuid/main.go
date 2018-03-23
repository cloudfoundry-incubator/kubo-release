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

	kubeconfig, err := clientcmd.LoadFromFile(opts.Kubeconfig)
	if err != nil {
		fmt.Printf("Failed to load kubeconfig: %s", err.Error())
		os.Exit(1)
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*kubeconfig, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		fmt.Printf("Something is wrong with kubeconfig: %s", err.Error())
		os.Exit(1)
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		fmt.Printf("Something is wrong with kubeconfig: %s", err.Error())
		os.Exit(1)
	}

	ListWatch(clientSet, nil)
}
