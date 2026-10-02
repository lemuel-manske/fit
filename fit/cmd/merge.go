package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"fit/fit/internal"

	"path/filepath"
)

const (
	errNotAncestor     = "target commit is not a descendant of the current HEAD"
	errMergeInProgress = "merge in progress"
)

type Hunk struct {
	Start int

	End int

	Lines []string
}

type MergeHunk struct {
	Start int
	End   int

	Lines []string

	Conflict bool

	Ours   []string
	Theirs []string
}

type MergeDecision struct {
	Blob     *internal.Hash
	Delete   bool
	Conflict bool
}

type TextMergeResult struct {
	Lines    []string
	Conflict bool
}

func MergeAbort(dir string) error {
	if !internal.MergeInProgress(dir) {
		return fmt.Errorf("no merge in progress")
	}

	head, err := internal.ReadHEAD(dir)
	if err != nil {
		return err
	}

	if err := ForceCheckout(dir, head); err != nil {
		return err
	}

	index, err := internal.LoadIndex(dir)
	if err != nil {
		return err
	}

	clear(index.Entries)

	if err := internal.WriteIndex(dir, index); err != nil {
		return err
	}

	if err := internal.ClearMergeState(dir); err != nil {
		return err
	}

	return nil
}

// Merge merges the specified commit into the current HEAD.
func Merge(dir string, theirsID internal.CommitID) error {
	currMergeHEAD, err := internal.ReadMergeHEAD(dir)
	if err == nil {
		err = fmt.Errorf(
			"%s: %s is the current merge target",
			errMergeInProgress,
			currMergeHEAD,
		)

		return err
	}

	if !os.IsNotExist(err) {
		return err
	}

	oursID, err := internal.ReadHEAD(dir)
	if err != nil {
		return err
	}

	baseID, err := internal.Ancestor(dir, oursID, theirsID)
	if err != nil {
		return err
	}

	store := internal.NewCommitStore(dir)

	base, err := store.Get(baseID)
	if err != nil {
		return err
	}

	ours, err := store.Get(oursID)
	if err != nil {
		return err
	}

	theirs, err := store.Get(theirsID)
	if err != nil {
		return err
	}

	if err := internal.WriteMergeHEAD(dir, theirsID); err != nil {
		return err
	}

	if err := internal.WriteMergeBase(dir, baseID); err != nil {
		return err
	}

	paths := mergePaths(base.Files, ours.Files, theirs.Files)

	blobStore := internal.NewBlobStore(dir)

	index, err := internal.LoadIndex(dir)
	if err != nil {
		return err
	}

	for _, path := range paths {
		if err := internal.ValidateRepoPath(path); err != nil {
			return err
		}

		index.Entries[path] = internal.IndexEntry{Conflict: true}
	}

	if err := internal.WriteIndex(dir, index); err != nil {
		return err
	}

	for _, path := range paths {
		baseHash := hashAt(base.Files, path)
		oursHash := hashAt(ours.Files, path)
		theirsHash := hashAt(theirs.Files, path)

		decision := TakeMergeDecision(baseHash, oursHash, theirsHash)

		absPath := filepath.Join(dir, path)

		if decision.Delete {
			if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
				return err
			}

			index.Entries[path] = internal.IndexEntry{Delete: true}

			continue
		}

		var content []byte

		conflict := false

		if decision.Blob != nil {
			content, err = blobStore.Get(*decision.Blob)
			if err != nil {
				return err
			}
		} else {
			baseContent, err := blobStore.Get(*baseHash)
			if err != nil {
				return err
			}

			oursContent, err := blobStore.Get(*oursHash)
			if err != nil {
				return err
			}

			theirsContent, err := blobStore.Get(*theirsHash)
			if err != nil {
				return err
			}

			var result TextMergeResult

			if oursHash == nil || theirsHash == nil {
				result = ApplyMergeHunks(nil, []MergeHunk{{
					Conflict: true,
					Ours:     splitLines(oursContent),
					Theirs:   splitLines(theirsContent),
				}})
			} else {
				result = MergeText(
					baseContent,
					oursContent,
					theirsContent,
				)
			}
			content = []byte(strings.Join(result.Lines, "\n"))

			conflict = result.Conflict
		}

		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(absPath, content, 0644); err != nil {
			return err
		}

		hash, err := blobStore.Put(content)

		if err != nil {
			return err
		}

		index.Entries[path] = internal.IndexEntry{Blob: string(hash), Conflict: conflict}
	}

	if err := internal.WriteIndex(dir, index); err != nil {
		return err
	}

	return nil
}

func hashAt(files map[string]internal.Hash, path string) *internal.Hash {
	hash, exists := files[path]
	if !exists {
		return nil
	}

	return &hash
}

func mergePaths(
	base,
	ours,
	theirs map[string]internal.Hash,
) []string {
	pathSet := make(map[string]struct{})

	for path := range base {
		pathSet[path] = struct{}{}
	}

	for path := range ours {
		pathSet[path] = struct{}{}
	}

	for path := range theirs {
		pathSet[path] = struct{}{}
	}

	paths := make([]string, 0, len(pathSet))

	for path := range pathSet {
		paths = append(paths, path)
	}

	slices.Sort(paths)

	return paths
}

func FastForward(dir string, target internal.CommitID) error {
	head, err := internal.ReadHEAD(dir)
	if err != nil {
		return err
	}

	ancestor, err := internal.Ancestor(dir, head, target)
	if err != nil {
		return err
	}

	if ancestor != head {
		return fmt.Errorf(
			"%s: %s is not an ancestor of %s",
			errNotAncestor,
			head,
			target,
		)
	}

	return Checkout(dir, target)
}

func TakeMergeDecision(
	base,
	ours,
	theirs *internal.Hash,
) MergeDecision {

	if sameHash(ours, theirs) {
		if ours == nil {
			return MergeDecision{
				Delete: true,
			}
		}

		return MergeDecision{
			Blob: ours,
		}
	}

	if sameHash(base, ours) {
		if theirs == nil {
			return MergeDecision{
				Delete: true,
			}
		}

		return MergeDecision{
			Blob: theirs,
		}
	}

	if sameHash(base, theirs) {
		if ours == nil {
			return MergeDecision{
				Delete: true,
			}
		}

		return MergeDecision{
			Blob: ours,
		}
	}

	return MergeDecision{
		Conflict: true,
	}
}

func MergeText(base, ours, theirs []byte) TextMergeResult {
	baseLines := splitLines(base)
	oursLines := splitLines(ours)
	theirsLines := splitLines(theirs)

	oursHunks := Diff(baseLines, oursLines)
	theirsHunks := Diff(baseLines, theirsLines)

	mergedHunks := MergeHunks(oursHunks, theirsHunks)

	return ApplyMergeHunks(baseLines, mergedHunks)
}

func MergeHunks(ours, theirs []Hunk) []MergeHunk {
	result := make([]MergeHunk, 0)

	i := 0
	j := 0

	for i < len(ours) || j < len(theirs) {

		if i >= len(ours) {
			for ; j < len(theirs); j++ {
				result = append(result, normalMergeHunk(theirs[j]))
			}

			break
		}

		if j >= len(theirs) {
			for ; i < len(ours); i++ {
				result = append(result, normalMergeHunk(ours[i]))
			}

			break
		}

		o := ours[i]
		t := theirs[j]

		switch {

		case hunksEqual(o, t):
			result = append(result, normalMergeHunk(o))

			i++
			j++

		case hunksOverlap(o, t):
			result = append(result, MergeHunk{
				Start:    min(o.Start, t.Start),
				End:      max(o.End, t.End),
				Conflict: true,
				Ours:     slices.Clone(o.Lines),
				Theirs:   slices.Clone(t.Lines),
			})

			i++
			j++

		case o.Start < t.Start:
			result = append(result, normalMergeHunk(o))
			i++

		default:
			result = append(result, normalMergeHunk(t))
			j++
		}
	}

	return result
}

func ApplyMergeHunks(
	base []string,
	hunks []MergeHunk,
) TextMergeResult {
	result := make([]string, 0)
	conflict := false

	pos := 0

	for _, h := range hunks {

		result = append(result, base[pos:h.Start]...)

		if h.Conflict {
			conflict = true

			result = append(result, "<<<<<<< ours")
			result = append(result, h.Ours...)
			result = append(result, "=======")
			result = append(result, h.Theirs...)
			result = append(result, ">>>>>>> theirs")
		} else {
			result = append(result, h.Lines...)
		}

		pos = h.End
	}

	result = append(result, base[pos:]...)

	return TextMergeResult{
		Lines:    result,
		Conflict: conflict,
	}
}

func Diff(base, target []string) []Hunk {
	lcsTable := lcs(base, target)

	hunks := make([]Hunk, 0)

	i := 0
	j := 0

	for i < len(base) || j < len(target) {

		if i < len(base) &&
			j < len(target) &&
			base[i] == target[j] {
			i++
			j++

			continue
		}

		start := i
		replacement := make([]string, 0)

		for i < len(base) || j < len(target) {

			if i < len(base) &&
				j < len(target) &&
				base[i] == target[j] {
				break
			}

			switch {

			case i < len(base) &&
				j < len(target) &&
				lcsTable[i+1][j] >= lcsTable[i][j+1]:
				i++

			case j < len(target):
				replacement = append(replacement, target[j])
				j++

			default:
				i++
			}
		}

		hunks = append(hunks, Hunk{
			Start: start,
			End:   i,
			Lines: replacement,
		})
	}

	return hunks
}

func lcs(a, b []string) [][]int {
	dp := make([][]int, len(a)+1)

	for i := range dp {
		dp[i] = make([]int, len(b)+1)
	}

	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = 1 + dp[i+1][j+1]
			} else {
				dp[i][j] = max(
					dp[i+1][j],
					dp[i][j+1],
				)
			}
		}
	}

	return dp
}

func normalMergeHunk(h Hunk) MergeHunk {
	return MergeHunk{
		Start: h.Start,
		End:   h.End,
		Lines: slices.Clone(h.Lines),
	}
}

func hunksEqual(a, b Hunk) bool {
	return a.Start == b.Start &&
		a.End == b.End &&
		slices.Equal(a.Lines, b.Lines)
}

func hunksOverlap(a, b Hunk) bool {
	aInsertion := a.Start == a.End
	bInsertion := b.Start == b.End

	if aInsertion && bInsertion {
		return a.Start == b.Start
	}

	if aInsertion {
		return a.Start >= b.Start &&
			a.Start < b.End
	}

	if bInsertion {
		return b.Start >= a.Start &&
			b.Start < a.End
	}

	return a.Start < b.End &&
		b.Start < a.End
}

func splitLines(data []byte) []string {
	if len(data) == 0 {
		return []string{}
	}

	lines := make([]string, 0)

	start := 0

	for i, b := range data {
		if b != '\n' {
			continue
		}

		lines = append(
			lines,
			string(data[start:i+1]),
		)

		start = i + 1
	}

	if start < len(data) {
		lines = append(
			lines,
			string(data[start:]),
		)
	}

	return lines
}

func sameHash(a, b *internal.Hash) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return *a == *b
}
