package cfcr_smoke_tests_test

import (
	"math/rand"
	"smoke-tests/runner"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestK8SCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CFCR Smoke-Tests Suite")
}

var (
	k8sRunner *runner.KubectlRunner
)

var _ = BeforeSuite(func() {
	rand.Seed(time.Now().UnixNano())
	k8sRunner = runner.NewKubectlRunner()
	k8sRunner.RunKubectlCommand("create", "namespace", k8sRunner.Namespace()).Wait("60s")
})

var _ = AfterSuite(func() {
	if k8sRunner != nil {
		k8sRunner.RunKubectlCommand("delete", "namespace", k8sRunner.Namespace()).Wait("60s")
	}
})
