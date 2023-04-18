package models

import (
	"fmt"

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
	// If `branch` does not exist in `branchMap`, add a key for it.
	_, ok := branchMap.Children[branch]
	if !ok {
		branchMap.Children[branch] = BranchList{}
	}

	curChildren := branchMap.Children[branch]
	newChildren := append(curChildren, children...)
	branchMap.Children[branch] = newChildren
}

// Remove `children` from the given branch `branch`.
func removeChildren(branchMap *BranchMap, branch *git.Branch, children []*git.Branch) {
	for _, child := range children {
		removeChild(branchMap, branch, child)
	}
}

// Remove a child from a list of children by value.
func removeChild(branchMap *BranchMap, branch *git.Branch, child *git.Branch) {
	children := branchMap.Children[branch]
	for i, c := range children {
		if c == child {
			children = append(children[:i], children[i+1:]...)
		}
	}
	branchMap.Children[branch] = children

	// If there are no more children left, delete map entry for `branch`.
	if len(children) == 0 {
		delete(branchMap.Children, branch)
	}
}

// Returns true if reference `a` is an ancestor of reference `b`.
func isAncestor(repo *git.Repository, a *git.Branch, b *git.Branch) bool {
	oidA := a.Target()
	oidB := b.Target()

	ancestorOid, _ := repo.MergeBase(oidA, oidB)
	return oidA.Equal(ancestorOid)
}

// Print a BranchMap (for debugging).
func (b *BranchMap) String() string {
	output := ""
	rootName, _ := b.Root.Name()
	output += fmt.Sprintf("Root: %s\n", rootName)

	for parent, children := range b.Children {
		parentName, _ := parent.Name()
		output += fmt.Sprintf(" - %s: ", parentName)

		if len(children) == 0 {
			output += "<empty>\n"
			continue
		}

		for _, child := range children {
			childName, _ := child.Name()
			output += fmt.Sprintf("%s ", childName)
		}
		output += "\n"
	}

	return output
}

func (b *BranchMap) FindBranch(branchName string) *git.Branch {
	for parent := range b.Children {
		name, _ := parent.Name()
		if name == branchName {
			return parent
		}
	}
	return nil
}

// Returns the parent of a branch.
func (b *BranchMap) FindParent(branchName string) *git.Branch {
	for parent, children := range b.Children {
		for _, child := range children {
			childName, _ := child.Name()
			if childName == branchName {
				return parent
			}
		}
	}
	return nil
}

func (b *BranchMap) FindChildren(branchName string) BranchList {
	for parent, children := range b.Children {
		name, _ := parent.Name()
		if name == branchName {
			return children
		}
	}
	return nil
}

func (b *BranchMap) RemoveChildren(parentName string, children []string) {
	parent := b.FindBranch(parentName)

	for _, child := range children {
		childBranch := b.FindBranch(child)
		removeChild(b, parent, childBranch)
	}
}

// Returns true if the branch `ancestor` is an ancestor of branch `descendant`.
//
// NOTE: Returns false if `ancestor` or `descendant` aren't tracked in the
// BranchMap.
func (b *BranchMap) IsBranchAncestor(ancestor string, descendant string) bool {
	ancestorBranch := b.FindBranch(ancestor)
	if ancestorBranch == nil {
		return false
	}

	if b.IsBranchParent(ancestor, descendant) {
		return true
	}

	// Recurse into each child to see if they are the parent of `descendant`.
	ancestorChildren := b.Children[ancestorBranch]
	for _, ancestorChild := range ancestorChildren {
		childName, _ := ancestorChild.Branch().Name()
		if b.IsBranchAncestor(childName, descendant) {
			return true
		}
	}

	return false
}

// Returns true if the branch `ancestor` is the parent of branch `descendant`.
//
// NOTE: Returns false if `ancestor` or `descendant` aren't tracked in the
// BranchMap.
func (b *BranchMap) IsBranchParent(ancestor string, child string) bool {
	ancestorBranch := b.FindBranch(ancestor)
	if ancestorBranch == nil {
		return false
	}

	ancestorChildren := b.Children[ancestorBranch]
	for _, ancestorChild := range ancestorChildren {
		name, _ := ancestorChild.Branch().Name()
		if name == child {
			return true
		}
	}

	return false
}
