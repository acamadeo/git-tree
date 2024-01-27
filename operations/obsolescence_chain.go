package operations

import (
	"fmt"
	"log"

	gitutil "github.com/acamadeo/git-tree/git"
	"github.com/acamadeo/git-tree/models"
	git "github.com/libgit2/git2go/v34"
)

// TODO: Consider moving into a dedicated `evolve/` directory!

// Represents the commits that obsoleted another set of commits in a Git action.
type obsolescenceChain struct {
	// Commits on the obsoleted side of the chain (from oldest to newest).
	obsoleted []*git.Commit
	// Commits on the obsoleter side of the chain (from oldest to newest).
	obsoleter []*git.Commit
}

func (oc *obsolescenceChain) HasObsoleteCommit(commit *git.Commit) bool {
	for _, c := range oc.obsoleted {
		if *c.Id() == *commit.Id() {
			return true
		}
	}
	return false
}

type obsolescenceChains []obsolescenceChain

// Find the obsolescence chain where the given commit got obsoleted.
//
// Returns nil if the given commit was not obsoleted.
func (ocs *obsolescenceChains) FindChainWithObsoleteCommit(commit *git.Commit) *obsolescenceChain {
	for _, chain := range *ocs {
		if chain.HasObsoleteCommit(commit) {
			return &chain
		}
	}
	return nil
}

func buildObsolescenceChains(repo *git.Repository, obsmap *models.ObsolescenceMap) obsolescenceChains {
	chains := []obsolescenceChain{}
	for _, action := range obsmap.Actions {
		chains = append(chains, buildObsolescenceChain(repo, action))
	}
	return chains
}

// Maps a commit to a list of the commit's children.
type commitTree struct {
	root git.Oid
	tree map[git.Oid]gitutil.CommitSet
}

// Steps to build obsolescence chain:
//  1. Filter out any commits that are no longer referenced in the repo.
//  2. Find the merge-base of all the commits that are listed. This is the root
//     commit.
//  3. All the commits should be in one of 2 chains extending from the root
//     commit. Make a tree (map[Oid][]Oid). When you add a commit, go through
//     and add all its parents until you reach the root.
//  4. Assert that the root node only has 2 children. Each subsequent child
//     should have 1 child.
//  5. Determine which chain is obsoleter/obsoleted. Look through each chain (not
//     the root), and see if it contains any obsolete commits that were obsoleted
//     by a commit in the other chain. If so, this chain is obsoleted; the other
//     is obsoleter.
//  6. Return the two chains. The root should not appear in either chain.
func buildObsolescenceChain(repo *git.Repository, action models.ObsolescenceAction) obsolescenceChain {
	commits := []*git.Commit{}
	for _, entry := range action.Entries {
		commits = append(commits, entry.Commit)
		commits = append(commits, entry.Obsoleter)
	}

	commits = filterStaleCommits(repo, uniqueCommits(repo, commits))

	root := gitutil.MergeBaseOctopus_Commits(repo, commits...)

	commitTree := createCommitTree(repo, root, commits)
	fmt.Println("Commit tree:")
	fmt.Println(commitTree)
	validateCommitTree(commitTree)

	leftChain, rightChain := splitTreeToChains(repo, commitTree)
	obsoleted, obsoleter := pickObsoletedObsoleter(action, leftChain, rightChain)

	return obsolescenceChain{obsoleted: obsoleted, obsoleter: obsoleter}
}

func createCommitTree(repo *git.Repository, root *git.Oid, commits []*git.Commit) commitTree {
	tree := map[git.Oid]gitutil.CommitSet{
		*root: {},
	}

	// Add every commit and its ancestors to the tree.
	for _, commit := range commits {
		for !commit.Id().Equal(root) {
			parent := commit.Parent(0)
			tree[*parent.Id()] = tree[*parent.Id()].Add(*commit.Id())
			commit = parent
		}
	}
	return commitTree{root: *root, tree: tree}
}

// The root node should have two children. All other commits should have 1 child.
func validateCommitTree(commitTree commitTree) {
	if len(commitTree.tree[commitTree.root]) != 2 {
		log.Fatalf("Invalid commitTree: root has %d children\n", len(commitTree.tree[commitTree.root]))
	}
	for oid, children := range commitTree.tree {
		if !oid.Equal(&commitTree.root) && len(children) != 1 {
			log.Fatalf("Invalid commitTree: %s has %d children\n", gitutil.OidShortHash(oid), len(children))
		}
	}
}

// Split the commit tree into the two chains extending from the root commit.
func splitTreeToChains(repo *git.Repository, commitTree commitTree) ([]*git.Commit, []*git.Commit) {
	leftChain, rightChain := []*git.Commit{}, []*git.Commit{}

	oid := commitTree.tree[commitTree.root][0]
	for {
		commit, _ := repo.LookupCommit(&oid)
		leftChain = append(leftChain, commit)
		if len(commitTree.tree[oid]) == 0 {
			break
		}
		oid = commitTree.tree[oid][0]
	}

	oid = commitTree.tree[commitTree.root][1]
	for {
		commit, _ := repo.LookupCommit(&oid)
		rightChain = append(rightChain, commit)
		if len(commitTree.tree[oid]) == 0 {
			break
		}
		oid = commitTree.tree[oid][0]
	}

	return leftChain, rightChain
}

// Returns (<obsoleted-chain>, <obsoleter-chain>).
func pickObsoletedObsoleter(action models.ObsolescenceAction, leftChain, rightChain []*git.Commit) ([]*git.Commit, []*git.Commit) {
	leftOids, rightOids := map[git.Oid]bool{}, map[git.Oid]bool{}
	for _, commit := range leftChain {
		leftOids[*commit.Id()] = true
	}
	for _, commit := range rightChain {
		rightOids[*commit.Id()] = true
	}

	for _, entry := range action.Entries {
		_, obsoletedLeft := leftOids[*entry.Commit.Id()]
		_, obsoletedRight := rightOids[*entry.Commit.Id()]
		_, obsoleterLeft := leftOids[*entry.Obsoleter.Id()]
		_, obsoleterRight := rightOids[*entry.Obsoleter.Id()]
		if obsoletedLeft && obsoleterRight {
			return leftChain, rightChain
		}
		if obsoletedRight && obsoleterLeft {
			return rightChain, leftChain
		}
	}

	log.Fatal("Neither `leftChain` nor `rightChain` had an obsoleted -> obsoleter relationship")
	return leftChain, rightChain
}

// Return a new list of commits that are still in the repo.
func filterStaleCommits(repo *git.Repository, commits []*git.Commit) []*git.Commit {
	ret := []*git.Commit{}
	for _, commit := range commits {
		if c, _ := repo.LookupCommit(commit.Id()); c != nil {
			ret = append(ret, commit)
		}
	}
	return ret
}

func uniqueCommits(repo *git.Repository, commits []*git.Commit) []*git.Commit {
	oidSet := map[git.Oid]bool{}
	for _, c := range commits {
		oidSet[*c.Id()] = true
	}

	unique := []*git.Commit{}
	for oid := range oidSet {
		unique = append(unique, gitutil.CommitByOid(repo, oid))
	}
	return unique
}

func (oc *obsolescenceChain) String() string {
	output := "┌"
	for _, commit := range oc.obsoleted {
		output += " " + gitutil.CommitShortHash(commit)
	}
	output += "\n↓\n└"
	for _, commit := range oc.obsoleter {
		output += " " + gitutil.CommitShortHash(commit)
	}
	return output
}

func (commitTree commitTree) String() string {
	return commitTree.stringRecurse(commitTree.root, 0)
}

func (commitTree commitTree) stringRecurse(node git.Oid, depth int) string {
	ret := ""
	for i := 0; i < depth; i++ {
		ret += "\t"
	}
	ret += gitutil.OidShortHash(node) + "\n"

	for _, child := range commitTree.tree[node] {
		ret += commitTree.stringRecurse(child, depth+1)
	}
	return ret
}
