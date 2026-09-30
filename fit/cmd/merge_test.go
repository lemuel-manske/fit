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

func TestDiffNoChanges(t *testing.T) {
	base := []string{"one", "two"}
	target := []string{"one", "two"}

	got := Diff(base, target)

	expected := []Hunk{}

	require.Equal(t, expected, got)
}

func TestDiffInsertion(t *testing.T) {
	base := []string{"one"}
	target := []string{"one", "two"}

	got := Diff(base, target)

	expected := []Hunk{
		{
			Start: 1,
			End:   1,
			Lines: []string{"two"},
		},
	}

	require.Equal(t, expected, got)
}

func TestDiffDeletion(t *testing.T) {
	base := []string{"one", "two"}
	target := []string{"one"}

	got := Diff(base, target)

	expected := []Hunk{
		{
			Start: 1,
			End:   2,
			Lines: []string{},
		},
	}

	require.Equal(t, expected, got)
}

func TestDiffSingleReplacement(t *testing.T) {
	base := []string{"one", "two"}
	target := []string{"one", "TWO"}

	got := Diff(base, target)

	expected := []Hunk{
		{
			Start: 1,
			End:   2,
			Lines: []string{"TWO"},
		},
	}

	require.Equal(t, expected, got)
}

func TestDiffMultipleReplacements(t *testing.T) {
	base := []string{"one", "two", "three"}
	target := []string{"ONE", "TWO", "three"}

	got := Diff(base, target)

	expected := []Hunk{
		{
			Start: 0,
			End:   2,
			Lines: []string{"ONE", "TWO"},
		},
	}

	require.Equal(t, expected, got)
}

func TestDiffIndependentChanges(t *testing.T) {
	base := []string{"one", "two"}
	target := []string{"ONE", "TWO"}

	got := Diff(base, target)

	expected := []Hunk{
		{
			Start: 0,
			End:   2,
			Lines: []string{"ONE", "TWO"},
		},
	}

	require.Equal(t, expected, got)
}

func TestMergeTextIndependentChanges(t *testing.T) {
	base := []byte("one\ntwo\n")
	ours := []byte("ONE\ntwo\n")
	theirs := []byte("one\nTWO\n")

	got := MergeText(base, ours, theirs)

	expected := []string{"ONE\n", "TWO\n"}

	require.Equal(t, expected, got.Lines)
}

func TestMergeTextConflictingChanges(t *testing.T) {
	base := []byte("one\ntwo\n")
	ours := []byte("ONE\nTwo\n")
	theirs := []byte("onE\nTWO\n")

	got := MergeText(base, ours, theirs)

	expected := []string{
		"<<<<<<< ours",
		"ONE\n",
		"Two\n",
		"=======",
		"onE\n",
		"TWO\n",
		">>>>>>> theirs",
	}

	require.Equal(t, expected, got.Lines)
	require.True(t, got.Conflict)
}

func TestMergeTextOursChanged(t *testing.T) {
	base := []byte("one\ntwo\n")
	ours := []byte("ONE\ntwo\n")
	theirs := []byte("one\ntwo\n")

	got := MergeText(base, ours, theirs)

	expected := []string{"ONE\n", "two\n"}

	require.Equal(t, expected, got.Lines)
}

func TestMergeTextTheirsChanged(t *testing.T) {
	base := []byte("one\ntwo\n")
	ours := []byte("one\ntwo\n")
	theirs := []byte("one\nTWO\n")

	got := MergeText(base, ours, theirs)

	expected := []string{"one\n", "TWO\n"}

	require.Equal(t, expected, got.Lines)
}

func TestMergeTextBothChangedToSameContent(t *testing.T) {
	base := []byte("one\ntwo\n")
	ours := []byte("ONE\ntwo\n")
	theirs := []byte("ONE\ntwo\n")

	got := MergeText(base, ours, theirs)

	expected := []string{"ONE\n", "two\n"}

	require.Equal(t, expected, got.Lines)
}

func TestMergeIndependentChanges(t *testing.T) {
	base := []byte("one\ntwo\n")
	ours := []byte("ONE\ntwo\n")
	theirs := []byte("one\nTWO\n")

	got := MergeText(base, ours, theirs)

	expected := []string{"ONE\n", "TWO\n"}

	require.Equal(t, expected, got.Lines)
}

func TestMergeCRLFChanges(t *testing.T) {
	base := []byte("one\r\ntwo\r\n")
	ours := []byte("ONE\r\ntwo\r\n")
	theirs := []byte("one\r\nTWO\r\n")

	got := MergeText(base, ours, theirs)

	expected := []string{"ONE\r\n", "TWO\r\n"}

	require.Equal(t, expected, got.Lines)
}

func TestMergeLFChanges(t *testing.T) {
	base := []byte("one\ntwo\n")
	ours := []byte("ONE\ntwo\n")
	theirs := []byte("one\nTWO\n")

	got := MergeText(base, ours, theirs)

	expected := []string{"ONE\n", "TWO\n"}

	require.Equal(t, expected, got.Lines)
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
