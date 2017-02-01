package main_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/route-registrar/config"
	. "github.com/onsi/ginkgo"
	gconfig "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

const (
	routeRegistrarPackage = "code.cloudfoundry.org/route-registrar/"
)

var (
	routeRegistrarBinPath string
	pidFile               string
	configFile            string
	rootConfig            config.ConfigSchema
	natsPort              int

	tempDir string
)

func TestRouteRegistrar(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}

var _ = BeforeSuite(func() {
	var err error
	routeRegistrarBinPath, err = gexec.Build(routeRegistrarPackage, "-race")
	Expect(err).ShouldNot(HaveOccurred())

	tempDir, err = ioutil.TempDir(os.TempDir(), "route-registrar")
	Expect(err).ShouldNot(HaveOccurred())

	pidFile = filepath.Join(tempDir, "route-registrar.pid")

	natsPort = 40000 + gconfig.GinkgoConfig.ParallelNode

	configFile = filepath.Join(tempDir, "registrar_settings.yml")
})

var _ = AfterSuite(func() {
	err := os.RemoveAll(tempDir)
	Expect(err).ShouldNot(HaveOccurred())

	gexec.CleanupBuildArtifacts()
})
