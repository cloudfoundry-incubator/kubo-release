package cfcr_smoke_tests_test

import (
	"math/rand"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var letters = []rune("abcdefghi")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var _ = Describe("CFCR Smoke Tests", func() {
	Context("When deploying workloads", func() {
		var dName string

		BeforeEach(func() {
			dName = randSeq(10)

			args := []string{"run", dName, "--image=nginx", "--replicas=2", "-l", "app=" + dName} //TODO. set specific image
			s := k8sRunner.RunKubectlCommand(args...)
			Eventually(s, "60s").Should(gexec.Exit(0))

			rolloutWatch := k8sRunner.RunKubectlCommand("rollout", "status", "deployment/"+dName, "-w")
			Eventually(rolloutWatch, "120s").Should(gexec.Exit(0))
		})

		AfterEach(func() {
			s := k8sRunner.RunKubectlCommand("delete", "deployment", dName)
			Eventually(s, "60s").Should(gexec.Exit(0))
		})

		It("shows the pods are healthy", func() {
			args := []string{"get", "pods", "-l", "app=" + dName, "-o", "jsonpath={.items[:].status.phase}"}
			s := k8sRunner.RunKubectlCommand(args...)
			Eventually(s, "60s").Should(gexec.Exit(0))
			Expect(s.Out).To(gbytes.Say("Running Running"))
		})

		It("allows commands to be executed on a container", func() {
			args := []string{"get", "pods", "-l", "app=" + dName, "-o", "jsonpath={.items[0].metadata.name}"}
			s := k8sRunner.RunKubectlCommand(args...)
			Eventually(s, "15s").Should(gexec.Exit(0))
			podName := string(s.Out.Contents())

			args1 := []string{"exec", podName, "--", "which", "nginx"}
			s1 := k8sRunner.RunKubectlCommand(args1...)
			Eventually(s1, "60s").Should(gexec.Exit(0))
			Expect(s1.Out).To(gbytes.Say("/usr/sbin/nginx"))
		})

		It("allows access to pod logs", func() {
			args := []string{"get", "pods", "-l", "app=" + dName, "-o", "jsonpath={.items[0].metadata.name}"}
			s := k8sRunner.RunKubectlCommand(args...)
			Eventually(s, "15s").Should(gexec.Exit(0))
			podName := string(s.Out.Contents())

			getLogs := k8sRunner.RunKubectlCommand("logs", podName)
			Eventually(getLogs, "15s").Should(gexec.Exit(0))
			logContent := string(getLogs.Out.Contents())
			// nginx pods do not log much, unless there is an error we should see an empty string as a result
			Expect(logContent).To(Equal(""))
		})

		Context("Port Forwarding", func() {
			var cmd *gexec.Session
			var port = "57869"

			BeforeEach(func() {
				args := []string{"get", "pods", "-l", "app=" + dName, "-o", "jsonpath={.items[0].metadata.name}"}
				s := k8sRunner.RunKubectlCommand(args...)
				Eventually(s, "15s").Should(gexec.Exit(0))
				podName := string(s.Out.Contents())

				args = []string{"port-forward", podName, port + ":80"}
				cmd = k8sRunner.RunKubectlCommand(args...)
			})

			AfterEach(func() {
				cmd.Terminate().Wait("15s")
			})

			It("successfully curls the nginx service", func() {
				curlNginx := func() (string, error) {
					cmd := exec.Command("curl", "--head", "http://127.0.0.1:"+port)
					out, err := cmd.CombinedOutput()
					return string(out), err
				}
				Eventually(curlNginx, "15s").Should(ContainSubstring("Server: nginx"))
			})
		})
	})
})
