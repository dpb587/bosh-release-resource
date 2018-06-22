package testing

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/onsi/ginkgo"
	"github.com/pkg/errors"
)

func RunCommands(dir string, cmds []string) error {
	ginkgo.GinkgoWriter.Write([]byte(fmt.Sprintf("$ cd %s\n", dir)))

	for _, cmd := range cmds {
		ginkgo.GinkgoWriter.Write([]byte(fmt.Sprintf("$ %s\n", cmd)))

		cmd := exec.Command("bash", "-euc", cmd)
		cmd.Dir = dir
		cmd.Stdout = ginkgo.GinkgoWriter
		cmd.Stderr = ginkgo.GinkgoWriter

		err := cmd.Run()
		if err != nil {
			return errors.Wrapf(err, "running `%s`", cmd)
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
