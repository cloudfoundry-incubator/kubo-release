package kubernetesdeployment_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestKubernetesdeployment(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubernetesdeployment Suite")
}
