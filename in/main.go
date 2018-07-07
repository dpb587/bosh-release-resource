package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/dpb587/bosh-release-resource/api"
	"github.com/dpb587/bosh-release-resource/boshrelease"
	"github.com/pkg/errors"
)

func main() {
	if len(os.Args) < 2 {
		api.Fatal(errors.Wrap(fmt.Errorf("%s DESTINATION-DIR", os.Args[0]), "in: bad invocation"))
	}

	destination := os.Args[1]

	request := DefaultRequest

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad stdin: parse error"))
	}

	tarballNameTmpl, err := template.New("tarball_name").Parse(request.Params.TarballName)
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad config: file_name"))
	}

	repository := boshrelease.NewRepository(request.Source.URI, request.Source.Branch, request.Source.PrivateKey)

	err = repository.Pull()
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad repository: pulling"))
	}

	release := boshrelease.NewRelease(repository, request.Source.PrivateConfig)

	releaseName := request.Source.Name

	if releaseName == "" {
		releaseName, err = release.Name()
		if err != nil {
			api.Fatal(errors.Wrap(err, "bad release: discovering name"))
		}
	}

	tarballNameBuffer := &bytes.Buffer{}
	err = tarballNameTmpl.Execute(tarballNameBuffer, struct {
		Name    string
		Version string
	}{
		Name:    releaseName,
		Version: request.Version.Version,
	})
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad config: generating tarball name"))
	}

	if request.Params.Tarball {
		var f func(string, string, string) error

		if request.Source.DevReleases {
			f = release.CreateDevTarball
		} else {
			f = release.CreateTarball
		}

		err = f(
			releaseName,
			request.Version.Version,
			filepath.Join(destination, tarballNameBuffer.String()),
		)
		if err != nil {
			api.Fatal(errors.Wrap(err, "bad release"))
		}
	}

	err = ioutil.WriteFile(filepath.Join(destination, "name"), []byte(releaseName), 0644)
	if err != nil {
		api.Fatal(errors.Wrap(err, "fs metadata: name"))
	}

	err = ioutil.WriteFile(filepath.Join(destination, "version"), []byte(request.Version.Version), 0644)
	if err != nil {
		api.Fatal(errors.Wrap(err, "fs metadata: version"))
	}

	err = json.NewEncoder(os.Stdout).Encode(Response{
		Version: request.Version,
		Metadata: []api.Metadata{
			{
				Name:  "bosh",
				Value: boshrelease.BoshVersion(),
			},
			{
				Name:  "time",
				Value: time.Now().Format(time.RFC3339),
			},
		},
	})
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad stdout: json"))
	}
}
