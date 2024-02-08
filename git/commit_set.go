package gitutil

import git "github.com/libgit2/git2go/v34"

// A list of commits that does not have duplicate entries.
type CommitSet []git.Oid

func NewCommitSet(oids ...git.Oid) CommitSet {
	oidSet := map[git.Oid]bool{}
	for _, oid := range oids {
		oidSet[oid] = true
	}

	set := CommitSet{}
	for oid := range oidSet {
		set = append(set, oid)
	}
	return set
}

func (set CommitSet) Add(oid git.Oid) CommitSet {
	existingOids := map[git.Oid]bool{}
	for _, o := range set {
		existingOids[o] = true
	}

	if _, ok := existingOids[oid]; ok {
		return set
	}
	return append(set, oid)
}

func (set CommitSet) AddAll(oids ...git.Oid) CommitSet {
	new := NewCommitSet(oids...)
	for _, o := range new {
		set = set.Add(o)
	}
	return set
}

func (set CommitSet) Remove(oid git.Oid) CommitSet {
	for i, o := range set {
		if o == oid {
			return append(set[:i], set[i+1:]...)
		}
	}
	return set
}
