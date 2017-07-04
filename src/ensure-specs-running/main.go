package main

import (
    "log"
    "fmt"
    "strings"
    "sort"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

type deployment struct {
    replicas int
    containers []string
}

type kuboDeployments struct {
    heapster, influxdb, dashboard, dns deployment
}

var expectedDeployments = kuboDeployments{
    heapster: deployment{
        replicas: 1,
        containers: []string{"heapster"},
    },
    influxdb: deployment{
        replicas: 1,
        containers: []string{"influxdb"},
    },
    dashboard: deployment{
        replicas: 1,
        containers: []string{"kubernetes-dashboard"},
    },
    dns: deployment{
        replicas: 1,
        containers: []string{"kubedns", "dnsmasq", "sidecar"},
    },
}

var actualDeployments = kuboDeployments{}

const (
    heapster = "heapster"
    influxdb = "influxdb"
    dashboard = "kubernetes-dashboard"
    dns = "kube-dns"
    unrecognized = "unrecognized"
)

func main() {
    config, err := clientcmd.BuildConfigFromFlags("", "/var/vcap/jobs/kubeconfig/config/kubeconfig")
    if err != nil {
        log.Fatal(err.Error())
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatal(err.Error())
    }

    pods, err := clientset.CoreV1().Pods("kube-system").List(metav1.ListOptions{})
    if err != nil {
        log.Fatal(err.Error())
    }

    for _, pod := range pods.Items {
        if deployment := extractDeployment(pod); deployment != unrecognized {
            runningContainers := extractRunningContainers(pod)
            switch deployment {
            case heapster:
                actualDeployments.heapster.replicas++
                actualDeployments.heapster.containers = append(actualDeployments.heapster.containers, runningContainers...)
            case influxdb:
                actualDeployments.influxdb.replicas++
                actualDeployments.influxdb.containers = append(actualDeployments.influxdb.containers, runningContainers...)
            case dashboard:
                actualDeployments.dashboard.replicas++
                actualDeployments.dashboard.containers = append(actualDeployments.dashboard.containers, runningContainers...)
            case dns:
                actualDeployments.dns.replicas++
                actualDeployments.dns.containers = append(actualDeployments.dns.containers, runningContainers...)
            }
        }
    }

    problems := []string{}
    if expectedDeployments.heapster.replicas != actualDeployments.heapster.replicas {
        problems = append(problems, fmt.Sprintf(
            "expected %d %s replicas, found %d",
            expectedDeployments.heapster.replicas,
            heapster,
            actualDeployments.heapster.replicas,
        ))
    }
    if !match(expectedDeployments.heapster.containers, actualDeployments.heapster.containers) {
        problems = append(problems, fmt.Sprintf(
            "expected %s to have [%s] containers running, but found [%s] containers running",
            heapster,
            strings.Join(expectedDeployments.heapster.containers, ", "),
            strings.Join(actualDeployments.heapster.containers, ", "),
        ))
    }
    if expectedDeployments.influxdb.replicas != actualDeployments.influxdb.replicas {
        problems = append(problems, fmt.Sprintf(
            "expected %d %s replicas, found %d",
            expectedDeployments.influxdb.replicas,
            influxdb,
            actualDeployments.influxdb.replicas,
        ))
    }
    if !match(expectedDeployments.influxdb.containers, actualDeployments.influxdb.containers) {
        problems = append(problems, fmt.Sprintf(
            "expected %s to have [%s] containers running, but found [%s] containers running",
            influxdb,
            strings.Join(expectedDeployments.influxdb.containers, ", "),
            strings.Join(actualDeployments.influxdb.containers, ", "),
        ))
    }
    if expectedDeployments.dashboard.replicas != actualDeployments.dashboard.replicas {
        problems = append(problems, fmt.Sprintf(
            "expected %d %s replicas, found %d",
            expectedDeployments.dashboard.replicas,
            dashboard,
            actualDeployments.dashboard.replicas,
        ))
    }
    if !match(expectedDeployments.dashboard.containers, actualDeployments.dashboard.containers) {
        problems = append(problems, fmt.Sprintf(
            "expected %s to have [%s] containers running, but found [%s] containers running",
            dashboard,
            strings.Join(expectedDeployments.dashboard.containers, ", "),
            strings.Join(actualDeployments.dashboard.containers, ", "),
        ))
    }
    if expectedDeployments.dns.replicas != actualDeployments.dns.replicas {
        problems = append(problems, fmt.Sprintf(
            "expected %d %s replicas, found %d",
            expectedDeployments.dns.replicas,
            dns,
            actualDeployments.dns.replicas,
        ))
    }
    if !match(expectedDeployments.dns.containers, actualDeployments.dns.containers) {
        problems = append(problems, fmt.Sprintf(
            "expected %s to have [%s] containers running, but found [%s] containers running",
            dns,
            strings.Join(expectedDeployments.dns.containers, ", "),
            strings.Join(actualDeployments.dns.containers, ", "),
        ))
    }

    if len(problems) != 0 {
        dumpPods(pods.Items)
        log.Fatal("problems found:\n- " + strings.Join(problems, "\n- "))
    }
}

func extractDeployment(pod corev1.Pod) string {
    if _, hasPodTemplateHash := pod.ObjectMeta.Labels["pod-template-hash"]; !hasPodTemplateHash {
        return unrecognized
    }

    numLabels := len(pod.ObjectMeta.Labels)
    task := pod.ObjectMeta.Labels["task"]
    k8sApp := pod.ObjectMeta.Labels["k8s-app"]
    version := pod.ObjectMeta.Labels["version"]

    if numLabels == 4 && task == "monitoring" && k8sApp == "heapster" && version == "v6" {
        return heapster
    }
    if numLabels == 3 && task == "monitoring" && k8sApp == "influxdb" {
        return influxdb
    }
    if numLabels == 2 && k8sApp == "kube-dns" {
        return dns
    }
    if numLabels == 2 && k8sApp == "kubernetes-dashboard" {
        return dashboard
    }

    return unrecognized
}

func extractRunningContainers(pod corev1.Pod) []string {
    runningContainers := []string{}
    for _, container := range pod.Status.ContainerStatuses {
        if container.Ready { runningContainers = append(runningContainers, container.Name)}
    }
    return runningContainers
}

func dumpPods(pods []corev1.Pod) {
    for _, pod := range pods {
        log.Println("Labels:\n", pod.ObjectMeta.Labels, "\n")
        log.Println("Containers:\n", pod.Status.ContainerStatuses, "\n")
    }
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
