// +build github

package app

import "github.com/ktr0731/go-updater/github"

func init() {
	// This package is enabled when Evans is built for GitHub Releases binary.
	github.IsGitHubReleasedBinary = true
}
