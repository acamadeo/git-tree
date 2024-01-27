package gitutil

import git "github.com/libgit2/git2go/v34"

// A list of commits that does not have duplicate entries.
type CommitSet []git.Oid

func (set CommitSet) Add(oid git.Oid) CommitSet {
	for _, o := range set {
		if o == oid {
			return set
		}
	}
	set = append(set, oid)
	return set
}
