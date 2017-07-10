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

	expectedDeployments := k8sdeploy.DeploymentSet{
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

	actualDeployments := k8sdeploy.DeploymentSet{}
	for _, pod := range pods {
		if deploymentName, err := adapter.ExtractDeploymentName(pod); err != nil {
			log.Fatal(err.Error())
		} else {
			actualDeployments[deploymentName].AddPod(pod)
		}
	}

	discrepancies := k8sdeploy.Discrepancies(expectedDeployments, actualDeployments)
	if len(discrepancies) != 0 {
		log.Fatal("discrepancies found:\n- " + strings.Join(discrepancies, "\n- "))
	}
}