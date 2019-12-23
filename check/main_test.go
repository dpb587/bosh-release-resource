package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dpb587/bosh-release-resource/internal/testing"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	runCheck := func(stdin string) []map[string]interface{} {
		command := exec.Command(cli)
		command.Stdin = bytes.NewBufferString(stdin)

		stdout := &bytes.Buffer{}

		session, err := gexec.Start(command, stdout, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		session.Wait(time.Minute)

		var versions []map[string]interface{}

		err = json.Unmarshal(stdout.Bytes(), &versions)
		Expect(err).NotTo(HaveOccurred())

		return versions
	}

	Context("fake release", func() {
		var releasedir string

		BeforeEach(func() {
			var err error

			releasedir, err = testing.GenerateRelease()
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if releasedir != "" {
				err := os.RemoveAll(releasedir)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("gets the latest version", func() {
			versions := runCheck(fmt.Sprintf(`{
		"source": {
			"uri": "%s"
		}
	}`, releasedir))

			Expect(versions).To(HaveLen(1))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "2.0.0")))
		})

		It("repeats the latest version", func() {
			versions := runCheck(fmt.Sprintf(`{
		"source": {
			"uri": "%s"
		},
		"version": {
			"version": "2.0.0"
		}
	}`, releasedir))

			Expect(versions).To(HaveLen(1))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "2.0.0")))
		})

		It("respects version constraints", func() {
			versions := runCheck(fmt.Sprintf(`{
		"source": {
			"uri": "%s",
			"version": "1.x"
		}
	}`, releasedir))

			Expect(versions).To(HaveLen(1))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "1.1.0")))
		})

		It("fetches multiple new versions", func() {
			versions := runCheck(fmt.Sprintf(`{
		"source": {
			"uri": "%s"
		},
		"version": {
			"version": "1.0.0"
		}
	}`, releasedir))

			Expect(versions).To(HaveLen(3))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "1.0.0")))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "1.1.0")))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "2.0.0")))
		})

		It("supports referencing non-default release name", func() {
			versions := runCheck(fmt.Sprintf(`{
		"source": {
			"uri": "%s",
			"name": "custom-name"
		}
	}`, releasedir))

			Expect(versions).To(HaveLen(1))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "2.0.1")))
		})

		It("supports referencing non-default branch", func() {
			versions := runCheck(fmt.Sprintf(`{
		"source": {
			"uri": "%s",
			"branch": "custom-branch"
		}
	}`, releasedir))

			Expect(versions).To(HaveLen(1))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "3.0.1")))
		})

		Describe("dev_releases = true", func() {
			It("fetches dev releases", func() {
				versions := runCheck(fmt.Sprintf(`{
			"source": {
				"uri": "%s",
				"dev_releases": true
			}
		}`, releasedir))

				lastCommit, err := testing.RunCommandStdout(releasedir, "git", "rev-parse", "--short", "HEAD")
				Expect(err).NotTo(HaveOccurred())
				lastCommit = strings.TrimSpace(lastCommit)

				lastCommitDate, err := testing.RunCommandStdout(releasedir, "git", "log", "-n1", "--format=%ci", lastCommit)
				Expect(err).NotTo(HaveOccurred())

				lastCommitTime, err := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(lastCommitDate))
				Expect(err).NotTo(HaveOccurred())

				Expect(versions).To(HaveLen(1))
				Expect(versions).To(ContainElement(HaveKeyWithValue("version", fmt.Sprintf("2.0.1-dev.%s.commit.%s", lastCommitTime.UTC().Format("20060102T150405Z"), lastCommit))))
			})

			It("only follows the 0th parent in merges", func() {
				preMergeCommit, err := testing.RunCommandStdout(releasedir, "git", "rev-parse", "--short", "HEAD")
				Expect(err).NotTo(HaveOccurred())
				preMergeCommit = strings.TrimSpace(preMergeCommit)

				preMergeCommitDate, err := testing.RunCommandStdout(releasedir, "git", "log", "-n1", "--format=%ci", preMergeCommit)
				Expect(err).NotTo(HaveOccurred())

				preMergeCommitTime, err := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(preMergeCommitDate))
				Expect(err).NotTo(HaveOccurred())

				err = testing.RunCommands(
					releasedir,
					[]string{
						"git checkout -b mergeable",
						"touch mergeme",
						"git add mergeme",
						"git commit -m mergeit",
						"git checkout master",
						"git merge --no-ff mergeable",
					},
				)
				Expect(err).NotTo(HaveOccurred())

				lastCommit, err := testing.RunCommandStdout(releasedir, "git", "rev-parse", "--short", "HEAD")
				Expect(err).NotTo(HaveOccurred())
				lastCommit = strings.TrimSpace(lastCommit)

				lastCommitDate, err := testing.RunCommandStdout(releasedir, "git", "log", "-n1", "--format=%ci", lastCommit)
				Expect(err).NotTo(HaveOccurred())

				lastCommitTime, err := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(lastCommitDate))
				Expect(err).NotTo(HaveOccurred())

				versions := runCheck(fmt.Sprintf(`{
			"source": {
				"uri": "%s",
				"dev_releases": true
			},
			"version": {
				"version": "0.0.0-dev.YYYYMMDDTHHIISSZ.commit.%s"
			}
		}`, releasedir, preMergeCommit))

				Expect(versions).To(HaveLen(2))
				Expect(versions).To(ContainElement(HaveKeyWithValue("version", fmt.Sprintf("2.0.1-dev.%s.commit.%s", preMergeCommitTime.UTC().Format("20060102T150405Z"), preMergeCommit))))
				Expect(versions).To(ContainElement(HaveKeyWithValue("version", fmt.Sprintf("2.0.1-dev.%s.commit.%s", lastCommitTime.UTC().Format("20060102T150405Z"), lastCommit))))
			})

			It("repeats the latest version if there are no changes", func() {
				lastCommit, err := testing.RunCommandStdout(releasedir, "git", "rev-parse", "--short", "HEAD")
				Expect(err).NotTo(HaveOccurred())
				lastCommit = strings.TrimSpace(lastCommit)

				lastCommitDate, err := testing.RunCommandStdout(releasedir, "git", "log", "-n1", "--format=%ci", lastCommit)
				Expect(err).NotTo(HaveOccurred())

				lastCommitTime, err := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(lastCommitDate))
				Expect(err).NotTo(HaveOccurred())

				sinceVersion := fmt.Sprintf("2.0.1-dev.%s.commit.%s", lastCommitTime.UTC().Format("20060102T150405Z"), lastCommit)

				versions := runCheck(fmt.Sprintf(`{
			"source": {
				"uri": "%s",
				"dev_releases": true
			},
			"version": {
				"version": "%s"
			}
		}`, releasedir, sinceVersion))

				Expect(versions).To(HaveLen(1))
				Expect(versions).To(ContainElement(HaveKeyWithValue("version", sinceVersion)))
			})

			It("fetches multiple dev releases", func() {
				thirdCommit, err := testing.RunCommandStdout(releasedir, "git", "rev-parse", "--short", "HEAD~2")
				Expect(err).NotTo(HaveOccurred())
				thirdCommit = strings.TrimSpace(thirdCommit)

				thirdCommitDate, err := testing.RunCommandStdout(releasedir, "git", "log", "-n1", "--format=%ci", thirdCommit)
				Expect(err).NotTo(HaveOccurred())

				thirdCommitTime, err := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(thirdCommitDate))
				Expect(err).NotTo(HaveOccurred())

				versions := runCheck(fmt.Sprintf(`{
			"source": {
				"uri": "%s",
				"dev_releases": true
			},
			"version": {
				"version": "0.0.0-dev.currentlyignored.commit.%s"
			}
		}`, releasedir, strings.TrimSpace(thirdCommit)))

				lastCommit, err := testing.RunCommandStdout(releasedir, "git", "rev-parse", "--short", "HEAD")
				Expect(err).NotTo(HaveOccurred())
				lastCommit = strings.TrimSpace(lastCommit)

				lastCommitDate, err := testing.RunCommandStdout(releasedir, "git", "log", "-n1", "--format=%ci", lastCommit)
				Expect(err).NotTo(HaveOccurred())

				lastCommitTime, err := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(lastCommitDate))
				Expect(err).NotTo(HaveOccurred())

				secondCommit, err := testing.RunCommandStdout(releasedir, "git", "rev-parse", "--short", "HEAD~1")
				Expect(err).NotTo(HaveOccurred())
				secondCommit = strings.TrimSpace(secondCommit)

				secondCommitDate, err := testing.RunCommandStdout(releasedir, "git", "log", "-n1", "--format=%ci", secondCommit)
				Expect(err).NotTo(HaveOccurred())

				secondCommitTime, err := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(secondCommitDate))
				Expect(err).NotTo(HaveOccurred())

				Expect(versions).To(HaveLen(3))
				Expect(versions).To(ContainElement(HaveKeyWithValue("version", fmt.Sprintf("2.0.1-dev.%s.commit.%s", lastCommitTime.UTC().Format("20060102T150405Z"), lastCommit))))
				Expect(versions).To(ContainElement(HaveKeyWithValue("version", fmt.Sprintf("2.0.1-dev.%s.commit.%s", secondCommitTime.UTC().Format("20060102T150405Z"), secondCommit))))
				Expect(versions).To(ContainElement(HaveKeyWithValue("version", fmt.Sprintf("1.1.1-dev.%s.commit.%s", thirdCommitTime.UTC().Format("20060102T150405Z"), thirdCommit))))
			})
		})
	})
})
