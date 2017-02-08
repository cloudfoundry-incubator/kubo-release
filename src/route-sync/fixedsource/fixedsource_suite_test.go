package fixedsource_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFixedsource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FixedSource Suite")
}
