package gitutil

import (
	"sort"
	"strings"

	git "github.com/libgit2/git2go/v34"
)

// A list of commits that does not have duplicate entries.
type commitList []git.Oid

func (l commitList) add(oid git.Oid) commitList {
	for _, o := range l {
		if o == oid {
			return l
		}
	}
	l = append(l, oid)
	return l
}

type commitMap map[git.Oid]commitList

type repoTree struct {
	repo *git.Repository
	root git.Oid
	// Map from each commit to its children, where the children are sorted
	// alphabetically by commit message.
	commitChildren commitMap
	// Map each commit to branches that point to them. The branches are sorted
	// alphabetically by branch name.
	branches map[git.Oid][]string
}

func findRepoTreeRoot(repo *git.Repository) git.Oid {
	revWalk := InitWalkWithAllBranches(repo)

	// Iterate through the commits, looking for the root commit (which won't
	// have parents).
	rootOid := git.Oid{}
	revWalk.Iterate(func(commit *git.Commit) bool {
		if commit.ParentCount() == 0 {
			rootOid = *commit.Id()
		}
		return true
	})
	return rootOid
}

func initCommitChildren(allCommits []*git.Commit) commitMap {
	ret := commitMap{}
	for _, commit := range allCommits {
		ret[*commit.Id()] = commitList{}
	}
	return ret
}

// Sorts every commit's children alphabetically by commit message.
//
// Assumes that a commit won't have two child commits with the same message.
func sortCommitMap(repo *git.Repository, commitChildren commitMap) commitMap {
	sortedMap := commitMap{}
	for commit, children := range commitChildren {
		sort.Slice(children, func(i, j int) bool {
			commitA, _ := repo.LookupCommit(&children[i])
			commitB, _ := repo.LookupCommit(&children[j])
			return strings.Compare(commitA.Message(), commitB.Message()) < 0
		})
		sortedMap[commit] = children
	}
	return sortedMap
}

func createCommitChildren(repo *git.Repository) commitMap {
	commitChildren := initCommitChildren(AllLocalCommits(repo))

	// Iterate through the commits, constructing a commit descendancy tree.
	revWalk := InitWalkWithAllBranches(repo)
	revWalk.Iterate(func(commit *git.Commit) bool {
		// Add this commit as a child of its parent.
		if commit.ParentCount() > 0 {
			children := commitChildren[*commit.ParentId(0)]
			children = children.add(*commit.Id())
			commitChildren[*commit.ParentId(0)] = children
		}
		return true
	})

	return sortCommitMap(repo, commitChildren)
}

// Sort the branch names alphabetically.
func sortBranchesByName(branches map[git.Oid][]string) map[git.Oid][]string {
	ret := map[git.Oid][]string{}
	for oid, branchNames := range branches {
		sort.Strings(branchNames)
		ret[oid] = branchNames
	}
	return ret
}

func createBranches(repo *git.Repository) map[git.Oid][]string {
	branches := map[git.Oid][]string{}
	for _, commit := range AllLocalCommits(repo) {
		branches[*commit.Id()] = []string{}
	}

	for _, branch := range AllLocalBranches(repo) {
		names := branches[*branch.Target()]
		names = append(names, BranchName(branch))
		branches[*branch.Target()] = names
	}
	return sortBranchesByName(branches)
}

func createRepoTree(repo *git.Repository) repoTree {
	return repoTree{
		repo:           repo,
		root:           findRepoTreeRoot(repo),
		commitChildren: createCommitChildren(repo),
		branches:       createBranches(repo),
	}
}

func isIdenticalRecurse(nodeA git.Oid, treeA *repoTree, nodeB git.Oid, treeB *repoTree) bool {
	// Check whether the current node is identical.
	commitA, _ := treeA.repo.LookupCommit(&nodeA)
	commitB, _ := treeB.repo.LookupCommit(&nodeB)

	if commitA.Message() != commitB.Message() {
		return false
	}
	if !stringListEqual(treeA.branches[nodeA], treeB.branches[nodeB]) {
		return false
	}

	// Check that each child is identical.
	childrenA := treeA.commitChildren[nodeA]
	childrenB := treeB.commitChildren[nodeB]

	if len(childrenA) != len(childrenB) {
		return false
	}
	for i := range childrenA {
		if !isIdenticalRecurse(childrenA[i], treeA, childrenB[i], treeB) {
			return false
		}
	}

	return true
}

func isIdentical(a *repoTree, b *repoTree) bool {
	return isIdenticalRecurse(a.root, a, b.root, b)
}

// Returns true if both repo's have the same branches and commits.
//
// Equality requires branch names and commit messages to be the same across
// repo's.
//
// *Assumes that a commit won't have two child commits with the same message*.
func TreesEqual(a *git.Repository, b *git.Repository) bool {
	// We must check whether the repositories are isomorphic.
	//
	// Create a commit descendancy tree for each repository. Address isomorphism
	// by sorting each tree:
	//  * Sort the children of each commit alphabetically by commit message.
	//  * Sort the names of the branches associated with a particular commit.
	repoTreeA := createRepoTree(a)
	repoTreeB := createRepoTree(b)

	// Run through both trees, confirming that they are identical.
	return isIdentical(&repoTreeA, &repoTreeB)
}

func stringListEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
