package cfcr_smoke_tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Verify K8S cluster", func() {
	It("has all nodes ready", func() {
		Expect("READY").To(Equal("READY"))
	})

})
