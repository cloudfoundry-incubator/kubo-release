package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

    "github.com/onsi/gomega/gexec"
)

var _ = Describe("EnsureSpecsRunningMain", func() {
    It("compiles", func() {
        _, err := gexec.Build("ensure-specs-running")
        Expect(err).NotTo(HaveOccurred())
        gexec.CleanupBuildArtifacts()
    })
})
