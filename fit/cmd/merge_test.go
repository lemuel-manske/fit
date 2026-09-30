package cmd

import (
	"testing"

	"fit/fit/internal"
	utils "fit/fit/test_utils"

	"github.com/stretchr/testify/require"
)

func TestMergeFile(t *testing.T) {
	a := internal.Hash("A")
	b := internal.Hash("B")
	c := internal.Hash("C")

	tests := []struct {
		name     string
		base     *internal.Hash
		ours     *internal.Hash
		theirs   *internal.Hash
		expected MergeDecision
	}{
		{
			name:   "unchanged on both sides",
			base:   &a,
			ours:   &a,
			theirs: &a,
			expected: MergeDecision{
				Blob: &a,
			},
		},
		{
			name:   "only ours changed",
			base:   &a,
			ours:   &b,
			theirs: &a,
			expected: MergeDecision{
				Blob: &b,
			},
		},
		{
			name:   "only theirs changed",
			base:   &a,
			ours:   &a,
			theirs: &b,
			expected: MergeDecision{
				Blob: &b,
			},
		},
		{
			name:   "both changed to same blob",
			base:   &a,
			ours:   &b,
			theirs: &b,
			expected: MergeDecision{
				Blob: &b,
			},
		},
		{
			name:   "both changed differently",
			base:   &a,
			ours:   &b,
			theirs: &c,
			expected: MergeDecision{
				Conflict: true,
			},
		},
		{
			name:   "both deleted",
			base:   &a,
			ours:   nil,
			theirs: nil,
			expected: MergeDecision{
				Delete: true,
			},
		},
		{
			name:   "only ours deleted",
			base:   &a,
			ours:   nil,
			theirs: &a,
			expected: MergeDecision{
				Delete: true,
			},
		},
		{
			name:   "only theirs deleted",
			base:   &a,
			ours:   &a,
			theirs: nil,
			expected: MergeDecision{
				Delete: true,
			},
		},
		{
			name:   "ours modified theirs deleted",
			base:   &a,
			ours:   &b,
			theirs: nil,
			expected: MergeDecision{
				Conflict: true,
			},
		},
		{
			name:   "ours deleted theirs modified",
			base:   &a,
			ours:   nil,
			theirs: &b,
			expected: MergeDecision{
				Conflict: true,
			},
		},
		{
			name:   "added only by ours",
			base:   nil,
			ours:   &b,
			theirs: nil,
			expected: MergeDecision{
				Blob: &b,
			},
		},
		{
			name:   "added only by theirs",
			base:   nil,
			ours:   nil,
			theirs: &b,
			expected: MergeDecision{
				Blob: &b,
			},
		},
		{
			name:   "added same blob on both sides",
			base:   nil,
			ours:   &b,
			theirs: &b,
			expected: MergeDecision{
				Blob: &b,
			},
		},
		{
			name:   "added different blobs on both sides",
			base:   nil,
			ours:   &b,
			theirs: &c,
			expected: MergeDecision{
				Conflict: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TakeMergeDecision(tt.base, tt.ours, tt.theirs)

			require.Equal(t, tt.expected, got)
		})
	}
}

func TestFastForwardMovesHeadAndWorkingTreeToDescendant(t *testing.T) {
	dir := t.TempDir()

	err := Init(dir, "test")
	require.NoError(t, err)

	_, err = utils.WriteFile(dir, "a.txt", "A")
	require.NoError(t, err)

	require.NoError(t, Add(dir, "a.txt"))

	a, err := CommitChanges(dir, "A")
	require.NoError(t, err)

	_, err = utils.WriteFile(dir, "a.txt", "B")
	require.NoError(t, err)

	require.NoError(t, Add(dir, "a.txt"))

	b, err := CommitChanges(dir, "B")
	require.NoError(t, err)

	require.NoError(t, internal.WriteHEAD(dir, a))

	_, err = utils.WriteFile(dir, "a.txt", "A")
	require.NoError(t, err)

	err = FastForward(dir, b)
	require.NoError(t, err)

	head, err := internal.ReadHEAD(dir)
	require.NoError(t, err)
	require.Equal(t, b, head)

	content, err := utils.ReadFile(dir, "a.txt")
	require.NoError(t, err)

	require.Equal(t, "B", string(content))
}

func TestFastForwardCannotMoveHeadBackwardsButCheckoutCan(t *testing.T) {
	dir := t.TempDir()

	err := Init(dir, "test")
	require.NoError(t, err)

	_, err = utils.WriteFile(dir, "a.txt", "A")
	require.NoError(t, err)

	require.NoError(t, Add(dir, "a.txt"))

	a, err := CommitChanges(dir, "A")
	require.NoError(t, err)

	_, err = utils.WriteFile(dir, "a.txt", "B")
	require.NoError(t, err)

	require.NoError(t, Add(dir, "a.txt"))

	b, err := CommitChanges(dir, "B")
	require.NoError(t, err)

	// A <- B
	// HEAD = B

	// fast-forward cannot go back
	err = FastForward(dir, a)
	require.Error(t, err)

	head, err := internal.ReadHEAD(dir)
	require.NoError(t, err)
	require.Equal(t, b, head)

	err = Checkout(dir, a)
	require.NoError(t, err)

	head, err = internal.ReadHEAD(dir)
	require.NoError(t, err)
	require.Equal(t, a, head)

	content, err := utils.ReadFile(dir, "a.txt")
	require.NoError(t, err)

	require.Equal(t, "A", string(content))
}
