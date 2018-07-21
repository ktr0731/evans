package controller

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/meta"
	semver "github.com/ktr0731/go-semver"
	updater "github.com/ktr0731/go-updater"
	"github.com/ktr0731/go-updater/brew"
	"github.com/ktr0731/go-updater/github"
	"github.com/pkg/errors"
	spin "github.com/tj/go-spin"
)

func checkUpdate(ctx context.Context, cfg *config.Config, c *cache.Cache) error {
	errCh := make(chan error, 1)
	go func() {
		var m updater.Means
		var err error
		switch c.InstalledBy {
		case cache.MeansTypeUndefined:
			m, err = updater.SelectAvailableMeansFrom(
				ctx,
				brew.HomebrewMeans("ktr0731/evans", "evans"),
				github.GitHubReleaseMeans("ktr0731", "evans", github.TarDecompresser),
			)
			// if ErrUnavailable, user installed Evans by manually, ignore
			if err == updater.ErrUnavailable {
				errCh <- nil
				return
			} else if err != nil {
				errCh <- errors.Wrap(err, "failed to instantiate new means, available means not found")
				return
			}
			if err := cache.SetInstalledBy(cache.MeansType(m.Type())); err != nil {
				errCh <- err
				return
			}
		default:
			m, err = newMeans(c)
			if err == updater.ErrUnavailable {
				errCh <- nil
				return
			} else if err != nil {
				errCh <- errors.Wrapf(err, "failed to instantiate new means, installed by %s", c.InstalledBy)
				return
			}
		}

		u := newUpdater(cfg, meta.Version, m)
		updatable, latest, err := u.Updatable(ctx)
		if errors.Cause(err) != context.Canceled && err != nil {
			errCh <- errors.Wrap(err, "failed to check updatable")
			return
		}
		if updatable {
			if err := cache.SetUpdateInfo(latest); err != nil {
				errCh <- errors.Wrap(err, "failed to write update info to cache")
				return
			}
		}

		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func update(ctx context.Context, infoWriter io.Writer, updater *updater.Updater) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- updater.Update(ctx)
	}()

	s := spin.New()
	tick := time.Tick(100 * time.Millisecond)
	for {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			if errors.Cause(err) != context.Canceled && err != nil {
				return errors.Wrap(err, "failed to update Evans")
			}
			// update successful
			fmt.Fprintf(infoWriter, "\r             \r✔ updated!\n\n")
			return cache.Clear()
		case <-tick:
			fmt.Fprintf(infoWriter, "\r%s updating...", s.Next())
		}
	}
}

var updateInfoFormat = `
new update available:
  current version: %s
   latest version: %s

`

func printUpdateInfo(w io.Writer, latest string) {
	fmt.Fprintf(w, updateInfoFormat, meta.Version, latest)
}

// newUpdater creates new updater from cached information.
// updater checks whether UpdateIf is true or false
// to display update information to the user.
func newUpdater(cfg *config.Config, v *semver.Version, m updater.Means) *updater.Updater {
	u := updater.New(v, m)
	switch cfg.Meta.UpdateLevel {
	case "patch":
		u.UpdateIf = updater.FoundPatchUpdate
	case "minor":
		u.UpdateIf = updater.FoundMinorUpdate
	case "major":
		u.UpdateIf = updater.FoundMajorUpdate
	default:
		panic("unknown update level")
	}
	return u
}

// newMeans creates new available means from cached infomation.
// if InstalledBy is MeansTypeUndefined, returns updater.ErrUnavailable.
func newMeans(c *cache.Cache) (updater.Means, error) {
	switch c.InstalledBy {
	case cache.MeansType(github.MeansTypeGitHubRelease):
		return updater.NewMeans(github.GitHubReleaseMeans("ktr0731", "evans", github.TarDecompresser))
	case cache.MeansType(brew.MeansTypeHomebrew):
		return updater.NewMeans(brew.HomebrewMeans("ktr0731/evans", "evans"))
	}
	return nil, updater.ErrUnavailable
}
