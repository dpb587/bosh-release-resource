package testing

import (
	. "github.com/onsi/ginkgo"

  "fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func GenerateRelease() (string, error) {
	var cmds = []string{
		"git init .",
		"bosh init-release --git",
		"echo '--- { name: fake, blobstore: { provider: local, options: { blobstore_path: $releasedir/tmp/blobstore } } }' > $releasedir/config/final.yml",
		"bosh generate-job fake1",
		"bosh generate-package fake1",
		"git add . && git commit -m 'init'",
		"bosh create-release --final --version=1.0.0",
		"git add . && git commit -m 'v1.0.0'",
		"bosh generate-job fake2",
		"git add . && git commit -m 'fake2'",
		"bosh create-release --final --version=1.1.0",
		"git add . && git commit -m 'v1.1.0'",
		"bosh create-release --final --version=2.0.0",
		"git add . && git commit -m 'v2.0.0'",
		"bosh create-release --final --version=2.0.1 --name=custom-name",
		"git add . && git commit -m 'v2.0.1'",
		"git checkout -b custom-branch",
		"bosh create-release --final --version=3.0.1",
		"git add . && git commit -m 'v3.0.1'",
		"git checkout master",
	}

	releasedir, err := ioutil.TempDir("", "bosh-release-resource-fake-release")
	if err != nil {
		return "", err
	}

  GinkgoWriter.Write([]byte(fmt.Sprintf("$ cd %s", releasedir)))

	for _, cmd := range cmds {
    cmd = strings.Replace(cmd, "$releasedir", releasedir, -1)

    GinkgoWriter.Write([]byte(fmt.Sprintf("$ %s\n", cmd)))

		cmd := exec.Command("bash", "-euc", cmd)
		cmd.Dir = releasedir
    cmd.Stdout = GinkgoWriter
    cmd.Stderr = GinkgoWriter

		err := cmd.Run()
		if err != nil {
			os.RemoveAll(releasedir)

			return "", errors.Wrapf(err, "running `%s`", cmd)
		}
	}

	return releasedir, nil
}
