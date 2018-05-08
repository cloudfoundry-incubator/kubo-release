package cfcr_smoke_tests_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var nginxSpec = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-smoke-test
spec:
  selector:
    matchLabels:
      app: nginx-smoke-test
  replicas: 2
  template:
    metadata:
      labels:
        app: nginx-smoke-test
    spec:
      containers:
      - name: nginx
        image: nginx:1.8 #TODO something that we upload in blobs
        ports:
        - containerPort: 80
`

var _ = Describe("CFCR Smoke Tests", func() {
	Context("When deploying workloads", func() {
		var file *os.File
		var err error

		BeforeEach(func() {
			file, err = ioutil.TempFile(os.TempDir(), "nginx-smoke-test")
			Expect(err).ToNot(HaveOccurred())
			Expect(ioutil.WriteFile(file.Name(), []byte(nginxSpec), 0644)).ToNot(HaveOccurred())

			s := k8sRunner.RunKubectlCommand("create", "-f", file.Name())
			Eventually(s, "60s").Should(gexec.Exit(0))

			rolloutWatch := k8sRunner.RunKubectlCommand("rollout", "status", "deployment/nginx-smoke-test", "-w")
			Eventually(rolloutWatch, "120s").Should(gexec.Exit(0))
		})

		AfterEach(func() {
			s := k8sRunner.RunKubectlCommand("delete", "-f", file.Name())
			Eventually(s, "60s").Should(gexec.Exit(0))
			os.Remove(file.Name())
		})

		It("shows the pods are healthy", func() {
			args := []string{"get", "pods", "-l", "app=nginx-smoke-test", "-o", "jsonpath={.items[:].status.phase}"}
			s := k8sRunner.RunKubectlCommand(args...)
			Eventually(s, "60s").Should(gexec.Exit(0))
			Expect(s.Out).To(gbytes.Say("Running Running"))
		})

		It("allows commands to be executed on a container", func() {
			getPodName := k8sRunner.RunKubectlCommand("get", "pods", "-o", "jsonpath={.items[0].metadata.name}")
			Eventually(getPodName, "15s").Should(gexec.Exit(0))
			podName := string(getPodName.Out.Contents())

			args := []string{"exec", podName, "--", "which", "nginx"}
			s := k8sRunner.RunKubectlCommand(args...)
			Eventually(s, "60s").Should(gexec.Exit(0))
			Expect(s.Out).To(gbytes.Say("/usr/sbin/nginx"))
		})

		It("allows access to pod logs", func() {
			getPodName := k8sRunner.RunKubectlCommand("get", "pods", "-o", "jsonpath={.items[0].metadata.name}")
			Eventually(getPodName, "15s").Should(gexec.Exit(0))
			podName := string(getPodName.Out.Contents())

			getLogs := k8sRunner.RunKubectlCommand("logs", podName)
			Eventually(getLogs, "15s").Should(gexec.Exit(0))
			logContent := string(getLogs.Out.Contents())
			// nginx pods do not log much, unless there is an error we should see an empty string as a result
			Expect(logContent).To(Equal(""))
		})

		Context("Data Encryption", func() {
			It("successfully encrypts the data in ETCD", func() {
			})
		})
	})
})
