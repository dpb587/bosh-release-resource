package main

import (
	"github.com/dpb587/bosh-release-resource/api"
)

type Request struct {
	Source  api.Source   `json:"source"`
	Version *api.Version `json:"version"`
}

type Response []api.Version
