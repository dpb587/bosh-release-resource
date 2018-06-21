package main

import (
	"github.com/dpb587/bosh-release-resource/api"
)

var DefaultRequest = Request{
	Params: Params{
		AuthorName:  "CI Bot",
		AuthorEmail: "ci@localhost",
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

	CommitFile  string `json:"commit_file,omitempty"`
	AuthorName  string `json:"author_name,omitempty"`
	AuthorEmail string `json:"author_email,omitempty"`
	Rebase      bool   `json:"rebase,omitempty"`
	SkipTag     bool   `json:"skip_tag,omitempty"`
}

type Response struct {
	Version  api.Version    `json:"version"`
	Metadata []api.Metadata `json:"metadata,omitempty"`
}
