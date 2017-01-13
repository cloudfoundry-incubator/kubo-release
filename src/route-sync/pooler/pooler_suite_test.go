package pooler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPooler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pooler Suite")
}
