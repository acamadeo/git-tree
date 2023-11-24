package store

import (
	"errors"
	"fmt"
	"strings"

	"github.com/acamadeo/git-tree/models"
	git "github.com/libgit2/git2go/v34"
)

var eventTypeStrings = map[models.EventType]string{
	models.EventTypeUnknown: "unknown",
	models.EventTypeRebase:  "rebase",
	models.EventTypeAmend:   "amend",
	models.EventTypeCommit:  "commit",
}

var hookTypeStrings = map[models.HookType]string{
	models.HookTypeUnknown:   "unknown",
	models.PostRewriteAmend:  "post-rewrite.amend",
	models.PostRewriteRebase: "post-rewrite.rebase",
	models.PostCommit:        "post-commit",
}

// Read obsolescence map file
func ReadObsolescenceMap(repo *git.Repository, filepath string) *models.ObsolescenceMap {
	contents := ReadFile(filepath)
	if contents == "" {
		return &models.ObsolescenceMap{}
	}

	obsmap := models.ObsolescenceMap{}
	for _, line := range strings.Split(contents, "\n") {
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
		obsmap.Events[len(obsmap.Events)-1] = lastEvent
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

func LastObsolescenceEventType(repo *git.Repository, filepath string) models.EventType {
	obsmap := ReadObsolescenceMap(repo, filepath)
	return obsmap.Events[len(obsmap.Events)-1].EventType
}

func SetLastObsolescenceEventType(repo *git.Repository, filepath string, eventType models.EventType) {
	obsmap := ReadObsolescenceMap(repo, filepath)
	obsmap.Events[len(obsmap.Events)-1].EventType = eventType
	WriteObsolescenceMap(obsmap, filepath)
}

func AppendObsolescenceEvent(repo *git.Repository, filepath string, eventType models.EventType) {
	obsmap := ReadObsolescenceMap(repo, filepath)

	obsmap.Events = append(obsmap.Events, models.ObsolescenceEvent{
		EventType: eventType,
	})
	WriteObsolescenceMap(obsmap, filepath)
}

// Append entries to obsolescence map file under the last ObsolescenceEvent.
func AppendEntriesToLastObsolescenceEvent(repo *git.Repository, filepath string, entries ...models.ObsolescenceEntry) error {
	obsmap := ReadObsolescenceMap(repo, filepath)
	if len(obsmap.Events) < 1 {
		return errors.New("cannot append entry to obsolete map without events")
	}

	lastEvent := obsmap.Events[len(obsmap.Events)-1]
	lastEvent.Entries = append(lastEvent.Entries, entries...)
	obsmap.Events[len(obsmap.Events)-1] = lastEvent

	WriteObsolescenceMap(obsmap, filepath)
	return nil
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
