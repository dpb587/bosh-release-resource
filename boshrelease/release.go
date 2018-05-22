package boshrelease

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"

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

func (r Release) Versions(name string, constraints []*semver.Constraints) ([]*semver.Version, error) {
	bytes, err := ioutil.ReadFile(path.Join(r.repository.Path(), "releases", name, "index.yml"))
	if err != nil {
		return nil, errors.Wrap(err, "reading index.yml")
	}

	var index releaseIndex

	err = yaml.Unmarshal(bytes, &index)
	if err != nil {
		return nil, errors.Wrap(err, "parsing index.yml")
	}

	var versions []*semver.Version

	for _, build := range index.Builds {
		version, err := semver.NewVersion(build.Version)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing version %s", build.Version)
		}

		match := true

		for _, constraint := range constraints {
			match = match && constraint.Check(version)
		}

		if !match {
			continue
		}

		versions = append(versions, version)
	}

	sort.Sort(sort.Reverse(semver.Collection(versions)))

	return versions, nil
}

func (r Release) CreateTarball(name, version, tarball string) error {
	err := r.writePrivateConfig()
	if err != nil {
		return errors.Wrap(err, "private.yml")
	}

	cmd := exec.Command(
		"bosh",
		"create-release",
		"--tarball",
		tarball,
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
