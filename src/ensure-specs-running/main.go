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

	log.Println("creating kubernetes client")
	client, err := k8sadapter.DefaultClient("/var/vcap/jobs/kubeconfig/config/kubeconfig")
	if err != nil {
		log.Println("failed creating kubernetes client")
		log.Fatalln(err.Error())
	}
	log.Println("succeeded creating kubernetes client")
	adapter := k8sadapter.NewAdapter(client, "kube-system")

	log.Println("fetching pods")
	pods, err := adapter.Pods()
	if err != nil {
		log.Println("failed fetching pods")
		log.Fatalln(err.Error())
	}
	log.Printf("succeeded fetching %d pods\n", len(pods))

	log.Println("extracting deployment names for pods")
	actualDeployments := k8sdeploy.DeploymentSet{}
	for _, pod := range pods {
		if deploymentName, err := adapter.ExtractDeploymentName(pod); err != nil {
			log.Printf("failed extracting deployment name for pod %s\n", pod.ObjectMeta.Name)
			log.Fatalln(err.Error())
		} else {
			if _, ok := actualDeployments[deploymentName]; !ok {
				actualDeployments[deploymentName] = k8sdeploy.NewDeployment(0, []string{})
			}
			actualDeployments[deploymentName].AddPod(pod)
		}
	}
	log.Println("succeeded extracting deployment names for pods")

	discrepancies := k8sdeploy.Discrepancies(expectedDeployments, actualDeployments)
	if len(discrepancies) != 0 {
		log.Fatalln("discrepancies found:\n- " + strings.Join(discrepancies, "\n- "))
	}
	log.Println("no discrepancies found")
}
