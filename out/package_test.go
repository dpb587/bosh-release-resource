package main_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	pkgtesting "github.com/dpb587/bosh-release-resource/internal/testing"
	"github.com/onsi/gomega/gexec"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "github.com/dpb587/bosh-release-resource/out")
}

var cli string
var releasedir string

var _ = BeforeSuite(func() {
	var err error

	cli, err = gexec.Build("github.com/dpb587/bosh-release-resource/out")
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	var err error

	releasedir, err = pkgtesting.GenerateRelease()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterEach(func() {
	if releasedir != "" {
		err := os.RemoveAll(releasedir)
		Expect(err).NotTo(HaveOccurred())
	}
})
