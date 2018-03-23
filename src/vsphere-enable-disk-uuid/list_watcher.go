package main

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeAddedCallback func(*v1.Node)

func ListWatch(k8s kubernetes.Interface, nodeAddedCallback NodeAddedCallback) error {
	watcher, err := k8s.CoreV1().Nodes().Watch(meta_v1.ListOptions{})
	if err != nil {
		return err
	}

	defer watcher.Stop()

	for {
		event, stillAlive := <-watcher.ResultChan()

		if !stillAlive {
			break
		}

		if event.Type == watch.Added {
			nodeAddedCallback(event.Object.(*v1.Node))
		}

	}

	return nil
}
