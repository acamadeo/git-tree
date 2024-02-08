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

func LookupBranches(repo *git.Repository, branchNames ...string) []*git.Branch {
	branches := []*git.Branch{}
	for _, name := range branchNames {
		branch, _ := repo.LookupBranch(name, git.BranchLocal)
		branches = append(branches, branch)
	}
	return branches
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

func MergeBaseMany_Branches(repo *git.Repository, branches ...*git.Branch) *git.Oid {
	branchOids := branchOids(branches)
	if len(branchOids) == 1 {
		return branchOids[0]
	}
	root, _ := repo.MergeBaseMany(branchOids)
	return root
}

func MergeBaseOctopus_Branches(repo *git.Repository, branches ...*git.Branch) *git.Oid {
	branchOids := branchOids(branches)
	if len(branchOids) == 1 {
		return branchOids[0]
	}
	root, _ := repo.MergeBaseOctopus(branchOids)
	return root
}

// Update `branch` to point to commit `target`.
func MoveBranchTarget(repo *git.Repository, branch **git.Branch, target *git.Oid) error {
	commit, _ := repo.LookupCommit(target)
	commitTree, _ := commit.Tree()

	// Check out the working tree at the given branch.
	if err := repo.CheckoutTree(commitTree, checkoutOpts()); err != nil {
		return fmt.Errorf("Could not checkout tree: %s", err)
	}

	msg := fmt.Sprintf("[git-tree] update branch target for %s", BranchName(*branch))
	newRef, _ := (*branch).SetTarget(target, msg)
	*branch = newRef.Branch()
	return nil
}

func CreateBranchAtCommit(repo *git.Repository, commit *git.Commit, name string) *git.Branch {
	branch, _ := repo.CreateBranch(UniqueBranchName(repo, name), commit, false)
	return branch
}

func AnnotatedCommitFromBranch(repo *git.Repository, branch *git.Branch) *git.AnnotatedCommit {
	return AnnotatedCommitForReference(repo, branch.Reference)
}

func branchOids(branches []*git.Branch) []*git.Oid {
	oidSet := map[git.Oid]*git.Oid{}
	for _, branch := range branches {
		oidSet[*branch.Target()] = branch.Target()
	}

	unique := []*git.Oid{}
	for _, ptr := range oidSet {
		unique = append(unique, ptr)
	}
	return unique
}
