package cfcr_smoke_tests_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Verify K8S cluster", func() {
	Context("when get the componentstatus", func() {
		It("has all components healthy", func() {
			command := exec.Command("kubectl", "get", "componentstatuses", "-o", "jsonpath={.items[*].conditions[*].type}")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, "60s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(Equal("Healthy Healthy Healthy"))
		})
	})
})
