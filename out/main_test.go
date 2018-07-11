package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dpb587/bosh-release-resource/internal/testing"
	"github.com/onsi/gomega/gexec"
	yaml "gopkg.in/yaml.v2"
)

var _ = Describe("Main", func() {
	runCLI := func(stdin string) map[string]interface{} {
		command := exec.Command(cli, os.TempDir())
		command.Stdin = bytes.NewBufferString(stdin)

		stdout := &bytes.Buffer{}

		session, err := gexec.Start(command, stdout, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		session.Wait(time.Minute)

		var result map[string]interface{}

		err = json.Unmarshal(stdout.Bytes(), &result)
		Expect(err).NotTo(HaveOccurred())

		return result
	}

	var versionfile string

	BeforeEach(func() {
		version, err := ioutil.TempFile("", "bosh-release-resource-version-file")
		Expect(err).NotTo(HaveOccurred())

		versionfile = version.Name()
		_, err = version.WriteString("6.3.1")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if versionfile != "" {
			Expect(os.RemoveAll(versionfile)).NotTo(HaveOccurred())
		}
	})

	Describe("using a repository directory", func() {
		var forkdir string

		BeforeEach(func() {
			var err error

			forkdir, err = ioutil.TempDir("", "bosh-release-resource-fake-release")
			Expect(err).NotTo(HaveOccurred())

			err = testing.RunCommands(
				forkdir,
				[]string{
					fmt.Sprintf("git clone %s .", releasedir),
					"bosh generate-job fake12",
					"git add . && git commit -m fake12",
					"git push",
				},
			)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if forkdir != "" {
				Expect(os.RemoveAll(forkdir)).NotTo(HaveOccurred())
			}
		})

		It("finalizes the release", func() {
			result := runCLI(fmt.Sprintf(`{
		"source": {
			"uri": "%s",
			"branch": "master",
			"private_config": {
				"test": "private-config"
			}
		},
		"params": {
			"repository": "%s",
			"version": "%s"
		}
	}`, releasedir, forkdir, versionfile))
			Expect(result["version"].(map[string]interface{})["version"]).To(Equal("6.3.1"))
			Expect(result).To(HaveLen(2))
			Expect(result["metadata"].([]interface{})).To(ContainElement(HaveKeyWithValue("name", "bosh")))

			forkCommit, err := testing.RunCommandStdout(forkdir, "git", "rev-parse", "HEAD")
			Expect(err).NotTo(HaveOccurred())

			privateYmlBytes, err := ioutil.ReadFile(path.Join(forkdir, "config", "private.yml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(privateYmlBytes).To(ContainSubstring("test: private-config\n"))

			By("finalizing the release", func() {
				releaseBytes, err := ioutil.ReadFile(path.Join(releasedir, "releases/fake/fake-6.3.1.yml"))
				Expect(err).NotTo(HaveOccurred())

				var releaseManifest map[string]interface{}

				Expect(yaml.Unmarshal(releaseBytes, &releaseManifest)).NotTo(HaveOccurred())

				Expect(releaseManifest["commit_hash"]).To(Equal(forkCommit[0:7]))
				Expect(releaseManifest["jobs"].([]interface{})).To(ContainElement(HaveKeyWithValue("name", "fake12")))
			})

			By("pushing the final commit", func() {
				releasedirCommit, err := testing.RunCommandStdout(releasedir, "git", "rev-parse", "HEAD")
				Expect(err).NotTo(HaveOccurred())
				Expect(result["metadata"].([]interface{})).To(ContainElement(map[string]interface{}{
					"name":  "commit",
					"value": strings.TrimSpace(releasedirCommit),
				}))

				By("using correct commitership", func() {
					authorName, err := testing.RunCommandStdout(releasedir, "git", "log", "-n1", `--format=%cn`)
					Expect(err).NotTo(HaveOccurred())
					Expect(strings.TrimSpace(authorName)).To(Equal("CI Bot"))

					authorEmail, err := testing.RunCommandStdout(releasedir, "git", "log", "-n1", `--format=%ce`)
					Expect(err).NotTo(HaveOccurred())
					Expect(strings.TrimSpace(authorEmail)).To(Equal("ci@localhost"))
				})

				By("using correct authorship", func() {
					authorName, err := testing.RunCommandStdout(releasedir, "git", "log", "-n1", `--format=%an`)
					Expect(err).NotTo(HaveOccurred())
					Expect(strings.TrimSpace(authorName)).To(Equal("CI Bot"))

					authorEmail, err := testing.RunCommandStdout(releasedir, "git", "log", "-n1", `--format=%ae`)
					Expect(err).NotTo(HaveOccurred())
					Expect(strings.TrimSpace(authorEmail)).To(Equal("ci@localhost"))
				})
			})

			By("annotate-tagging the commit_hash", func() {
				taggedCommit, err := testing.RunCommandStdout(releasedir, "git", "rev-parse", "v6.3.1^{}")
				Expect(err).NotTo(HaveOccurred())
				Expect(taggedCommit).To(Equal(forkCommit))
			})
		})
	})
})
