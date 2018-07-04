package main

import (
	"github.com/dpb587/bosh-release-resource/api"
)

var DefaultRequest = Request{
	Params: Params{
		TarballName: "{{.Name}}-{{.Version}}.tgz",
		Tarball:     true,
	},
}

type Request struct {
	Source  api.Source  `json:"source"`
	Version api.Version `json:"version"`
	Params  Params      `json:"params"`
}

type Params struct {
	TarballName string `json:"tarball_name"`
	Tarball     bool   `json:"tarball"`
}

type Response struct {
	Version  api.Version    `json:"version"`
	Metadata []api.Metadata `json:"metadata,omitempty"`
}
