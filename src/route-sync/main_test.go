package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

    "github.com/onsi/gomega/gexec"
)

var _ = Describe("RouteSyncMain", func() {
    It("compiles", func() {
        _, err := gexec.Build("route-sync")
        Expect(err).NotTo(HaveOccurred())
        gexec.CleanupBuildArtifacts()
    })
})
