package main

import (
    "log"
    "strings"
)

func main() {
    const (
        heapster  = "heapster"
        influxdb  = "monitoring-influxdb"
        dashboard = "kubernetes-dashboard"
        dns       = "kube-dns"
    )

    expectedDeployments := map[string]deployment{
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

    kubernetesClient, err := newKubernetesClient(
        "/var/vcap/jobs/kubeconfig/config/kubeconfig",
        "kube-system",
    )
    if err != nil {
        log.Fatal(err.Error())
    }

    pods, err := kubernetesClient.pods()
    if err != nil {
        log.Fatal(err.Error())
    }

    actualDeployments := map[string]deployment{}
    for _, pod := range pods {
        deployment, err := kubernetesClient.extractDeploymentName(pod)
        if err != nil {
            log.Fatal(err.Error())
        }

        switch deployment {
        case heapster, influxdb, dashboard, dns:
            actualDeployments[deployment].addPod(pod)
        default:
            // ignore deployments not managed by this BOSH release
        }
    }

    heapsterProblems := discrepanciesForDeployment(
        heapster,
        expectedDeployments[heapster],
        actualDeployments[heapster],
    )
    influxdbProblems := discrepanciesForDeployment(
        influxdb,
        expectedDeployments[influxdb],
        actualDeployments[influxdb],
    )
    dashboardProblems := discrepanciesForDeployment(
        dashboard,
        expectedDeployments[dashboard],
        actualDeployments[dashboard],
    )
    dnsProblems := discrepanciesForDeployment(
        dns,
        expectedDeployments[dns],
        actualDeployments[dns],
    )
    problems := append(heapsterProblems, append(influxdbProblems, append(dashboardProblems, dnsProblems...)...)...)
    if len(problems) != 0 {
        log.Fatal("problems found:\n- " + strings.Join(problems, "\n- "))
    }
}
