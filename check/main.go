package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/dpb587/bosh-release-resource/api"
	"github.com/dpb587/bosh-release-resource/boshrelease"
	"github.com/pkg/errors"
)

func main() {
	var request Request

	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad stdin: parse error"))
	}

	var constraints []*semver.Constraints

	if request.Source.VersionConstraints != nil {
		constraints = append(constraints, request.Source.VersionConstraints)
	}

	repository := boshrelease.NewRepository(request.Source.Repository, request.Source.Branch, request.Source.PrivateKey)

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

	var versionsRaw []*semver.Version

	if request.Source.DevReleases {
		var sinceCommit string

		if request.Version != nil {
			version, err := semver.NewVersion(request.Version.Version)
			if err != nil {
				api.Fatal(errors.Wrap(err, "bad version: parsing"))
			}

			prereleaseSplit := strings.Split(version.Prerelease(), ".")
			if prereleaseSplit[2] != "commit" {
				api.Fatal(errors.New("bad version: commit expected in prerelease"))
			}

			sinceCommit = prereleaseSplit[3]
		}

		versionsRaw, err = release.DevVersions(releaseName, sinceCommit)
		if err != nil {
			api.Fatal(errors.Wrap(err, "bad release: versions"))
		}
	} else {
		if request.Version != nil {
			constraint, err := semver.NewConstraint(fmt.Sprintf(">%s", request.Version.Version))
			if err != nil {
				api.Fatal(errors.Wrap(err, "bad version: version"))
			}

			constraints = append(constraints, constraint)
		}

		versionsRaw, err = release.Versions(releaseName, constraints)
		if err != nil {
			api.Fatal(errors.Wrap(err, "bad release: versions"))
		}
	}

	response := Response{}

	for _, version := range versionsRaw {
		response = append(response, api.Version{
			Version: version.Original(),
		})
	}

	if request.Version == nil {
		// if no prior version, only enumerate the most recent
		response = response[0:1]
	}

	err = json.NewEncoder(os.Stdout).Encode(response)
	if err != nil {
		api.Fatal(errors.Wrap(err, "bad stdout: json"))
	}
}
