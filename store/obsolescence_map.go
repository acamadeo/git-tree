package store

import (
	"fmt"
	"strings"

	"github.com/acamadeo/git-tree/models"
)

var hookTypeStrings = map[models.HookType]string{
	models.HookTypeUnknown:   "unknown",
	models.PostRewriteAmend:  "post-rewrite.amend",
	models.PostRewriteRebase: "post-rewrite.rebase",
	models.PostCommit:        "post-commit",
}

// Read obsolescence map file
func ReadObsolescenceMap(filepath string) *models.ObsolescenceMap {
	obsmap := models.ObsolescenceMap{}

	lines := strings.Split(ReadFile(filepath), "\n")
	for _, line := range lines {
		lineParts := strings.Fields(line)
		oldHash := lineParts[0]
		newHash := lineParts[1]
		hookType := lineParts[2]

		obsmap.Entries = append(obsmap.Entries, models.ObsolescenceMapEntry{
			Commit:    oldHash,
			Obsoleter: newHash,
			HookType:  hookTypeFromString(hookType),
		})
	}

	return &obsmap
}

// Write obsolescence map file
func WriteObsolescenceMap(obsmap *models.ObsolescenceMap, filepath string) {
	OverwriteFile(filepath, obsolescenceMapString(obsmap))
}

// Append entries to obsolescence map file
func AppendToObsolescenceMap(filepath string, entries ...models.ObsolescenceMapEntry) {
	obsmap := &models.ObsolescenceMap{}

	if FileExists(filepath) {
		obsmap = ReadObsolescenceMap(filepath)
	}

	obsmap.Entries = append(obsmap.Entries, entries...)
	WriteObsolescenceMap(obsmap, filepath)
}

func obsolescenceMapString(obsmap *models.ObsolescenceMap) string {
	output := []string{}

	for _, entry := range obsmap.Entries {
		entryString := fmt.Sprintf("%s %s %s", entry.Commit, entry.Obsoleter, hookTypeStrings[entry.HookType])
		output = append(output, entryString)
	}

	return strings.Join(output, "\n")
}

func hookTypeFromString(value string) models.HookType {
	for hookType, str := range hookTypeStrings {
		if str == value {
			return hookType
		}
	}
	return models.HookTypeUnknown
}
