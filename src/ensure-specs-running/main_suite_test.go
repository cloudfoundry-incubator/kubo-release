package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEnsureSpecsRunningMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EnsureSpecsRunningMain Suite")
}
