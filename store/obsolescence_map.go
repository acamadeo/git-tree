package store

import (
	"fmt"
	"strings"

	"github.com/acamadeo/git-tree/models"
)

// Read obsolescence map file
func ReadObsolescenceMap(filepath string) *models.ObsolescenceMap {
	// TODO: Implement this!!
	return &models.ObsolescenceMap{}
}

// Write obsolescence map file
func WriteObsolescenceMap(obsmap *models.ObsolescenceMap, filepath string) {
	OverwriteFile(filepath, obsolescenceMapString(obsmap))
}

func obsolescenceMapString(obsmap *models.ObsolescenceMap) string {
	output := []string{}

	for _, entry := range obsmap.Entries {
		entryString := fmt.Sprintf("%s %s %s", entry.Commit, entry.Obsoleter, entry.ObsoleterBranch)
		output = append(output, entryString)
	}

	return strings.Join(output, "\n")
}
