package boshrelease

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

type Release struct {
	repository    *Repository
	privateConfig map[string]interface{}
}

func NewRelease(repository *Repository, privateConfig map[string]interface{}) *Release {
	return &Release{
		repository:    repository,
		privateConfig: privateConfig,
	}
}

func (r Release) Name() (string, error) {
	bytes, err := ioutil.ReadFile(path.Join(r.repository.Path(), "config", "final.yml"))
	if err != nil {
		return "", errors.Wrap(err, "reading final.yml")
	}

	var config releaseConfig

	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return "", errors.Wrap(err, "parsing final.yml")
	}

	return config.Name(), nil
}

func (r Release) DevVersions(name, latestVersionCommit string) ([]*semver.Version, error) {
	commits, err := r.repository.GetCommitList(latestVersionCommit)
	if err != nil {
		return nil, errors.Wrap(err, "loading commits")
	}

	var versions []*semver.Version

	for _, commit := range commits {
		indexBytes, err := r.repository.Show(commit.Commit, path.Join("releases", name, "index.yml"))
		if err != nil {
			if err.Error() == "exit status 128" {
				// inexact error checking, but check if it's a top-level release
				// this is hacky to support old, stubborn releases like cloudfoundry/bosh which currently use a symlink
				indexBytes, err = r.repository.Show(commit.Commit, path.Join("releases/index.yml"))
			}

			if err != nil {
				return nil, errors.Wrapf(err, "loading releases index.yml for %s", commit.Commit)
			}
		}

		parsedVersions, err := r.parseReleaseIndex(indexBytes)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing releases index.yml for %s", commit.Commit)
		}

		baseVersion, err := parsedVersions[0].IncPatch().SetPrerelease(fmt.Sprintf("dev.%s.commit.%s", commit.CommitDate.Format("20060102T150405Z"), commit.Commit))
		if err != nil {
			return nil, errors.Wrapf(err, "creating version for %s", commit.Commit)
		}

		versions = append(versions, &baseVersion)
	}

	sort.Sort(sort.Reverse(semver.Collection(versions)))

	return versions, nil
}

func (r Release) Versions(name string, constraints []*semver.Constraints, latestVersion string) ([]*semver.Version, error) {
	bytes, err := ioutil.ReadFile(path.Join(r.repository.Path(), "releases", name, "index.yml"))
	if err != nil {
		return nil, errors.Wrap(err, "reading index.yml")
	}

	parsedVersions, err := r.parseReleaseIndex(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "parsing index.yml")
	}

	var versions []*semver.Version

	for _, version := range parsedVersions {
		if version.Original() == latestVersion {
			// always include
		} else {
			// rely on constraints
			match := true

			for _, constraint := range constraints {
				match = match && constraint.Check(version)
			}

			if !match {
				continue
			}
		}

		versions = append(versions, version)
	}

	return versions, nil
}

func (r Release) CreateDevTarball(name, version, tarball string) error {
	parsedVersion, err := semver.NewVersion(version)
	if err != nil {
		return errors.Wrap(err, "parsing dev version")
	}

	prereleaseSplit := strings.Split(parsedVersion.Prerelease(), ".")
	if prereleaseSplit[2] != "commit" {
		return errors.New("commit expected in prerelease")
	}

	err = r.repository.Checkout(prereleaseSplit[3])
	if err != nil {
		return errors.Wrap(err, "checking out dev release")
	}

	err = r.writePrivateConfig()
	if err != nil {
		return errors.Wrap(err, "private.yml")
	}

	cmd := exec.Command(
		"bosh",
		"create-release",
		"--force",
		"--tarball", tarball,
		"--version", version,
	)
	cmd.Dir = r.repository.Path()
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "creating tarball")
	}

	return nil
}

func (r Release) CreateTarball(name, version, tarball string) error {
	err := r.writePrivateConfig()
	if err != nil {
		return errors.Wrap(err, "private.yml")
	}

	cmd := exec.Command(
		"bosh",
		"create-release",
		"--tarball", tarball,
		filepath.Join("releases", name, fmt.Sprintf("%s-%s.yml", name, version)),
	)
	cmd.Dir = r.repository.Path()
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "creating tarball")
	}

	return nil
}

func (r Release) FinalizeRelease(name, version, tarball string) (string, error) {
	err := r.writePrivateConfig()
	if err != nil {
		return "", errors.Wrap(err, "private.yml")
	}

	cmd := exec.Command(
		"bosh",
		"finalize-release",
		"--name", name,
		"--version", version,
		tarball,
	)
	cmd.Dir = r.repository.Path()
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return "", errors.Wrap(err, "finalizing release")
	}

	releaseManifestBytes, err := ioutil.ReadFile(path.Join(r.repository.Path(), "releases", name, fmt.Sprintf("%s-%s.yml", name, version)))
	if err != nil {
		return "", errors.Wrap(err, "reading finalized release")
	}

	var parsed releaseVersion

	err = yaml.Unmarshal(releaseManifestBytes, &parsed)
	if err != nil {
		return "", errors.Wrap(err, "parsing finalized release")
	}

	return parsed.CommitHash, nil
}

func (r Release) writePrivateConfig() error {
	if r.privateConfig == nil {
		return nil
	}

	bytes, err := yaml.Marshal(r.privateConfig)
	if err != nil {
		return errors.Wrap(err, "marshalling private.yml")
	}

	err = ioutil.WriteFile(path.Join(r.repository.Path(), "config", "private.yml"), bytes, 0700)
	if err != nil {
		return errors.Wrap(err, "writing private.yml")
	}

	return nil
}

func (r Release) parseReleaseIndex(bytes []byte) ([]*semver.Version, error) {
	var index releaseIndex

	err := yaml.Unmarshal(bytes, &index)
	if err != nil {
		return nil, errors.Wrap(err, "parsing index.yml")
	}

	var versions []*semver.Version

	for _, build := range index.Builds {
		version, err := semver.NewVersion(build.Version)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing version %s", build.Version)
		}

		versions = append(versions, version)
	}

	sort.Sort(sort.Reverse(semver.Collection(versions)))

	return versions, nil
}
