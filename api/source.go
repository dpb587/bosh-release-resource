package api

import (
	"encoding/json"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
)

type Source struct {
	Repository         string                 `json:"repository"`
	Branch             string                 `json:"branch"`
	Name               string                 `json:"name,omitempty"`
	Version            string                 `json:"version,omitempty"`
	DevReleases        bool                   `json:"dev_releases,omitempty"`
	VersionConstraints *semver.Constraints    `json:"-"`
	PrivateConfig      map[string]interface{} `json:"private_config,omitempty"`
	PrivateKey         string                 `json:"private_key"`
}

func (s *Source) UnmarshalJSON(data []byte) error {
	type unmarshal Source
	if err := json.Unmarshal(data, (*unmarshal)(s)); err != nil {
		return err
	}

	if s.Version != "" {
		constraints, err := semver.NewConstraint(s.Version)
		if err != nil {
			return errors.Wrap(err, "parsing version")
		}

		s.VersionConstraints = constraints
	}

	return nil
}
