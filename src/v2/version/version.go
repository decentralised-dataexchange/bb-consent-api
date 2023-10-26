package version

var version = "0.0.0"
var candidate = "dev"
var gitCommit = ""

type versionResp struct {
	Version   string `json:"version"`
	Candidate string `json:"candidate"`
	Revision  string `json:"revision"`
}

func VersionFill() versionResp {
	versionJSON := versionResp{
		Version:  version,
		Revision: gitCommit,
	}

	if candidate != "" {
		versionJSON.Candidate = candidate
	}

	return versionJSON
}
