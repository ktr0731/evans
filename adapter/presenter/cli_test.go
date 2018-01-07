package presenter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCLIPresenter(t *testing.T) {
	presenter := NewCLIPresenter()

	t.Run("Call", func(t *testing.T) {
		_, err := presenter.Call()
		require.NoError(t, err)
	})
}
