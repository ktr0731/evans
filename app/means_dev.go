// +build dev

// means_dev.go treats the default means for updating to a dummy.
// To enable dev build, run the following command.
//
//   $ make build-dev
//
// If $DEV_LATEST_VERSION is enabled, the dummy means returns it as the latest version.
// Therefore, you can change the latest version to arbitrary values.
// If it is empty, the dummy returns the current version as the latest version.

package app

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/logger"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/go-updater"
)

var dummyLatestVersion string

const meansTypeDummy = updater.MeansType("dummy")

func init() {
	if dummyLatestVersion = os.Getenv("DEV_LATEST_VERSION"); dummyLatestVersion == "" {
		dummyLatestVersion = meta.Version.String()
	}
	cache.Get = func() (*cache.Cache, error) {
		logger.Printf("dummy latest version: %s", dummyLatestVersion)
		return &cache.Cache{
			Version: meta.Version.String(),
			UpdateInfo: cache.UpdateInfo{
				LatestVersion: dummyLatestVersion,
				InstalledBy:   cache.MeansType(meansTypeDummy),
			},
		}, nil
	}
}

var means = map[updater.MeansType]updater.MeansBuilder{
	meansTypeDummy: func() (updater.Means, error) {
		return &dummyMeans{}, nil
	},
}

type dummyMeans struct {
	updater.Means
}

func (m *dummyMeans) Installed(context.Context) bool {
	return true
}

func (m *dummyMeans) Type() updater.MeansType {
	return meansTypeDummy
}

func (m *dummyMeans) LatestTag(context.Context) (*version.Version, error) {
	return version.NewSemver(dummyLatestVersion)
}

func (m *dummyMeans) Update(context.Context, *version.Version) error {
	time.Sleep(1 * time.Second)
	return nil
}
