package update_checker

import (
	"context"

	"github.com/ktr0731/evans/meta"
)

type UpdateChecker struct {
	client  client
	version string
}

func NewUpdateChecker() *UpdateChecker {
	return &UpdateChecker{
		client:  newGitHubClient(),
		version: meta.Version,
	}
}

func (u *UpdateChecker) IsLatest(ctx context.Context) (bool, error) {
	tag, err := u.client.FetchLatestTag(ctx)
	if err != nil {
		return false, err
	}
	return tag == u.version, nil
}
