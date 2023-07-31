package store

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	git "github.com/libgit2/git2go/v34"
)

// A map from each temporary branch (used in RebaseTree) to the branch it
// replaced.
type TempBranchMap map[*git.Branch]*git.Branch

// Read temporary branches file.
//
// It is expected that the file exists.
func ReadTemporaryBranches(repo *git.Repository, filepath string) TempBranchMap {
	readFile, _ := os.Open(filepath)
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	lines := []string{}
	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}

	return string2TempBranchMap(repo, lines)
}

// Write temporary branches file.
func WriteTemporaryBranches(tempMap TempBranchMap, filepath string) {
	OverwriteFile(filepath, tempBranchMap2String(tempMap))
}

func string2TempBranchMap(repo *git.Repository, input []string) TempBranchMap {
	output := TempBranchMap{}
	for _, line := range input {
		parts := strings.Fields(line)
		tempBranch, _ := repo.LookupBranch(parts[0], git.BranchLocal)
		origBranch, _ := repo.LookupBranch(parts[1], git.BranchLocal)

		output[tempBranch] = origBranch
	}
	return output
}

func tempBranchMap2String(tempMap TempBranchMap) string {
	output := []string{}
	for tempBranch, origBranch := range tempMap {
		tempName, _ := tempBranch.Name()
		origName, _ := origBranch.Name()

		output = append(output, fmt.Sprintf("%s %s", tempName, origName))
	}
	return strings.Join(output, "\n")
}
