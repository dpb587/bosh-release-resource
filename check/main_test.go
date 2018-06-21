package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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

		It("gets latest the version", func() {
			versions := runCheck(fmt.Sprintf(`{
		"source": {
			"repository": "%s"
		}
	}`, releasedir))

			Expect(versions).To(HaveLen(1))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "2.0.0")))
		})

		It("respects version constraints", func() {
			versions := runCheck(fmt.Sprintf(`{
		"source": {
			"repository": "%s",
			"version": "1.x"
		}
	}`, releasedir))

			Expect(versions).To(HaveLen(1))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "1.1.0")))
		})

		It("fetches multiple new versions", func() {
			versions := runCheck(fmt.Sprintf(`{
		"source": {
			"repository": "%s"
		},
		"version": {
			"version": "1.0.0"
		}
	}`, releasedir))

			Expect(versions).To(HaveLen(2))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "1.1.0")))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "2.0.0")))
		})

		It("supports referencing non-default release name", func() {
			versions := runCheck(fmt.Sprintf(`{
		"source": {
			"repository": "%s",
			"name": "custom-name"
		}
	}`, releasedir))

			Expect(versions).To(HaveLen(1))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "2.0.1")))
		})

		It("supports referencing non-default branch", func() {
			versions := runCheck(fmt.Sprintf(`{
		"source": {
			"repository": "%s",
			"branch": "custom-branch"
		}
	}`, releasedir))

			Expect(versions).To(HaveLen(1))
			Expect(versions).To(ContainElement(HaveKeyWithValue("version", "3.0.1")))
		})
	})
})
