package store

import (
	"bufio"
	"os"
	"strings"

	"github.com/abaresk/git-tree/models"
	git "github.com/libgit2/git2go/v34"
	"golang.org/x/exp/maps"
)

// Read branch map file.
//
// It is expected that the file exists.
func ReadBranchMap(repo *git.Repository, filepath string) *models.BranchMap {
	readFile, _ := os.Open(filepath)
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	lines := []string{}
	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}

	return string2BranchMap(repo, lines)
}

// Write branch map file.
func WriteBranchMap(branchMap *models.BranchMap, filepath string) {
	overwriteFile(filepath, branchMap2String(branchMap))
}

/**
 * Converts branch to and from a string representation for storage.
 */

func branchAndChildrenString(branchMap *models.BranchMap, branch string) string {
	names := []string{branch}

	children := branchMap.FindChildren(branch)
	for _, child := range children {
		childBranchName, _ := child.Name()
		names = append(names, childBranchName)
	}
	return strings.Join(names, " ")
}

func branchMap2StringRecurse(branchMap *models.BranchMap, branch string) []string {
	// Skip branches without any children.
	children := branchMap.FindChildren(branch)
	if len(children) == 0 {
		return []string{}
	}

	output := []string{}

	// Add the current branch's children.
	branchAndChildren := branchAndChildrenString(branchMap, branch)
	output = append(output, branchAndChildren)

	// Add children in DFS order.
	for _, child := range children {
		childName, _ := child.Name()
		output = append(output, branchMap2StringRecurse(branchMap, childName)...)
	}
	return output
}

func branchMap2String(branchMap *models.BranchMap) string {
	// First print the name of the root branch.
	rootName, _ := branchMap.Root.Name()
	output := []string{rootName}

	output = append(output, branchMap2StringRecurse(branchMap, rootName)...)

	return strings.Join(output, "\n")
}

func string2BranchMap(repo *git.Repository, input []string) *models.BranchMap {
	// First line is the name of the root branch.
	root, _ := repo.LookupBranch(input[0], git.BranchLocal)

	// Create a lookup table from branch name to its *git.Branch.
	branchNames := extractBranchNames(input[1:])
	nameMap := namesToBranches(repo, branchNames)

	return &models.BranchMap{
		Root:     root,
		Children: populateChildrenMap(input[1:], nameMap),
	}
}

// Return all the branches used in the parent-children string representation.
func extractBranchNames(childrenLines []string) []string {
	nameSet := map[string]struct{}{}
	for _, line := range childrenLines {
		names := strings.Split(strings.TrimSpace(line), " ")
		for _, name := range names {
			nameSet[name] = struct{}{}
		}
	}
	return maps.Keys(nameSet)
}

// NOTE: This should probably return an error because it's easy for the branches
// to get clobbered (e.g. user manually updates a branch name).
func namesToBranches(repo *git.Repository, names []string) map[string]*git.Branch {
	branches := map[string]*git.Branch{}
	for _, name := range names {
		branch, _ := repo.LookupBranch(name, git.BranchLocal)
		branches[name] = branch
	}
	return branches
}

// Populate a mapping from each branch to its children branches.
//
// Receives the lines in the string representation of the children map. Also
// receives a lookup table from branch name to its *git.Branch.
func populateChildrenMap(childrenLines []string, nameMap map[string]*git.Branch) map[*git.Branch]models.BranchList {
	childrenMap := initChildrenMap(nameMap)

	for _, line := range childrenLines {
		names := strings.Split(strings.TrimSpace(line), " ")

		// First name is the parent branch. Following names are its children.
		parent := names[0]
		children := names[1:]

		parentBranch := nameMap[parent]
		childrenMap[parentBranch] = models.BranchList{}

		// Add each child branch under its parent.
		for _, child := range children {
			childBranch := nameMap[child]
			childrenMap[parentBranch] = append(childrenMap[parentBranch], childBranch)
		}
	}

	return childrenMap
}

// Initialize each branch as a key of the children map. The initial value for
// each branch is an empty list of child branches.
func initChildrenMap(nameMap map[string]*git.Branch) map[*git.Branch]models.BranchList {
	childrenMap := map[*git.Branch]models.BranchList{}

	for _, branch := range nameMap {
		childrenMap[branch] = models.BranchList{}
	}
	return childrenMap
}
