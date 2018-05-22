package main

import (
	"github.com/dpb587/bosh-release-resource/api"
)

var DefaultRequest = Request{
	Params: Params{
		Tarball: true,
	},
}

type Request struct {
	Source  api.Source  `json:"source"`
	Version api.Version `json:"version"`
	Params  Params      `json:"params"`
}

type Params struct {
	Tarball bool `json:"tarball"`
}

type Response struct {
	Version  api.Version    `json:"version"`
	Metadata []api.Metadata `json:"metadata,omitempty"`
}
