package controller

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/ktr0731/evans/meta"
	semver "github.com/ktr0731/go-semver"
	updater "github.com/ktr0731/go-updater"
	"github.com/ktr0731/go-updater/brew"
	"github.com/ktr0731/go-updater/github"
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

func checkUpdate(ctx context.Context, cache *meta.Meta, errCh chan<- error) {
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != context.Canceled {
			errCh <- err
			return
		}
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
	case meta.MeansTypeGitHubRelease:
		m, err = updater.NewMeans(github.GitHubReleaseMeans("ktr0731", "evans"))
	case meta.MeansTypeHomeBrew:
		m, err = updater.NewMeans(brew.HomeBrewMeans("ktr0731/evans", "evans"))
	}

	// if ErrUnavailable, user installed Evans by manually, ignore
	if err != nil && err != updater.ErrUnavailable {
		errCh <- err
		return
	}

	u := updater.New(meta.Version, m)
	updatable, latest, err := u.Updatable(ctx)
	if err != nil {
		errCh <- err
		return
	}
	if updatable {
		if err := meta.SetUpdateInfo(latest); err != nil {
			errCh <- err
			return
		}
	}

	errCh <- nil
	return
}

func update(ctx context.Context, infoWriter io.Writer, updater *updater.Updater) error {
	errCh := make(chan error, 1)
	go func(ctx context.Context, errCh chan<- error) {
		errCh <- updater.Update(ctx)
	}(ctx, errCh)

	s := spin.New()
	for {
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
			// update successful
			fmt.Fprintf(infoWriter, "\r             \r✔ updated!\n\n")
			return meta.Clear()
		case <-ctx.Done():
			if ctx.Err() != context.Canceled {
				return ctx.Err()
			}
			return nil
		default:
			fmt.Fprintf(infoWriter, "\r%s updating...", s.Next())
			time.Sleep(100 * time.Millisecond)
		}
	}
}

var updateInfoFormat string = `
new update available:
  current version: %s
  latest version:  %s
`

var commandText = `  $ brew upgrade evans
  $ go get -u github.com/ktr0731/evans
  $ curl -sL https://github.com/ktr0731/evans/releases/download/%s/evans_%s_%s.tar.gz | tar xf -
`

func printUpdateInfo(w io.Writer, latest string) {
	fmt.Fprintf(w, updateInfoFormat+"\n", meta.Version, latest)
}

func printUpdateInfoWithCommandText(w io.Writer, latest string) {
	sb := &strings.Builder{}
	printUpdateInfo(sb, latest)
	fmt.Fprintf(sb, commandText, latest, runtime.GOOS, runtime.GOARCH)
	fmt.Fprintln(w, sb)
}

func newUpdater(cache *meta.Meta) (updater.Means, error) {
	switch cache.InstalledBy {
	case meta.MeansTypeGitHubRelease:
		m, err := updater.NewMeans(github.GitHubReleaseMeans("ktr0731", "evans"))
		if err != nil {
			return nil, err
		}
		m.(*github.GitHubClient).Decompresser = github.TarDecompresser
		return m, nil
	case meta.MeansTypeHomeBrew:
		return updater.NewMeans(brew.HomeBrewMeans("ktr0731/evans", "evans"))
	}
	return nil, updater.ErrUnavailable
}
