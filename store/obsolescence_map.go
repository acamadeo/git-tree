package store

import (
	"fmt"
	"strings"

	"github.com/acamadeo/git-tree/models"
	git "github.com/libgit2/git2go/v34"
)

var eventTypeStrings = map[models.EventType]string{
	models.EventTypeUnknown: "event-type-unknown",
	models.EventTypeRebase:  "event-type-rebase",
	models.EventTypeAmend:   "event-type-amend",
	models.EventTypeCommit:  "event-type-commit",
}

var hookTypeStrings = map[models.HookType]string{
	models.HookTypeUnknown:   "unknown",
	models.PostRewriteAmend:  "post-rewrite.amend",
	models.PostRewriteRebase: "post-rewrite.rebase",
	models.PostCommit:        "post-commit",
}

// Read obsolescence map file
func ReadObsolescenceMap(repo *git.Repository, filepath string) *models.ObsolescenceMap {
	obsmap := models.ObsolescenceMap{}

	lines := strings.Split(ReadFile(filepath), "\n")
	for _, line := range lines {
		// Check for the start of a new event.
		if strings.HasPrefix(line, "event") {
			lineParts := strings.Fields(line)
			obsmap.Events = append(obsmap.Events, models.ObsolescenceEvent{
				EventType: eventTypeFromString(lineParts[1]),
			})
			continue
		}

		// Line does not indicate the start of a new event. Append an entry to
		// the latest event.
		lastEvent := obsmap.Events[len(obsmap.Events)-1]
		lastEvent.Entries = append(lastEvent.Entries,
			obsolescenceEntryFromLine(repo, line))
	}

	return &obsmap
}

func obsolescenceEntryFromLine(repo *git.Repository, line string) models.ObsolescenceEntry {
	lineParts := strings.Fields(line)
	oldHash := lineParts[0]
	newHash := lineParts[1]
	hookType := lineParts[2]

	commitOid, _ := git.NewOid(oldHash)
	commit, _ := repo.LookupCommit(commitOid)

	obsoleterOid, _ := git.NewOid(newHash)
	obsoleter, _ := repo.LookupCommit(obsoleterOid)

	return models.ObsolescenceEntry{
		Commit:    commit,
		Obsoleter: obsoleter,
		HookType:  hookTypeFromString(hookType),
	}
}

func eventTypeFromString(value string) models.EventType {
	for eventType, str := range eventTypeStrings {
		if str == value {
			return eventType
		}
	}
	return models.EventTypeUnknown
}

func hookTypeFromString(value string) models.HookType {
	for hookType, str := range hookTypeStrings {
		if str == value {
			return hookType
		}
	}
	return models.HookTypeUnknown
}

// Write obsolescence map file
func WriteObsolescenceMap(obsmap *models.ObsolescenceMap, filepath string) {
	OverwriteFile(filepath, obsolescenceMapString(obsmap))
}

// Append entries to obsolescence map file under the last ObsolescenceEvent.
//
// TODO: Consider whether if it would be better to append an entire Event at
// once.
func AppendEntriesToLastObsolescenceEvent(repo *git.Repository, filepath string, entries ...models.ObsolescenceEntry) {
	obsmap := &models.ObsolescenceMap{}

	if FileExists(filepath) {
		obsmap = ReadObsolescenceMap(repo, filepath)
	}

	lastEvent := obsmap.Events[len(obsmap.Events)-1]
	lastEvent.Entries = append(lastEvent.Entries, entries...)
	WriteObsolescenceMap(obsmap, filepath)
}

func obsolescenceMapString(obsmap *models.ObsolescenceMap) string {
	output := []string{}

	for _, event := range obsmap.Events {
		eventHeader := fmt.Sprintf("event %s",
			eventTypeStrings[event.EventType])
		output = append(output, eventHeader)

		for _, entry := range event.Entries {
			entryString := fmt.Sprintf("%s %s %s",
				entry.Commit.Id().String(),
				entry.Obsoleter.Id().String(),
				hookTypeStrings[entry.HookType])
			output = append(output, entryString)
		}
	}

	return strings.Join(output, "\n")
}
