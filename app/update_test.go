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
		cache *cache.Cache

		meansBuilderOption dummyMeansBuilderOption

		// Meta config params.
		updateLevel string

		hasErr    bool
		noChanges bool
	}{
		"no updates if the means for installation is unavailable (in the case of means builder returns error)": {
			cache: &cache.Cache{UpdateInfo: cache.UpdateInfo{InstalledBy: "dummy"}},
			meansBuilderOption: dummyMeansBuilderOption{
				err: errors.New("an error"),
			},
		},
		"no updates if no available means (in the case of MeansBuilder.Installed returns false )": {
			cache: &cache.Cache{},
			meansBuilderOption: dummyMeansBuilderOption{
				installed: false,
			},
		},
		"an error returns if all candidate means return an error": {
			cache: &cache.Cache{},
			meansBuilderOption: dummyMeansBuilderOption{
				err: errors.New("an error"),
			},
			hasErr: true,
		},
		"means found from condidate means": {
			cache: &cache.Cache{
				SaveFunc: func() error { return nil },
			},
			meansBuilderOption: dummyMeansBuilderOption{
				installed: true,
				typeName:  string(github.MeansTypeGitHubRelease),
				version:   patchUpdatedVersion,
			},
			updateLevel: "patch",
		},
		"the means for installation returns an error": {
			cache: &cache.Cache{
				UpdateInfo: cache.UpdateInfo{
					InstalledBy: cache.MeansType(github.MeansTypeGitHubRelease),
				},
			},
			meansBuilderOption: dummyMeansBuilderOption{
				err: errors.New("an error"),
			},
		},
		"patch update available": {
			cache: &cache.Cache{
				SaveFunc: func() error { return nil },
				UpdateInfo: cache.UpdateInfo{
					InstalledBy: cache.MeansType(github.MeansTypeGitHubRelease),
				},
			},
			meansBuilderOption: dummyMeansBuilderOption{
				installed: true,
				typeName:  string(github.MeansTypeGitHubRelease),
				version:   patchUpdatedVersion,
			},
			updateLevel: "patch",
		},
		"minor update available": {
			cache: &cache.Cache{
				SaveFunc: func() error { return nil },
				UpdateInfo: cache.UpdateInfo{
					InstalledBy: cache.MeansType(github.MeansTypeGitHubRelease),
				},
			},
			meansBuilderOption: dummyMeansBuilderOption{
				installed: true,
				typeName:  string(github.MeansTypeGitHubRelease),
				version:   minorUpdatedVersion,
			},
			updateLevel: "minor",
		},
		"major update available": {
			cache: &cache.Cache{
				SaveFunc:   func() error { return nil },
				UpdateInfo: cache.UpdateInfo{InstalledBy: cache.MeansType(github.MeansTypeGitHubRelease)},
			},
			meansBuilderOption: dummyMeansBuilderOption{
				installed: true,
				typeName:  string(github.MeansTypeGitHubRelease),
				version:   majorUpdatedVersion,
			},
			updateLevel: "major",
		},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			means = map[updater.MeansType]updater.MeansBuilder{
				// Assign a dummy means builder as the GitHub Releases means.
				github.MeansTypeGitHubRelease: dummyMeansBuilder(c.meansBuilderOption),
			}
			err := checkUpdate(context.Background(), &config.Config{Meta: &config.Meta{UpdateLevel: c.updateLevel}}, c.cache)
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
				c, err := cache.Get()
				if err != nil {
					t.Fatalf("cache.Get must not return an error, but got '%s'", err)
				}
				if diff := cmp.Diff(*c, cache.Cache{}); diff != "" {
					t.Errorf("diff found:\n%s", diff)
				}
			}
		})
	}
}

func Test_processUpdate(t *testing.T) {
	oldMeans := means
	defer func() {
		means = oldMeans
	}()

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
		cache *cache.Cache

		cfgMeta config.Meta

		surveyAskOneResult bool
	}{
		"askOne prompt returns yes": {
			cache: &cache.Cache{
				SaveFunc: func() error { return nil },
				UpdateInfo: cache.UpdateInfo{
					InstalledBy:   cache.MeansType(github.MeansTypeGitHubRelease),
					LatestVersion: "0.2.0",
				},
				Version: "0.1.0",
			},
			cfgMeta: config.Meta{
				UpdateLevel: "patch",
			},
			surveyAskOneResult: true,
		},
		"askOne prompt returns false": {
			cache: &cache.Cache{
				UpdateInfo: cache.UpdateInfo{
					InstalledBy:   cache.MeansType(github.MeansTypeGitHubRelease),
					LatestVersion: "0.2.0",
				},
				Version: "0.1.0",
			},
			cfgMeta: config.Meta{
				UpdateLevel: "patch",
			},
			surveyAskOneResult: false,
		},
		"do nothing if cached version <= the current version": {
			cache: &cache.Cache{
				SaveFunc: func() error { return nil },
				UpdateInfo: cache.UpdateInfo{
					InstalledBy:   cache.MeansType(github.MeansTypeGitHubRelease),
					LatestVersion: "0.0.1",
				},
				Version: "0.1.0",
			},
			cfgMeta: config.Meta{
				UpdateLevel: "patch",
			},
			surveyAskOneResult: false,
		},
		"AutoUpdate enabled": {
			cache: &cache.Cache{
				SaveFunc: func() error { return nil },
				UpdateInfo: cache.UpdateInfo{
					InstalledBy:   cache.MeansType(github.MeansTypeGitHubRelease),
					LatestVersion: "0.2.0",
				},
				Version: "0.1.0",
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
			means = map[updater.MeansType]updater.MeansBuilder{
				// Assign a dummy means builder as the GitHub Releases means.
				github.MeansTypeGitHubRelease: dummyMeansBuilder(dummyMeansBuilderOption{
					installed: true,
					typeName:  string(c.cache.UpdateInfo.InstalledBy),
					version:   c.cache.UpdateInfo.LatestVersion,
				}),
			}

			surveyAskOne = func(_ survey.Prompt, res interface{}, _ survey.Validator, _ ...survey.AskOpt) error {
				rv := reflect.Indirect(reflect.ValueOf(res))
				rv.SetBool(c.surveyAskOneResult)
				return nil
			}

			var buf bytes.Buffer
			if err := processUpdate(context.Background(), &config.Config{Meta: &c.cfgMeta}, &buf, c.cache); err != nil {
				t.Errorf("must not return an error, but got '%s'", err)
			}
		})
	}
}

func Test_update(t *testing.T) {
	updater := updater.New(meta.Version, &dummyMeans{dummyMeansBuilderOption: dummyMeansBuilderOption{version: "v1.0.0"}})
	err := update(context.Background(), ioutil.Discard, updater, &cache.Cache{SaveFunc: func() error { return nil }})
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
