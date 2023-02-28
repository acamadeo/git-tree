package store

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/abaresk/git-tree/models"
	git "github.com/libgit2/git2go/v34"
	"golang.org/x/exp/maps"
)

// Read branch map file.
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

func branchMap2String(branchMap *models.BranchMap) string {
	output := []string{}

	// First print the name of the root branch.
	root, _ := branchMap.Root.Name()
	output = append(output, root)

	// Each branch in the tree has its own line. The line starts with the branch
	// name and is followed by a space-delimited list of the names of the branch's
	// children.
	for branch, children := range branchMap.Children {
		branchName, _ := branch.Name()
		entry := fmt.Sprintf("%s ", branchName)

		for _, child := range children {
			childBranchName, _ := child.Name()
			entry += fmt.Sprintf("%s ", childBranchName)
		}
		output = append(output, entry)
	}

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
		Children: childrenMap(input[1:], nameMap),
	}
}

// Return all the branches used in the parent-children string representation.
func extractBranchNames(childrenLines []string) []string {
	nameSet := map[string]struct{}{}
	for _, line := range childrenLines {
		names := strings.Split(line, " ")
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

// Construct a mapping from each branch to its children branches.
//
// Receives the lines in the string representation of the children map. Also
// receives a lookup table from branch name to its *git.Branch.
func childrenMap(childrenLines []string, nameMap map[string]*git.Branch) map[*git.Branch]models.BranchList {
	childrenMap := map[*git.Branch]models.BranchList{}

	for _, line := range childrenLines {
		names := strings.Split(line, " ")

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
