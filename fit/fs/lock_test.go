package fs

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLock(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, WriteFitDir(dir))
	first, err := Lock(dir, "serve", context.Background(), false)
	require.NoError(t, err)
	_, err = Lock(dir, "serve", context.Background(), false)
	require.Error(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = Lock(dir, "serve", ctx, true)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.NoError(t, first.Close())
	second, err := Lock(dir, "serve", context.Background(), false)
	require.NoError(t, err)
	require.NoError(t, second.Close())
}
