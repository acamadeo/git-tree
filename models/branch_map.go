package models

import (
	git "github.com/libgit2/git2go/v34"
)

type BranchList []*git.Branch

// Represents the parent-children relationships between branches.
type BranchMap struct {
	Root     *git.Branch
	Children map[*git.Branch]BranchList
}

// QUESTION LATER: What happens if the user skipped generations?
//  - For example, they provided branch <A> and its grandparent branch <C>, but
//    they omitted branch <A>'s parent, branch <B>.
//  - Initially, I think that maybe you should allow the user to call
//    `git tree init -b <B>` later to add any branches into the tree.
//  - What would be really nice is if you don't need to pass any branches into
//    `git tree init`. E.g., if it constructed a tree using all the local
//    branches. But I haven't thought about feasibility / design implications.

// Constructs a descendency tree describing the parent-child relationships
// between the given branches.
//
// The tree takes the form of a map, where the key is a parent branch and the
// value is the list of its children.
func BranchMapFromRepo(repo *git.Repository, root *git.Branch, branches []*git.Branch) *BranchMap {
	branchMap := new(BranchMap)
	branchMap.Children = map[*git.Branch]BranchList{}

	branchMap.Root = root
	branchMap.Children[root] = BranchList{}

	// Go through the provided branches and add them to the branchMap.
	for _, branch := range branches {
		addBranchToTree(repo, branchMap, branchMap.Root, branch)
	}

	return branchMap
}

// Add a new branch `newBranch` to the descendency tree.
//
// This is a recursive function that is initially passed in the root of the
// descendency tree.
func addBranchToTree(repo *git.Repository, branchMap *BranchMap, curBranch *git.Branch, newBranch *git.Branch) {
	// Get the children of the current branch.
	curChildren := branchMap.Children[curBranch]

	// Check if any children are ancestors of the new branch. If so, recurse
	// into that child.
	for _, child := range curChildren {
		if isAncestor(repo, child, newBranch) {
			addBranchToTree(repo, branchMap, child, newBranch)
			return
		}
	}

	// Check if any of the current branch's children are descendents of the new
	// branch.
	descendents := []*git.Branch{}
	for _, child := range curChildren {
		if isAncestor(repo, newBranch, child) {
			descendents = append(descendents, child)
		}
	}

	// If none of the current branch's children are descendents of the new
	// branch, add the new branch as a child of the current branch.
	if len(descendents) == 0 {
		addChildren(branchMap, curBranch, []*git.Branch{newBranch})
		return
	}

	// Some children of the current branch are descendents of the new branch.
	// Move the descendents under the new branch and add the new branch as a
	// child of the current branch.
	removeChildren(branchMap, curBranch, descendents)
	addChildren(branchMap, newBranch, descendents)
	addChildren(branchMap, curBranch, []*git.Branch{newBranch})
}

// Add `children` to the given branch `branch`.
func addChildren(branchMap *BranchMap, branch *git.Branch, children []*git.Branch) {
	curChildren := branchMap.Children[branch]
	newChildren := append(curChildren, children...)
	branchMap.Children[branch] = newChildren
}

// Remove `children` from the given branch `branch`.
func removeChildren(branchMap *BranchMap, branch *git.Branch, children []*git.Branch) {
	curChildren := branchMap.Children[branch]
	for _, child := range children {
		removeChild(curChildren, child)
	}
}

// Remove a child from a list of children by value.
func removeChild(children []*git.Branch, child *git.Branch) []*git.Branch {
	for i, c := range children {
		if c == child {
			return append(children[:i], children[i+1:]...)
		}
	}
	return children
}

// Returns true if reference `a` is an ancestor of reference `b`.
func isAncestor(repo *git.Repository, a *git.Branch, b *git.Branch) bool {
	oidA := a.Target()
	oidB := b.Target()

	ancestorOid, _ := repo.MergeBase(oidA, oidB)
	return oidA.Equal(ancestorOid)
}
