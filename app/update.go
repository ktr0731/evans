package app

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/ktr0731/evans/cache"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/logger"
	"github.com/ktr0731/evans/meta"
	"github.com/ktr0731/go-updater"
	"github.com/pkg/errors"
	"github.com/tj/go-spin"
	"gopkg.in/AlecAivazis/survey.v1"
)

// checkUpdate checks whether an update exists. Update checking is instructed by following steps:
//   1. Extract the application cache.
//   2. If install means is known, use it as an update means.
//      If install means is unknown, checkUpdate selects an available means from candidates.
//   3. Check whether update exists. It it is found, cache the latest version.
func checkUpdate(ctx context.Context, cfg *config.Config) error {
	c, err := cache.Get()
	if err != nil {
		return errors.Wrap(err, "failed to get the cache")
	}

	var m updater.Means
	switch c.UpdateInfo.InstalledBy {
	case cache.MeansTypeUndefined:
		meansBuilders := make([]updater.MeansBuilder, 0, len(means))
		for _, mb := range means {
			meansBuilders = append(meansBuilders, mb)
		}
		m, err = updater.SelectAvailableMeansFrom(ctx, meansBuilders...)
		// if ErrUnavailable, user installed Evans by manually, ignore.
		if err == updater.ErrUnavailable {
			return nil
		} else if err != nil {
			return errors.Wrap(err, "failed to instantiate new means, available means not found")
		}
		c.UpdateInfo.InstalledBy = cache.MeansType(m.Type())
		if err := c.Save(); err != nil {
			return errors.Wrap(err, "failed to save a cache")
		}
	default:
		// If specified means builder is not found, skip update.
		mb, ok := means[updater.MeansType(c.UpdateInfo.InstalledBy)]
		if !ok {
			return nil
		}
		m, err = mb()
		if err != nil {
			logger.Printf("failed to build a new means: %s", err)
			return nil
		}
	}

	u := newUpdater(cfg, meta.Version, m)
	updatable, latest, err := u.Updatable(ctx)
	if errors.Cause(err) != context.Canceled && err != nil {
		return errors.Wrap(err, "failed to check updatable")
	}
	if updatable {
		c.UpdateInfo.LatestVersion = latest.String()
		if err := c.Save(); err != nil {
			return errors.Wrap(err, "failed to save a cache")
		}
	}
	return nil
}

var (
	syscallExec  = syscall.Exec
	surveyAskOne = survey.AskOne
)

// processUpdate checks new changes and updates Evans in accordance with user's selection.
// If config.Meta.AutoUpdate enabled, processUpdate is called asynchronously.
// Other than, processUpdate is called synchronously.
func processUpdate(ctx context.Context, cfg *config.Config, w io.Writer) error {
	c, err := cache.Get()
	if err != nil {
		return errors.Wrap(err, "failed to get the cache content")
	}
	if !c.UpdateInfo.UpdateAvailable() {
		return nil
	}

	// If cached version is less than or equal to current version, ignore it.
	v := version.Must(version.NewSemver(c.UpdateInfo.LatestVersion))
	if v.LessThan(meta.Version) || v.Equal(meta.Version) {
		c.UpdateInfo = cache.UpdateInfo{}
		if err := c.Save(); err != nil {
			return errors.Wrap(err, "failed to clear the cache")
		}
		return nil
	}

	// Instantiate the means.
	// If ErrUnavailable, user installed Evans by manually, ignore.
	mb, ok := means[updater.MeansType(c.UpdateInfo.InstalledBy)]
	if !ok {
		return nil
	}
	m, err := mb()
	if err != nil {
		logger.Printf("failed to build a new means: %s", err)
		return nil
	}

	// If auto update is enabled, do process update without user's confirmation.
	if cfg.Meta.AutoUpdate {
		// If canceled, ignore and return
		err := update(ctx, ioutil.Discard, newUpdater(cfg, meta.Version, m))
		if errors.Cause(err) == context.Canceled {
			return nil
		}
		return err
	}

	// If auto update is disabled, isplay update info

	printUpdateInfo(w, c.UpdateInfo.LatestVersion)

	var yes bool
	if err := surveyAskOne(&survey.Confirm{
		Message: "update?",
	}, &yes, nil); err != nil {
		return errors.Wrap(err, "failed to get survey answer")
	}
	if !yes {
		return nil
	}

	// If canceled, ignore and return
	err = update(ctx, w, newUpdater(cfg, meta.Version, m))
	if errors.Cause(err) == context.Canceled {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "failed to update binary")
	}

	// restart Evans
	if err := syscallExec(os.Args[0], os.Args, os.Environ()); err != nil {
		return errors.Wrapf(err, "failed to exec the command: args=%s", os.Args)
	}

	return nil
}

// update updates Evans to the latest version. If interrupted by a key, update will be canceled.
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
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
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
			c, err := cache.Get()
			if err != nil {
				return errors.Wrap(err, "failed to get the cache")
			}
			c.UpdateInfo = cache.UpdateInfo{}
			if err := c.Save(); err != nil {
				return errors.Wrap(err, "failed to clear the cache")
			}
			return nil
		case <-tick.C:
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

// newUpdater creates new updater from cached information. updater checks whether UpdateIf is true or false
// to display update information to the user.
func newUpdater(cfg *config.Config, v *version.Version, m updater.Means) *updater.Updater {
	u := updater.New(v, m)
	switch cfg.Meta.UpdateLevel {
	case "patch":
		u.UpdateIf = updater.FoundPatchUpdate
	case "minor":
		u.UpdateIf = updater.FoundMinorUpdate
	case "major":
		u.UpdateIf = updater.FoundMajorUpdate
	default:
		panic(fmt.Sprintf("unknown update level: '%s'", cfg.Meta.UpdateLevel))
	}
	return u
}
