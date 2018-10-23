package testing

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/onsi/ginkgo"
	"github.com/pkg/errors"
)

func RunCommands(dir string, commands []string) error {
	ginkgo.GinkgoWriter.Write([]byte(fmt.Sprintf("$ cd %s\n", dir)))

	for _, command := range commands {
		ginkgo.GinkgoWriter.Write([]byte(fmt.Sprintf("$ %s\n", command)))

		cmd := exec.Command("bash", "-euc", command)
		cmd.Dir = dir
		cmd.Stdout = ginkgo.GinkgoWriter
		cmd.Stderr = ginkgo.GinkgoWriter

		err := cmd.Run()
		if err != nil {
			return errors.Wrapf(err, "running `%s`", command)
		}
	}

	return nil
}

func RunCommandStdout(dir, executable string, args ...string) (string, error) {
	stdout := &bytes.Buffer{}

	cmd := exec.Command(executable, args...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = ginkgo.GinkgoWriter

	err := cmd.Run()

	return stdout.String(), err
}
