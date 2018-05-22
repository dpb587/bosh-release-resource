package boshrelease

import (
	"bytes"
	"os/exec"
	"strings"
)

func BoshVersion() string {
	stdout := bytes.NewBuffer(nil)

	cmd := exec.Command("bosh", "--version")
	cmd.Stdout = stdout

	err := cmd.Run()
	if err != nil {
		return "unknown"
	}

	pieces := strings.SplitN(strings.TrimPrefix(stdout.String(), "version "), "\n", 2)

	return pieces[0]
}
