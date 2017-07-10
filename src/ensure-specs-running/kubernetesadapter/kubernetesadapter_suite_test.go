package kubernetesadapter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestKubernetesadapter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubernetesadapter Suite")
}
