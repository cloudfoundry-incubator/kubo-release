package main

import (
    "fmt"
    "sort"
    "strings"

    corev1 "k8s.io/api/core/v1"
)

type deployment struct {
    replicas int
    containers []string
}

func (d deployment) addPod(pod corev1.Pod) {
    d.replicas++
    for _, container := range pod.Status.ContainerStatuses {
        if container.Ready { d.containers = append(d.containers, container.Name) }
    }
}

func discrepanciesForDeployment(name string, expected, actual deployment) []string {
    discrepancies := []string{}
    if expected.replicas != actual.replicas {
        discrepancies = append(discrepancies, fmt.Sprintf(
            "expected %d %s replicas, found %d",
            expected.replicas,
            name,
            actual.replicas,
        ))
    }
    if !match(expected.containers, actual.containers) {
        discrepancies = append(discrepancies, fmt.Sprintf(
            "expected %s to have [%s] containers running, but found [%s] containers running",
            name,
            strings.Join(expected.containers, ", "),
            strings.Join(actual.containers, ", "),
        ))
    }
    return discrepancies
}

func match(xs, ys []string) bool {
    if len(xs) != len(ys) { return false }

    sort.Strings(xs)
    sort.Strings(ys)
    for i, x := range xs {
        if x != ys[i] { return false }
    }

    return true
}
