package main

import (
	"flag"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig = flag.String("kubeconfig", "~/.kube/config", "absolute path to the kubeconfig file")
)

func main() {
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err)
	}

	services, err := clientset.CoreV1().Services("").List(v1.ListOptions{})
	println("Hi")

	if err != nil {
		panic(err)
	}

	fmt.Printf("There are %d services in the cluster\n", len(services.Items))

	for _, service := range services.Items {
		fmt.Printf("%q\n", service)
	}
}
