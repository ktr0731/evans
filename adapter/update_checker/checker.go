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
		version: meta.Version.String(),
	}
}

type ReleaseTag struct {
	LatestVersion   string
	CurrentIsLatest bool
}

func (u *UpdateChecker) Check(ctx context.Context) (*ReleaseTag, error) {
	tag, err := u.client.FetchLatestTag(ctx)
	if err != nil {
		return nil, err
	}
	return &ReleaseTag{
		LatestVersion:   tag,
		CurrentIsLatest: tag == u.version,
	}, nil
}
