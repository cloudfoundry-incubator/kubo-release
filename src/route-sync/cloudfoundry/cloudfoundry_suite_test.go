package cloudfoundry_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCloudfoundry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cloudfoundry Suite")
}
