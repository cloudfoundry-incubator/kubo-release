package cfcr_smoke_tests_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var nginxSpec = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 2
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.8 #TODO something that we upload in blobs
        ports:
        - containerPort: 80
`

var _ = Describe("Deploy workload", func() {
	var file *os.File
	BeforeEach(func() {
		var err error
		file, err = ioutil.TempFile(os.TempDir(), "smoke-test-nginx")
		Expect(err).ToNot(HaveOccurred())
		Expect(ioutil.WriteFile(file.Name(), []byte(nginxSpec), 0644)).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		os.Remove(file.Name())
	})

	It("allows access to pod logs", func() {
		deployNginx := k8sRunner.RunKubectlCommand("create", "-f", file.Name())
		Eventually(deployNginx, "60s").Should(gexec.Exit(0))
		rolloutWatch := k8sRunner.RunKubectlCommand("rollout", "status", "deployment/nginx", "-w")
		Eventually(rolloutWatch, "120s").Should(gexec.Exit(0))

		getPodName := k8sRunner.RunKubectlCommand("get", "pods", "-o", "jsonpath={.items[0].metadata.name}")
		Eventually(getPodName, "15s").Should(gexec.Exit(0))
		podName := string(getPodName.Out.Contents())

		getLogs := k8sRunner.RunKubectlCommand("logs", podName)
		Eventually(getLogs, "15s").Should(gexec.Exit(0))
		logContent := string(getLogs.Out.Contents())
		// nginx pods do not log much, unless there is an error we should see an empty string as a result
		Expect(logContent).To(Equal(""))

		session := k8sRunner.RunKubectlCommand("delete", "-f", nginxSpec)
		session.Wait("60s")
	})

})
