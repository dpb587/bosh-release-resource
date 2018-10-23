package testing

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

func GenerateRelease() (string, error) {
	releasedir, err := ioutil.TempDir("", "bosh-release-resource-fake-release")
	if err != nil {
		return "", err
	}

	err = RunCommands(
		releasedir,
		[]string{
			"git init .",
			"git config receive.denyCurrentBranch updateInstead",
			"bosh init-release --git",
			"echo '--- { name: fake, blobstore: { provider: local, options: { blobstore_path: $PWD/tmp/blobstore } } }' > $PWD/config/final.yml",
			"bosh generate-job fake1",
			"bosh generate-package fake1",
			"touch src/.gitkeep",
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
		},
	)

	if err != nil {
		os.RemoveAll(releasedir)

		return "", errors.Wrap(err, "generating release")
	}

	return releasedir, nil
}
