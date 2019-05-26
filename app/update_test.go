package app

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-version"
	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/go-updater"
	"github.com/ktr0731/go-updater/github"
	"github.com/pkg/errors"
	"gopkg.in/AlecAivazis/survey.v1"
)

type dummyMeans struct {
	dummyMeansBuilderOption

	updater.Means
}

func (m *dummyMeans) Installed(context.Context) bool {
	return m.installed
}

func (m *dummyMeans) Type() updater.MeansType {
	return updater.MeansType(m.typeName)
}

func (m *dummyMeans) LatestTag(context.Context) (*version.Version, error) {
	return version.NewSemver(m.version)
}

func (m *dummyMeans) Update(context.Context, *version.Version) error {
	time.Sleep(50 * time.Millisecond)
	return nil
}

type dummyMeansBuilderOption struct {
	// If err is nil, dummyMeansBuilder returns it as an error.
	err error

	// MeansBuilder params.
	installed bool
	typeName  string
	version   string
}

func dummyMeansBuilder(opt dummyMeansBuilderOption) updater.MeansBuilder {
	return func() (updater.Means, error) {
		if opt.err != nil {
			return nil, opt.err
		}
		return &dummyMeans{
			dummyMeansBuilderOption: opt,
		}, nil
	}
}

func Test_checkUpdate(t *testing.T) {
	oldMeans := means
	defer func() {
		means = oldMeans
	}()

	majorUpdatedVersion := fmt.Sprintf("v%d.%d.%d", meta.Version.Segments()[0]+1, 0, 0)
	minorUpdatedVersion := fmt.Sprintf("v%d.%d.%d", meta.Version.Segments()[0], meta.Version.Segments()[1]+1, 0)
	patchUpdatedVersion := fmt.Sprintf("v%d.%d.%d", meta.Version.Segments()[0], meta.Version.Segments()[1], meta.Version.Segments()[2]+1)

	cases := map[string]struct {
		// Package cache. If cacheGetFunc is not nil, replace real implementation by it.
		cacheGetFunc func() (*cache.Cache, error)

		meansBuilderOption dummyMeansBuilderOption

		// Meta config params.
		updateLevel string

		hasErr    bool
		noChanges bool
	}{
		"no updates if the means for installation is unavailable (in the case of means builder returns error)": {
			cacheGetFunc: func() (*cache.Cache, error) {
				return &cache.Cache{UpdateInfo: cache.UpdateInfo{InstalledBy: "dummy"}}, nil
			},
			meansBuilderOption: dummyMeansBuilderOption{
				err: errors.New("an error"),
			},
		},
		"no updates if no available means (in the case of MeansBuilder.Installed returns false )": {
			cacheGetFunc: func() (*cache.Cache, error) {
				return &cache.Cache{}, nil
			},
			meansBuilderOption: dummyMeansBuilderOption{
				installed: false,
			},
		},
		"an error returns if all candidate means return an error": {
			cacheGetFunc: func() (*cache.Cache, error) {
				return &cache.Cache{}, nil
			},
			meansBuilderOption: dummyMeansBuilderOption{
				err: errors.New("an error"),
			},
			hasErr: true,
		},
		"means found from condidate means": {
			cacheGetFunc: func() (*cache.Cache, error) {
				return &cache.Cache{}, nil
			},
			meansBuilderOption: dummyMeansBuilderOption{
				installed: true,
				typeName:  string(github.MeansTypeGitHubRelease),
				version:   patchUpdatedVersion,
			},
			updateLevel: "patch",
		},
		"the means for installation returns an error": {
			cacheGetFunc: func() (*cache.Cache, error) {
				return &cache.Cache{
					UpdateInfo: cache.UpdateInfo{
						InstalledBy: cache.MeansType(github.MeansTypeGitHubRelease),
					},
				}, nil
			},
			meansBuilderOption: dummyMeansBuilderOption{
				err: errors.New("an error"),
			},
		},
		"patch update available": {
			cacheGetFunc: func() (*cache.Cache, error) {
				return &cache.Cache{
					UpdateInfo: cache.UpdateInfo{
						InstalledBy: cache.MeansType(github.MeansTypeGitHubRelease),
					},
				}, nil
			},
			meansBuilderOption: dummyMeansBuilderOption{
				installed: true,
				typeName:  string(github.MeansTypeGitHubRelease),
				version:   patchUpdatedVersion,
			},
			updateLevel: "patch",
		},
		"minor update available": {
			cacheGetFunc: func() (*cache.Cache, error) {
				return &cache.Cache{
					UpdateInfo: cache.UpdateInfo{
						InstalledBy: cache.MeansType(github.MeansTypeGitHubRelease),
					},
				}, nil
			},
			meansBuilderOption: dummyMeansBuilderOption{
				installed: true,
				typeName:  string(github.MeansTypeGitHubRelease),
				version:   minorUpdatedVersion,
			},
			updateLevel: "minor",
		},
		"major update available": {
			cacheGetFunc: func() (*cache.Cache, error) {
				return &cache.Cache{
					UpdateInfo: cache.UpdateInfo{InstalledBy: cache.MeansType(github.MeansTypeGitHubRelease)},
				}, nil
			},
			meansBuilderOption: dummyMeansBuilderOption{
				installed: true,
				typeName:  string(github.MeansTypeGitHubRelease),
				version:   majorUpdatedVersion,
			},
			updateLevel: "major",
		},
		"checkUpdate fails because cache.Get returns an error": {
			cacheGetFunc: func() (*cache.Cache, error) { return nil, errors.New("an error") },
			hasErr:       true,
		},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			// TODO: remove this section.
			defer func() {
				if err := recover(); err != nil {
					t.Fatalf("panic occurred: %s", err)
				}
			}()

			if c.cacheGetFunc != nil {
				oldCacheGetFunc := cache.Get
				defer func() {
					cache.Get = oldCacheGetFunc
				}()
				cache.Get = c.cacheGetFunc
			}

			means = map[updater.MeansType]updater.MeansBuilder{
				// Assign a dummy means builder as the GitHub Releases means.
				github.MeansTypeGitHubRelease: dummyMeansBuilder(c.meansBuilderOption),
			}
			cache.CachedCache = &cache.Cache{}
			err := checkUpdate(context.Background(), &config.Config{Meta: &config.Meta{UpdateLevel: c.updateLevel}})
			if c.hasErr {
				if err == nil {
					t.Errorf("checkUpdate must return an error, but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("checkUpdate must not return an error, but got '%s'", err)
			}

			if c.noChanges {
				if diff := cmp.Diff(*cache.CachedCache, cache.Cache{}); diff != "" {
					t.Errorf("diff found:\n%s", diff)
				}
			}
		})
	}
}

func Test_processUpdate(t *testing.T) {
	oldSyscallExec := syscallExec
	defer func() {
		syscallExec = oldSyscallExec
	}()
	syscallExec = func(argv0 string, argv []string, envv []string) (err error) { return nil }

	oldSurveyAskOne := surveyAskOne
	defer func() {
		surveyAskOne = oldSurveyAskOne
	}()

	oldVersion := meta.Version
	meta.Version = version.Must(version.NewSemver("0.1.0"))
	defer func() {
		meta.Version = oldVersion
	}()

	cases := map[string]struct {
		// Package cache. If cacheGetFunc is not nil, replace real implementation by it.
		cacheGetFunc func() (*cache.Cache, error)

		cfgMeta config.Meta

		surveyAskOneResult bool
	}{
		"askOne prompt returns yes": {
			cacheGetFunc: func() (*cache.Cache, error) {
				return &cache.Cache{
					UpdateInfo: cache.UpdateInfo{
						InstalledBy:   cache.MeansType(github.MeansTypeGitHubRelease),
						LatestVersion: "0.2.0",
					},
					Version: "0.1.0",
				}, nil
			},
			cfgMeta: config.Meta{
				UpdateLevel: "patch",
			},
			surveyAskOneResult: true,
		},
		"askOne prompt returns false": {
			cacheGetFunc: func() (*cache.Cache, error) {
				return &cache.Cache{
					UpdateInfo: cache.UpdateInfo{
						InstalledBy:   cache.MeansType(github.MeansTypeGitHubRelease),
						LatestVersion: "0.2.0",
					},
					Version: "0.1.0",
				}, nil
			},
			cfgMeta: config.Meta{
				UpdateLevel: "patch",
			},
			surveyAskOneResult: false,
		},
		"do nothing if cached version <= the current version": {
			cacheGetFunc: func() (*cache.Cache, error) {
				return &cache.Cache{
					UpdateInfo: cache.UpdateInfo{
						InstalledBy:   cache.MeansType(github.MeansTypeGitHubRelease),
						LatestVersion: "0.0.1",
					},
					Version: "0.1.0",
				}, nil
			},
			cfgMeta: config.Meta{
				UpdateLevel: "patch",
			},
			surveyAskOneResult: false,
		},
		"AutoUpdate enabled": {
			cacheGetFunc: func() (*cache.Cache, error) {
				return &cache.Cache{
					UpdateInfo: cache.UpdateInfo{
						InstalledBy:   cache.MeansType(github.MeansTypeGitHubRelease),
						LatestVersion: "0.2.0",
					},
					Version: "0.1.0",
				}, nil
			},
			cfgMeta: config.Meta{
				AutoUpdate:  true,
				UpdateLevel: "patch",
			},
			surveyAskOneResult: false,
		},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			if c.cacheGetFunc != nil {
				oldCacheGetFunc := cache.Get
				defer func() {
					cache.Get = oldCacheGetFunc
				}()
				cache.Get = c.cacheGetFunc
			}

			surveyAskOne = func(_ survey.Prompt, res interface{}, _ survey.Validator, _ ...survey.AskOpt) error {
				rv := reflect.Indirect(reflect.ValueOf(res))
				rv.SetBool(c.surveyAskOneResult)
				return nil
			}

			var buf bytes.Buffer
			processUpdate(context.Background(), &config.Config{Meta: &c.cfgMeta}, &buf)
		})
	}
}

func Test_update(t *testing.T) {
	updater := updater.New(meta.Version, &dummyMeans{dummyMeansBuilderOption: dummyMeansBuilderOption{version: "v1.0.0"}})
	err := update(context.Background(), ioutil.Discard, updater)
	if err != nil {
		t.Errorf("update must not return an error, but got '%s'", err)
	}
}

var expectedUpdateInfo = `
new update available:
  current version: 0.1.0
   latest version: 1.0.0

`

func Test_printUpdateInfo(t *testing.T) {
	var buf bytes.Buffer
	old := meta.Version
	defer func() {
		meta.Version = old
	}()
	meta.Version = version.Must(version.NewSemver("v0.1.0"))
	printUpdateInfo(&buf, "1.0.0")
	if diff := cmp.Diff(buf.String(), expectedUpdateInfo); diff != "" {
		t.Errorf("diff found:\n%s", diff)
	}
}

func Test_newUpdater(t *testing.T) {
	t.Run("unknown update level", func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("newUpdater must panic, but not occurred")
			}
		}()
		newUpdater(&config.Config{Meta: &config.Meta{UpdateLevel: "foo"}}, nil, nil)
	})
}
