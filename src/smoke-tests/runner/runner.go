package runner

import (
	"os/exec"

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
	r := &KubectlRunner{}
	r.namespace = "test-" + uuid.NewV4().String()
	r.Timeout = "60s"

	return r
}
func (runner KubectlRunner) RunKubectlCommand(args ...string) *gexec.Session {
	return runner.RunKubectlCommandInNamespace(runner.namespace, args...)
}

func (runner KubectlRunner) RunKubectlCommandInNamespace(namespace string, args ...string) *gexec.Session {
	newArgs := append([]string{"--namespace", namespace}, args...)
	command := exec.Command("kubectl", newArgs...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

	Expect(err).NotTo(HaveOccurred())
	return session
}
func (runner KubectlRunner) Namespace() string {
	return runner.namespace
}
