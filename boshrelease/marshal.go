package boshrelease

type releaseConfig struct {
	Name_      string `yaml:"name"`
	FinalName_ string `yaml:"final_name"`
}

func (rc releaseConfig) Name() string {
	if rc.Name_ != "" {
		return rc.Name_
	}

	return rc.FinalName_
}

type releaseIndex struct {
	Builds map[string]releaseIndexBuild `yaml:"builds"`
}

type releaseIndexBuild struct {
	Version string `yaml:"version"`
}

type releaseVersion struct {
	CommitHash string `yaml:"commit_hash"`
}
