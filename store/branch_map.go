package store

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/abaresk/git-tree/models"
)

// Storage representation of a BranchMap.
//
// Mirrors the structure of each line in the branch_map text file.
type BranchMap struct {
	Entries []BranchMapEntry
}

type BranchMapEntry struct {
	// Name of parent branch.
	Parent string
	// Name of a child branch.
	Child string
}

// Read obsolescence map file
func ReadBranchMap(filepath string) *models.BranchMap {
	readFile, _ := os.Open(filepath)
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	lines := []string{}
	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}

	readFile.Close()

	branchMap := string2BranchMap(lines)
	modelBranchMap := toBranchMapModel(&branchMap)
	return &modelBranchMap
}

// Write obsolescence map file
func WriteBranchMap(branchMap *models.BranchMap, filepath string) {
	storeBranchMap := fromBranchMapModel(branchMap)
	overwriteFile(filepath, branchMap2String(&storeBranchMap))
}

/**
 * Translates to and from the model data structure.
 */

func fromBranchMapModel(branchMap *models.BranchMap) BranchMap {
	output := BranchMap{}

	for _, entry := range branchMap.Entries {
		for _, child := range entry.Children {
			outputEntry := BranchMapEntry{Parent: entry.Parent, Child: child}
			output.Entries = append(output.Entries, outputEntry)
		}
	}

	return output
}

func toBranchMapModel(branchMap *BranchMap) models.BranchMap {
	childMap := map[string][]string{}

	for _, entry := range branchMap.Entries {
		children := childMap[entry.Parent]
		children = append(children, entry.Child)
	}

	output := models.BranchMap{}
	for parent, children := range childMap {
		entry := models.BranchMapEntry{Parent: parent, Children: children}
		output.Entries = append(output.Entries, entry)
	}

	return output
}

/**
 * Converts between storage data structure and string representation.
 */

func branchMap2String(branchMap *BranchMap) string {
	output := []string{}

	for _, entry := range branchMap.Entries {
		entryString := fmt.Sprintf("%s %s", entry.Parent, entry.Child)
		output = append(output, entryString)
	}

	return strings.Join(output, "\n")
}

func string2BranchMap(input []string) BranchMap {
	output := BranchMap{}

	for _, line := range input {
		parts := strings.Split(line, " ")
		entry := BranchMapEntry{Parent: parts[0], Child: parts[1]}
		output.Entries = append(output.Entries, entry)
	}

	return output
}
