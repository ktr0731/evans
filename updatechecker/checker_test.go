package updatechecker

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockClient struct {
	tag string
	err error
}

func (c *mockClient) FetchLatestTag(ctx context.Context) (string, error) {
	return c.tag, c.err
}

func TestUpdateChecker(t *testing.T) {
	client := &mockClient{
		tag: "0.1.0",
	}
	u := &UpdateChecker{
		client:  client,
		version: "0.1.0",
	}

	t.Run("latest", func(t *testing.T) {
		tag, err := u.Check(context.TODO())
		require.NoError(t, err)
		require.True(t, tag.CurrentIsLatest)
		require.Equal(t, "0.1.0", tag.LatestVersion)
	})

	t.Run("old", func(t *testing.T) {
		client.tag = "0.1.1"
		tag, err := u.Check(context.TODO())
		require.NoError(t, err)
		require.False(t, tag.CurrentIsLatest)
		require.Equal(t, "0.1.1", tag.LatestVersion)
	})

	t.Run("error", func(t *testing.T) {
		client.err = errors.New("an error")
		_, err := u.Check(context.TODO())
		require.Error(t, err)
	})
}
