package gitutil

import (
	"sort"

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

// An in-memory representation of the local branches and commits.
//
// Git does not provide an easy way to find all the children of a particular
// commit. Furthermore, traversing the commit graph using libgit2 can be an
// expensive operation.
//
// This data structure provides a simple representation of the commit
// descendancy graph. It is best to populate it early and use it for the
// remainder of the operation.
//
// The tree is sorted during construction so that it is easy to compare whether
// two RepoTree's are identical. They are sorted as follows:
//   - Sort the children of each commit by Oid value.
//   - Sort the names of the branches associated with a particular commit.
//
// TODO: Add support for merge commits.
//   - NOTE: libgit2 rebase might not work if the commit history has merge commits
//     https://github.com/libgit2/libgit2/blob/198a1b209a929389c739a8a6abef13e717fdfda9/src/libgit2/rebase.c#L818
//   - NOTE: We could ask users to avoid using `git merge` to sync with the
//     upstream repo. Our `git tree sync` should *rebase* (not merge) working
//     changes onto main HEAD.
//   - - Down the line, we should get all the operations working even if there are
//     merge commits.
type RepoTree struct {
	Repo *git.Repository
	Root git.Oid
	// Map from each commit to its children.
	//
	// The list of children are sorted in ascending order by the Oid value.
	CommitChildren map[git.Oid]commitList
	// Map each commit to branches that point to them. The branches are sorted
	// alphabetically by branch name.
	branches map[git.Oid][]string
}

// Returns a list of Oid's for each child of the given `commit`.
func (r *RepoTree) FindChildren(commit git.Oid) []git.Oid {
	children := r.CommitChildren[commit]

	ret := make([]git.Oid, len(children))
	copy(ret, children)
	return ret
}

// Returns a list of branches that point to the given `commit`.
func (r *RepoTree) FindBranches(commit git.Oid) []string {
	branches := r.branches[commit]

	ret := make([]string, len(branches))
	copy(ret, branches)
	return ret
}

// Returns true if commit `a` is an ancestor of commit `b`.
//
// TODO: Consider moving this under `gitutil` if it's not used elsewhere in this
// file!
func (r *RepoTree) IsAncestor(a *git.Commit, b *git.Commit) bool {
	ancestorOid, _ := r.Repo.MergeBase(a.Id(), b.Id())
	return a.Id().Equal(ancestorOid)
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

func initCommitChildren(allCommits []*git.Commit) map[git.Oid]commitList {
	ret := map[git.Oid]commitList{}
	for _, commit := range allCommits {
		ret[*commit.Id()] = commitList{}
	}
	return ret
}

// Sorts every commit's children ascending order by the Oid value.
func sortCommitMap(repo *git.Repository, commitChildren map[git.Oid]commitList) map[git.Oid]commitList {
	sortedMap := map[git.Oid]commitList{}
	for commit, children := range commitChildren {
		sort.Slice(children, func(i, j int) bool {
			commitA, _ := repo.LookupCommit(&children[i])
			commitB, _ := repo.LookupCommit(&children[j])
			return compareOids(*commitA.Id(), *commitB.Id()) < 0
		})
		sortedMap[commit] = children
	}
	return sortedMap
}

func createCommitChildren(repo *git.Repository, root *git.Branch, branches ...*git.Branch) map[git.Oid]commitList {
	commitChildren := initCommitChildren(LocalCommitsFromBranches(repo, root, branches...))

	// Iterate through the commits, constructing a commit descendancy tree.
	revWalk := InitWalkWithAllBranches(repo)
	revWalk.Iterate(func(commit *git.Commit) bool {
		// Stop adding commits to RepoTree once we hit `root` (if specified).
		if root != nil && *root.Target() == *commit.Id() {
			return false
		}

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

func createBranches(repo *git.Repository, root *git.Branch, branches ...*git.Branch) map[git.Oid][]string {
	commitToBranches := map[git.Oid][]string{}
	for _, commit := range LocalCommitsFromBranches(repo, root, branches...) {
		commitToBranches[*commit.Id()] = []string{}
	}

	// For each local commit, add the names of any branches pointing to that
	// commit.
	for _, branch := range branches {
		names := commitToBranches[*branch.Target()]
		if names != nil {
			names = append(names, BranchName(branch))
			commitToBranches[*branch.Target()] = names
		}
	}
	return sortBranchesByName(commitToBranches)
}

// Creates a `RepoTree` representation of the given repository.
//
// If `root` is nil, the entire commit history is loaded in-memory. If `root` is
// specified, the RepoTree will start at the `root` branch.
//
// The RepoTree will include ancestor commits of the provided `branches`. If no
// `branches` are provided, the RepoTree will include commits from *all local branches*.
// If `branches` are provided, it is assumed that `root` is nil or an ancestor of
// all `branches`.
func CreateRepoTree(repo *git.Repository, root *git.Branch, branches ...*git.Branch) *RepoTree {
	var rootOid git.Oid
	if root != nil {
		rootOid = *root.Target()
	} else {
		rootOid = findRepoTreeRoot(repo)
	}

	if len(branches) == 0 {
		branches = AllLocalBranches(repo)
	}

	return &RepoTree{
		Repo:           repo,
		Root:           rootOid,
		CommitChildren: createCommitChildren(repo, root, branches...),
		branches:       createBranches(repo, root, branches...),
	}
}

func isIdenticalRecurse(nodeA git.Oid, treeA *RepoTree, nodeB git.Oid, treeB *RepoTree) bool {
	// Check whether the current node is identical.
	commitA, _ := treeA.Repo.LookupCommit(&nodeA)
	commitB, _ := treeB.Repo.LookupCommit(&nodeB)

	if commitA.Message() != commitB.Message() {
		return false
	}
	if !stringListEqual(treeA.branches[nodeA], treeB.branches[nodeB]) {
		return false
	}

	// Check that each child is identical.
	childrenA := treeA.CommitChildren[nodeA]
	childrenB := treeB.CommitChildren[nodeB]

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

func isIdentical(a *RepoTree, b *RepoTree) bool {
	return isIdenticalRecurse(a.Root, a, b.Root, b)
}

// Returns true if both repo's have the same branches and commits.
//
// Equality requires branch names and commit messages to be the same across
// repo's.
func TreesEqual(a *RepoTree, b *RepoTree) bool {
	// Run through both trees, confirming that they are identical.
	return isIdentical(a, b)
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
