package main

import (
	"log"
	"strings"

	k8sadapter "ensure-specs-running/kubernetesadapter"
	k8sdeploy "ensure-specs-running/kubernetesdeployment"
)

func main() {
	const (
		heapster  = "heapster"
		influxdb  = "monitoring-influxdb"
		dashboard = "kubernetes-dashboard"
		dns       = "kube-dns"
	)

	expectedDeployments := map[string]k8sdeploy.Deployment{
		heapster:  k8sdeploy.NewDeployment(1, []string{"heapster"}),
		influxdb:  k8sdeploy.NewDeployment(1, []string{"influxdb"}),
		dashboard: k8sdeploy.NewDeployment(1, []string{"kubernetes-dashboard"}),
		dns:       k8sdeploy.NewDeployment(1, []string{"kubedns", "dnsmasq", "sidecar"}),
	}

    client, err := k8sadapter.DefaultClient("/var/vcap/jobs/kubeconfig/config/kubeconfig")
    if err != nil {
        log.Fatal(err.Error())
    }
    adapter := k8sadapter.NewAdapter(client, "kube-system")

	pods, err := adapter.Pods()
	if err != nil {
		log.Fatal(err.Error())
	}

	actualDeployments := map[string]k8sdeploy.Deployment{}
	for _, pod := range pods {
		deploymentName, err := adapter.ExtractDeploymentName(pod)
		if err != nil {
			log.Fatal(err.Error())
		}

		switch deploymentName {
		case heapster, influxdb, dashboard, dns:
			actualDeployments[deploymentName].AddPod(pod)
		default:
			// ignore deployments not managed by this BOSH release
		}
	}

	heapsterProblems := k8sdeploy.DiscrepanciesForDeployment(
		heapster,
		expectedDeployments[heapster],
		actualDeployments[heapster],
	)
	influxdbProblems := k8sdeploy.DiscrepanciesForDeployment(
		influxdb,
		expectedDeployments[influxdb],
		actualDeployments[influxdb],
	)
	dashboardProblems := k8sdeploy.DiscrepanciesForDeployment(
		dashboard,
		expectedDeployments[dashboard],
		actualDeployments[dashboard],
	)
	dnsProblems := k8sdeploy.DiscrepanciesForDeployment(
		dns,
		expectedDeployments[dns],
		actualDeployments[dns],
	)
	problems := append(heapsterProblems, append(influxdbProblems, append(dashboardProblems, dnsProblems...)...)...)
	if len(problems) != 0 {
		log.Fatal("problems found:\n- " + strings.Join(problems, "\n- "))
	}
}
