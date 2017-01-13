package fixed_source_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFixedSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FixedSource Suite")
}
