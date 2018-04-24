package meta

import (
	semver "github.com/ktr0731/go-semver"
)

const AppName = "evans"

var (
	Version = semver.MustParse("0.3.0")
)
