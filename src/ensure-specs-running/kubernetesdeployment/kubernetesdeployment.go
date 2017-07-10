package kubernetesdeployment

import (
	"fmt"
	"sort"
	"strings"
	
	corev1 "k8s.io/api/core/v1"
)

type Deployment struct {
	replicas   int
	containers []string
}

func NewDeployment(replicas int, containers []string) *Deployment {
	return &Deployment{
		replicas:   replicas,
		containers: containers,
	}
}

func (d *Deployment) AddPod(pod corev1.Pod) {
	d.replicas++
	for _, container := range pod.Status.ContainerStatuses {
		if container.Ready {
			d.containers = append(d.containers, container.Name)
		}
	}
}

type DeploymentSet map[string]*Deployment

func Discrepancies(expected, actual DeploymentSet) []string {
	discrepancies := []string{}
	
	for name, expectedDeployment := range expected {
		if actualDeployment, ok := actual[name]; !ok {
			discrepancies = append(discrepancies, fmt.Sprintf("unable to find any pods for %s", name))
		} else {
			if actualDeployment.replicas != expectedDeployment.replicas {
				discrepancies = append(
					discrepancies, 
					fmt.Sprintf(
						"expected replica count for %s is %d, found %d",
						name,
						expectedDeployment.replicas,
						actualDeployment.replicas,
					),
				)
			}

			if !match(actualDeployment.containers, expectedDeployment.containers) {
				discrepancies = append(
					discrepancies,
					fmt.Sprintf(
						"expected ready containers [%s] for %s, found [%s]",
						strings.Join(expectedDeployment.containers, ", "),
						name,
						strings.Join(actualDeployment.containers, ", "),
					),
				)
			}
		}
	}

	return discrepancies
}

func match(xs, ys []string) bool {
    if len(xs) != len(ys) {
        return false
    }

    sort.Strings(xs)
    sort.Strings(ys)
    for i, x := range xs {
        if x != ys[i] {
            return false
        }
    }

    return true
}
