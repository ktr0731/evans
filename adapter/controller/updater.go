package controller

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/meta"
	semver "github.com/ktr0731/go-semver"
	updater "github.com/ktr0731/go-updater"
	"github.com/ktr0731/go-updater/brew"
	"github.com/ktr0731/go-updater/github"
	"github.com/pkg/errors"
	spin "github.com/tj/go-spin"
)

func joinCommandText(sv string, builders ...updater.MeansBuilder) (string, error) {
	sb := &strings.Builder{}
	v := semver.MustParse(sv)
	for _, b := range builders {
		m, err := b()
		if err != nil {
			return "", err
		}
		fmt.Fprintf(sb, "  $ %s\n", m.CommandText(v))
	}
	fmt.Fprintln(sb)
	return sb.String(), nil
}

func checkUpdate(ctx context.Context, cfg *config.Config, cache *meta.Meta, errCh chan<- error) {
	go func() {
		<-ctx.Done()
		errCh <- nil
		return
	}()

	var m updater.Means
	var err error
	switch cache.InstalledBy {
	case meta.MeansTypeUndefined:
		m, err = updater.SelectAvailableMeansFrom(
			ctx,
			github.GitHubReleaseMeans("ktr0731", "evans"),
			brew.HomeBrewMeans("ktr0731/evans", "evans"),
		)
		if err := meta.SetInstalledBy(meta.MeansType(m.Type())); err != nil {
			errCh <- err
			return
		}
	default:
		m, err = newMeans(cache)
	}

	// if ErrUnavailable, user installed Evans by manually, ignore
	if err == updater.ErrUnavailable {
		errCh <- nil
		return
	} else if err != nil {
		errCh <- errors.Wrapf(err, "failed to instantiate new means, installed by %s", cache.InstalledBy)
		return
	}

	u := newUpdater(cfg, meta.Version, m)
	updatable, latest, err := u.Updatable(ctx)
	if errors.Cause(err) != context.Canceled && err != nil {
		errCh <- errors.Wrap(err, "failed to check updatable")
		return
	}
	if updatable {
		if err := meta.SetUpdateInfo(latest); err != nil {
			errCh <- errors.Wrap(err, "failed to write update info to cache")
			return
		}
	}

	errCh <- nil
	return
}

func update(infoWriter io.Writer, updater *updater.Updater) error {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- updater.Update(ctx)
	}()

	s := spin.New()
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
			return meta.Clear()
		default:
			fmt.Fprintf(infoWriter, "\r%s updating...", s.Next())
			time.Sleep(100 * time.Millisecond)
		}
	}
}

var updateInfoFormat string = `
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
func newMeans(cache *meta.Meta) (updater.Means, error) {
	switch cache.InstalledBy {
	case meta.MeansType(github.MeansTypeGitHubRelease):
		m, err := updater.NewMeans(github.GitHubReleaseMeans("ktr0731", "evans"))
		if err != nil {
			return nil, err
		}
		m.(*github.GitHubClient).Decompresser = github.TarDecompresser
		return m, nil
	case meta.MeansType(brew.MeansTypeHomeBrew):
		return updater.NewMeans(brew.HomeBrewMeans("ktr0731/evans", "evans"))
	}
	return nil, updater.ErrUnavailable
}
