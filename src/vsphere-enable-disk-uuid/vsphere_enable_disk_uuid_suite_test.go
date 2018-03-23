package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVsphereEnableDiskUuid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VsphereEnableDiskUuid Suite")
}
