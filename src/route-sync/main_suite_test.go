package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRouteSyncMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RouteSyncMain Suite")
}
