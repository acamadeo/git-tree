package gitutil

import (
	"fmt"

	git "github.com/libgit2/git2go/v34"
)

func AllLocalBranches(repo *git.Repository) []*git.Branch {
	output := []*git.Branch{}

	iterator, _ := repo.NewBranchIterator(git.BranchLocal)
	iterator.ForEach(func(b *git.Branch, bt git.BranchType) error {
		output = append(output, b)
		return nil
	})
	iterator.Free()

	return output
}

func BranchName(branch *git.Branch) string {
	name, _ := branch.Name()
	return name
}

// Returns true if branch `a` is an ancestor of branch `b`.
func IsBranchAncestor(repo *git.Repository, a *git.Branch, b *git.Branch) bool {
	commitOidA := a.Target()

	revWalk, _ := repo.Walk()
	revWalk.Sorting(git.SortTopological)
	revWalk.Push(b.Target())

	found := false
	revWalk.Iterate(func(commit *git.Commit) bool {
		if *commit.Id() == *commitOidA {
			found = true
			return false
		}
		return true
	})

	return found
}

func UniqueBranchName(repo *git.Repository, name string) string {
	for i := 0; i < 10000; i++ {
		tryName := nameWithNumber(name, i)
		branch, _ := repo.LookupBranch(tryName, git.BranchLocal)
		if branch == nil {
			return tryName
		}
	}
	return ""
}

func nameWithNumber(name string, number int) string {
	if number == 0 {
		return name
	}
	return fmt.Sprintf("%s-%d", name, number)
}

func HeadBranch(repo *git.Repository) *git.Branch {
	headRef, _ := repo.Head()
	return headRef.Branch()
}

func UpdateBranchTarget(branch **git.Branch, target *git.Oid) {
	msg := fmt.Sprintf("[git-tree] update branch target for %s", BranchName(*branch))
	newRef, _ := (*branch).SetTarget(target, msg)
	*branch = newRef.Branch()
}

func CreateBranchAtCommit(repo *git.Repository, commit *git.Commit, name string) *git.Branch {
	branch, _ := repo.CreateBranch(UniqueBranchName(repo, name), commit, false)
	return branch
}

func AnnotatedCommitFromBranch(repo *git.Repository, branch *git.Branch) *git.AnnotatedCommit {
	return AnnotatedCommitForReference(repo, branch.Reference)
}
