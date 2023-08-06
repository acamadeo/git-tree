package store

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	gitutil "github.com/abaresk/git-tree/git"
	"github.com/abaresk/git-tree/models"
	git "github.com/libgit2/git2go/v34"
)

// Read temporary branches file.
//
// It is expected that the file exists.
func ReadTemporaryBranches(repo *git.Repository, filepath string) models.TempBranchMap {
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
func WriteTemporaryBranches(tempMap models.TempBranchMap, filepath string) {
	OverwriteFile(filepath, tempBranchMap2String(tempMap))
}

func string2TempBranchMap(repo *git.Repository, input []string) models.TempBranchMap {
	output := models.TempBranchMap{}
	for _, line := range input {
		parts := strings.Fields(line)
		tempBranch, _ := repo.LookupBranch(parts[0], git.BranchLocal)
		origBranch, _ := repo.LookupBranch(parts[1], git.BranchLocal)

		output[tempBranch] = origBranch
	}
	return output
}

func sortedTempBranches(tempMap models.TempBranchMap) []*git.Branch {
	keys := []*git.Branch{}
	for tempBranch := range tempMap {
		keys = append(keys, tempBranch)
	}
	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(gitutil.BranchName(keys[i]), gitutil.BranchName(keys[j])) < 0
	})
	return keys
}

// Temporary branches are listed alphabetically for consistency.
func tempBranchMap2String(tempMap models.TempBranchMap) string {
	sortedTempBranches := sortedTempBranches(tempMap)

	output := []string{}
	for _, tempBranch := range sortedTempBranches {
		origBranch := tempMap[tempBranch]

		tempName := gitutil.BranchName(tempBranch)
		origName := gitutil.BranchName(origBranch)
		output = append(output, fmt.Sprintf("%s %s", tempName, origName))
	}
	return strings.Join(output, "\n")
}
