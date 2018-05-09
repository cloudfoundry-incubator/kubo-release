package cfcr_smoke_tests_test

import (
	"fmt"
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

func curlLater(endpoint string) func() (string, error) {
	return func() (string, error) {
		cmd := exec.Command("curl", "--head", endpoint)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
}

var _ = Describe("CFCR Smoke Tests", func() {
	Describe("System Compenents", func() {
		It("should be healthy", func() {
			command := exec.Command("kubectl", "get", "componentstatuses", "-o", "jsonpath={.items[*].conditions[*].type}")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "60s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(Equal("Healthy Healthy Healthy"))
		})
	})

	Context("Deployment", func() {
		var dName string

		BeforeEach(func() {
			dName = randSeq(10)

			args := []string{"run", dName, "--image=nginx:1.13-alpine", "--image-pull-policy=Never", "-l", "app=" + dName}
			session := k8sRunner.RunKubectlCommand(args...)
			Eventually(session, "60s").Should(gexec.Exit(0))

			exposeArgs := []string{"expose", "deployment", dName, "--port=80", "--type=NodePort"}
			session = k8sRunner.RunKubectlCommand(exposeArgs...)
			Eventually(session, "120s").Should(gexec.Exit(0))

			rolloutWatch := k8sRunner.RunKubectlCommand("rollout", "status", "deployment/"+dName, "-w")
			Eventually(rolloutWatch, "120s").Should(gexec.Exit(0))
		})

		AfterEach(func() {
			session := k8sRunner.RunKubectlCommand("delete", "deployment", dName)
			Eventually(session, "60s").Should(gexec.Exit(0))
		})

		It("shows the pods are healthy", func() {
			args := []string{"get", "pods", "-l", "app=" + dName, "-o", "jsonpath={.items[:].status.phase}"}
			session := k8sRunner.RunKubectlCommand(args...)
			Eventually(session, "60s").Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("Running"))
		})

		It("allows commands to be executed on a container", func() {
			args := []string{"get", "pods", "-l", "app=" + dName, "-o", "jsonpath={.items[0].metadata.name}"}
			session := k8sRunner.RunKubectlCommand(args...)
			Eventually(session, "15s").Should(gexec.Exit(0))
			podName := string(session.Out.Contents())

			execArgs := []string{"exec", podName, "--", "which", "nginx"}
			execSession := k8sRunner.RunKubectlCommand(execArgs...)
			Eventually(execSession, "60s").Should(gexec.Exit(0))
			Expect(execSession.Out).To(gbytes.Say("/usr/sbin/nginx"))
		})

		It("allows access to pod logs", func() {
			args := []string{"get", "pods", "-l", "app=" + dName, "-o", "jsonpath={.items[0].metadata.name}"}
			session := k8sRunner.RunKubectlCommand(args...)
			Eventually(session, "15s").Should(gexec.Exit(0))
			podName := string(session.Out.Contents())

			session = k8sRunner.RunKubectlCommand("get", "nodes", "-o", "jsonpath={.items[0].status.addresses[?(@.type == \"InternalIP\")].address}")
			Eventually(session).Should(gexec.Exit(0))
			nodeIP := session.Out.Contents()

			session = k8sRunner.RunKubectlCommand("get", "svc", dName, "-o", "jsonpath={.spec.ports[0].nodePort}")
			Eventually(session).Should(gexec.Exit(0))
			port := session.Out.Contents()

			_, err := curlLater(fmt.Sprintf("http://%s:%s", nodeIP, port))()
			Expect(err).ToNot(HaveOccurred())

			getLogs := k8sRunner.RunKubectlCommand("logs", podName)
			Eventually(getLogs, "15s").Should(gexec.Exit(0))
			logContent := string(getLogs.Out.Contents())

			Expect(logContent).To(ContainSubstring("curl"))
		})

		Context("Port Forwarding", func() {
			var cmd *gexec.Session
			var port = "57869"

			BeforeEach(func() {
				args := []string{"get", "pods", "-l", "app=" + dName, "-o", "jsonpath={.items[0].metadata.name}"}
				session := k8sRunner.RunKubectlCommand(args...)
				Eventually(session, "15s").Should(gexec.Exit(0))
				podName := string(session.Out.Contents())

				args = []string{"port-forward", podName, port + ":80"}
				cmd = k8sRunner.RunKubectlCommand(args...)
			})

			AfterEach(func() {
				cmd.Terminate().Wait("15s")
			})

			It("successfully curls the nginx service", func() {
				Eventually(curlLater("http://localhost:"+port), "15s").Should(ContainSubstring("Server: nginx"))
			})
		})
	})
})
