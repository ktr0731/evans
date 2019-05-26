// +build !dev

package app

import (
	"github.com/ktr0731/go-updater"
	"github.com/ktr0731/go-updater/brew"
	"github.com/ktr0731/go-updater/github"
)

var means = map[updater.MeansType]updater.MeansBuilder{
	brew.MeansTypeHomebrew:        brew.HomebrewMeans("ktr0731/evans", "evans"),
	github.MeansTypeGitHubRelease: github.GitHubReleaseMeans("ktr0731", "evans", github.TarGZIPDecompresser),
}
