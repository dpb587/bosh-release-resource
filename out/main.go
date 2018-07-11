package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/dpb587/bosh-release-resource/api"
	"github.com/dpb587/bosh-release-resource/boshrelease"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	err := os.Chdir(os.Args[1])
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad args: source dir"))
	}

	request := DefaultRequest

	err = json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad stdin: parse error"))
	}

	if request.Source.Branch == "" {
		api.Fatal(errors.New("bad source: branch is required"))
	}

	version := loadVersion(request)
	commitMessage := loadCommitMessage(request, version)

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

	tarballPath := loadTarballPath(request, release)

	err = repository.Configure(request.Params.AuthorName, request.Params.AuthorEmail)
	if err != nil {
		api.Fatal(errors.Wrap(err, "configuring"))
	}

	versionCommitHash, err := release.FinalizeRelease(releaseName, version, tarballPath)
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad release tarball"))
	}

	commit, err := repository.Commit(commitMessage, request.Params.Rebase)
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad commit"))
	}

	if !request.Params.SkipTag {
		tag := fmt.Sprintf("v%s", version)
		err = repository.Tag(versionCommitHash, tag, tag)
		if err != nil {
			api.Fatal(errors.Wrap(err, "bad tag"))
		}
	}

	err = json.NewEncoder(os.Stdout).Encode(Response{
		Version: api.Version{
			Version: version,
		},
		Metadata: []api.Metadata{
			{
				Name:  "bosh",
				Value: boshrelease.BoshVersion(),
			},
			{
				Name:  "commit",
				Value: commit,
			},
		},
	})
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad stdout: json"))
	}
}

func loadVersion(request Request) string {
	versionPaths, err := filepath.Glob(request.Params.Version)
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad params: globbing version"))
	} else if len(versionPaths) == 0 {
		api.Fatal(errors.New("bad params: version path not found"))
	} else if len(versionPaths) > 1 {
		api.Fatal(errors.New("bad params: multiple files matched version path"))
	}

	versionBytes, err := ioutil.ReadFile(versionPaths[0])
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad version: reading"))
	}

	return strings.TrimSpace(string(versionBytes))
}

func loadCommitMessage(request Request, version string) string {
	var commitMessage = fmt.Sprintf("Version %s", version)

	if request.Params.CommitFile != "" {
		commitFilePaths, err := filepath.Glob(request.Params.CommitFile)
		if err != nil {
			api.Fatal(errors.Wrap(err, "bad params: globbing commit"))
		} else if len(commitFilePaths) == 0 {
			api.Fatal(errors.New("bad params: commit path not found"))
		} else if len(commitFilePaths) > 1 {
			api.Fatal(errors.New("bad params: multiple files matched commit path"))
		}

		commitBytes, err := ioutil.ReadFile(commitFilePaths[0])
		if err != nil {
			api.Fatal(errors.Wrap(err, "bad commit file: reading"))
		}

		commitMessage = strings.TrimSpace(string(commitBytes))
	}

	return commitMessage
}

func loadTarballPath(request Request, release *boshrelease.Release) string {
	if request.Params.Tarball != "" && request.Params.Repository != "" {
		api.Fatal(errors.New("bad params: only tarball or repository may be configured"))
	} else if request.Params.Repository != "" {
		if request.Source.PrivateConfig != nil {
			privateYmlBytes, err := yaml.Marshal(request.Source.PrivateConfig)
			if err != nil {
				api.Fatal(errors.Wrap(err, "marshalling private.yml"))
			}

			err = ioutil.WriteFile(path.Join(request.Params.Repository, "config", "private.yml"), privateYmlBytes, 0700)
			if err != nil {
				api.Fatal(errors.Wrap(err, "writing private.yml"))
			}
		}

		tarballPath := path.Join(request.Params.Repository, "release.tgz")

		cmd := exec.Command("bosh", "create-release", "--force", "--tarball", tarballPath)
		cmd.Dir = request.Params.Repository
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			api.Fatal(errors.Wrap(err, "bad repository: creating release"))
		}

		return tarballPath
	}

	tarballPaths, err := filepath.Glob(request.Params.Tarball)
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad params: globbing tarball"))
	} else if len(tarballPaths) == 0 {
		api.Fatal(errors.New("bad params: tarball path not found"))
	} else if len(tarballPaths) > 1 {
		api.Fatal(errors.New("bad params: multiple files matched tarball path"))
	}

	tarballPath, err := filepath.Abs(tarballPaths[0])
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad params: absolute tarball path"))
	}

	return tarballPath
}
