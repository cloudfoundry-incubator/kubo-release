package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	var executable string
	BeforeEach(func() {
		var err error
		executable, err = gexec.Build("vsphere-enable-disk-uuid")
		Expect(err).ToNot(HaveOccurred())
	})

	It("should fail if kubeconfig is not passed", func() {
		command := exec.Command(executable)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
	})

	It("should execute", func() {
		command := exec.Command(executable, "--kubeconfig", "fixtures/kubeconfig")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
	})
})
