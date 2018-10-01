package main

import (
	"flag"
	"fmt"
	"path/filepath"

	core "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var (
		kubeconfig *string
		nodeIP     *string
	)
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	nodeIP = flag.String("nodeip", "", "ip of the node being run against")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	nodesAPI := clientset.CoreV1().Nodes()
	nodelist, err := nodesAPI.List(metav1.ListOptions{LabelSelector: fmt.Sprintf("spec.ip=%s", *nodeIP)})
	if err != nil {
		panic(err)
	}
	node := &nodelist.Items[0]
	for i, condition := range node.Status.Conditions {
		if condition.Type == core.NodeNetworkUnavailable {
			condition.Status = core.ConditionFalse
			condition.Reason = "NetworkProvidedByFlannel"
			condition.Message = "Status manually modified by CFCR kubelet post-start"
			node.Status.Conditions[i] = condition
		}
	}

	_, err = nodesAPI.UpdateStatus(node)
	if err != nil {
		panic(err)
	}
}
