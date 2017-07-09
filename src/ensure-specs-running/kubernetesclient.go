package main

import (
    "fmt"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

type kubernetesClient struct {
    clientset *kubernetes.Clientset
    namespace string
}

func newKubernetesClient(configPath, namespace string) (*kubernetesClient, error) {
    config, err := clientcmd.BuildConfigFromFlags("", configPath)
    if err != nil {
        return nil, err
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }

    return &kubernetesClient{
        clientset: clientset,
        namespace: namespace,
    }, nil
}

func (k *kubernetesClient) pods() ([]corev1.Pod, error) {
    if podList, err := k.clientset.CoreV1().Pods(k.namespace).List(metav1.ListOptions{}); err != nil {
        return nil, err
    } else {
        return podList.Items, nil
    }
}

func (k *kubernetesClient) extractDeploymentName(pod corev1.Pod) (string, error) {
    podOwnerReferences := pod.ObjectMeta.OwnerReferences
    if len(podOwnerReferences) != 1 {
        return "", fmt.Errorf(
            "expected pod %s to have 1 owner, has %d",
            pod.ObjectMeta.Name,
            len(podOwnerReferences),
        )
    }

    owningReplicaSet, err := k.clientset.ExtensionsV1beta1().ReplicaSets(k.namespace).Get(
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

    owningDeployment, err := k.clientset.ExtensionsV1beta1().Deployments(k.namespace).Get(
        replicaSetOwnerReferences[0].Name,
        metav1.GetOptions{},
    )
    if err != nil {
        return "", err
    }

    return owningDeployment.ObjectMeta.Name, nil
}
