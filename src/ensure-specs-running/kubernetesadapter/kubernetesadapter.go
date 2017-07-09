package kubernetesadapter

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type adapter struct {
	client kubernetes.Interface
	namespace string
}

func NewAdapter(client kubernetes.Interface, namespace string) *adapter {
    return &adapter{client: client, namespace: namespace}
}

func DefaultClient(configPath string) (kubernetes.Interface, error) {
	if config, err := clientcmd.BuildConfigFromFlags("", configPath); err != nil {
		return nil, err
	} else {
	    return kubernetes.NewForConfig(config)
    }
}

func (k *adapter) Pods() ([]corev1.Pod, error) {
	if podList, err := k.client.CoreV1().Pods(k.namespace).List(metav1.ListOptions{}); err != nil {
		return nil, err
	} else {
		return podList.Items, nil
	}
}

func (k *adapter) ExtractDeploymentName(pod corev1.Pod) (string, error) {
	podOwnerReferences := pod.ObjectMeta.OwnerReferences
	if len(podOwnerReferences) != 1 {
		return "", fmt.Errorf(
			"expected pod %s to have 1 owner, has %d",
			pod.ObjectMeta.Name,
			len(podOwnerReferences),
		)
	}

	owningReplicaSet, err := k.client.ExtensionsV1beta1().ReplicaSets(k.namespace).Get(
		podOwnerReferences[0].Name,
		metav1.GetOptions{},
	)
	if err != nil {
		return "", err
	}

	replicaSetOwnerReferences := owningReplicaSet.ObjectMeta.OwnerReferences
	if len(replicaSetOwnerReferences) != 1 {
		return "", fmt.Errorf(
			"expected replicaset %s to have 1 owner, has %d",
			owningReplicaSet.ObjectMeta.Name,
			len(replicaSetOwnerReferences),
		)
	}

	owningDeployment, err := k.client.ExtensionsV1beta1().Deployments(k.namespace).Get(
		replicaSetOwnerReferences[0].Name,
		metav1.GetOptions{},
	)
	if err != nil {
		return "", err
	}

	return owningDeployment.ObjectMeta.Name, nil
}
