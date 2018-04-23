package controller

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ktr0731/evans/meta"
	updater "github.com/ktr0731/go-updater"
	"github.com/ktr0731/go-updater/brew"
	"github.com/ktr0731/go-updater/github"
	spin "github.com/tj/go-spin"
)

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

	// write result

	errCh <- nil
	return
}

func update(ctx context.Context, infoWriter io.Writer, updater *updater.Updater) error {
	errCh := make(chan error, 1)
	go func(errCh chan<- error) {
		// errCh <- c.updater.Update()
		time.Sleep(2 * time.Second)
		errCh <- nil
	}(errCh)

	s := spin.New()
LOOP:
	for {
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
			break LOOP
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
	fmt.Fprintf(infoWriter, "\r             \r✔ updated!\n\n")
	return meta.Clear()
}
