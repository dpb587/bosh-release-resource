package api

import (
	"fmt"
	"os"
	"path"
)

func Fatal(err error) {
	executable, _ := os.Executable()
	executable = path.Base(executable)
	if executable == "" {
		executable = "main"
	}

	fmt.Fprintln(os.Stderr, fmt.Sprintf("%s: %s", executable, err))

	os.Exit(1)
}
