package cfcr_smoke_tests_test

import (
	"os/exec"
	"testing"

	uuid "github.com/satori/go.uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type KubectlRunner struct {
	configPath string
	namespace  string
	Timeout    string
}

func NewKubectlRunner() *KubectlRunner {
	runner := &KubectlRunner{}
	runner.namespace = "test-" + uuid.NewV4().String()
	runner.Timeout = "60s"

	return runner
}
func (runner KubectlRunner) RunKubectlCommand(args ...string) *gexec.Session {
	return runner.RunKubectlCommandInNamespace(runner.namespace, args...)
}

func (runner KubectlRunner) RunKubectlCommandInNamespace(namespace string, args ...string) *gexec.Session {
	newArgs := append([]string{"--kubeconfig", runner.configPath, "--namespace", namespace}, args...)
	command := exec.Command("kubectl", newArgs...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

	Expect(err).NotTo(HaveOccurred())
	return session
}
func (runner KubectlRunner) Namespace() string {
	return runner.namespace
}

func TestPodLogs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PodLogs Suite")
}

var (
	runner *KubectlRunner
)

var _ = BeforeSuite(func() {
	runner = NewKubectlRunner()
	runner.RunKubectlCommand("create", "namespace", runner.Namespace()).Wait("60s")
})

var _ = AfterSuite(func() {
	if runner != nil {
		runner.RunKubectlCommand("delete", "namespace", runner.Namespace()).Wait("60s")
	}
})
