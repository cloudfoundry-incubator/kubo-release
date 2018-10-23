package smoke_tests_test

import (
	"html/template"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"smoke-tests/runner"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestK8SCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CFCR Smoke-Tests Suite")
}

var (
	k8sRunner *runner.KubectlRunner
	tmpDir    string
	pspSpec   string
)

var _ = BeforeSuite(func() {
	var err error
	rand.Seed(time.Now().UnixNano())
	tmpDir, err = ioutil.TempDir("", "smoke-tests")
	Expect(err).NotTo(HaveOccurred())

	k8sRunner = runner.NewKubectlRunner()
	k8sRunner.RunKubectlCommand("create", "namespace", k8sRunner.Namespace()).Wait("60s")
	pspSpec = templatePSPWithNamespace(tmpDir, k8sRunner.Namespace())
	k8sRunner.RunKubectlCommand("apply", "-f", pspSpec)
})

func getFixturePath(filename string) string {
	srcDir, err := filepath.Abs(filepath.Dir(filename))
	Expect(err).NotTo(HaveOccurred())

	return filepath.Join(srcDir, "fixtures", "smoke-test-psp.yml")

}

func templatePSPWithNamespace(tmpDir, namespace string) string {
	_, filename, _, _ := runtime.Caller(0)
	file := getFixturePath(filename)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		filename, err = os.Executable()
		Expect(err).NotTo(HaveOccurred())
		file = getFixturePath(filename)
	}

	pspName := "smoke-test-" + namespace
	t, err := template.ParseFiles(file)
	Expect(err).NotTo(HaveOccurred())

	f, err := ioutil.TempFile(tmpDir, filepath.Base(file))
	Expect(err).NotTo(HaveOccurred())
	defer f.Close()

	type templateInfo struct{ PSPName, Namespace string }
	Expect(t.Execute(f, templateInfo{PSPName: pspName, Namespace: namespace})).To(Succeed())

	return f.Name()
}

var _ = AfterSuite(func() {
	if k8sRunner != nil {
		k8sRunner.RunKubectlCommand("delete", "-f", pspSpec).Wait("60s")
		k8sRunner.RunKubectlCommand("delete", "namespace", k8sRunner.Namespace()).Wait("60s")
	}
})
