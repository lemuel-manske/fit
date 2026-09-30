package cmd

import (
	"fmt"

	"fit/fit/internal"
)

const (
	errNotAncestor = "target commit is not a descendant of the current HEAD"
)

type MergeDecision struct {
	Blob     *internal.Hash
	Delete   bool
	Conflict bool
}

func FastForward(dir string, target internal.CommitID) error {
	head, err := internal.ReadHEAD(dir)
	if err != nil {
		return err
	}

	targetAncestor, err := internal.Ancestor(dir, head, target)
	if err != nil {
		return err
	}

	if targetAncestor != head {
		err := fmt.Errorf("%s: %s is not an ancestor of %s", errNotAncestor, head, target)

		return err
	}

	return Checkout(dir, target)
}

func TakeMergeDecision(base, ours, theirs *internal.Hash) MergeDecision {
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

func sameHash(a, b *internal.Hash) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return *a == *b
}
