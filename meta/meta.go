// Package meta provides information about Evans itself.
package meta

import version "github.com/hashicorp/go-version"

// AppName defines the application name.
const AppName = "evans"

var (
	// Version defines the version of Evans.
	Version = version.Must(version.NewSemver("0.7.3"))
)
