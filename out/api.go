package main

import (
	"github.com/dpb587/bosh-release-resource/api"
)

var DefaultRequest = Request{
	Params: Params{
		CommitterName:  "CI Bot",
		CommitterEmail: "ci@localhost",
	},
}

type Request struct {
	Source api.Source `json:"source"`
	Params Params     `json:"params"`
}

type Params struct {
	Tarball    string `json:"tarball,omitempty"`
	Repository string `json:"repository,omitempty"`
	Version    string `json:"version"`

	CommitFile     string `json:"commit_file,omitempty"`
	CommitterName  string `json:"committer_name,omitempty"`
	CommitterEmail string `json:"committer_email,omitempty"`
	Rebase         bool   `json:"rebase,omitempty"`
	SkipTag        bool   `json:"skip_tag,omitempty"`
}

type Response struct {
	Version  api.Version    `json:"version"`
	Metadata []api.Metadata `json:"metadata,omitempty"`
}
