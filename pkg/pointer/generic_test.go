package pointer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultOrValueOf(t *testing.T) {
	require.Equal(t, 1, DefaultOrValueOf((nil), 1))
	require.Equal(t, 10, DefaultOrValueOf(Of(10), 1))
}

func TestValueOf(t *testing.T) {
	require.Equal(t, 0, ValueOf((*int)(nil)))
	require.Equal(t, 10, ValueOf(Of(10)))
}
