package meta

import version "github.com/hashicorp/go-version"

const AppName = "evans"

var (
	Version = version.Must(version.NewSemver("0.10.4"))
)
