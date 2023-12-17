package store

import (
	"errors"
	"fmt"
	"strings"

	"github.com/acamadeo/git-tree/models"
	"github.com/acamadeo/git-tree/utils"
	git "github.com/libgit2/git2go/v34"
)

var actionTypeStrings = map[models.ActionType]string{
	models.ActionTypeUnknown: "unknown",
	models.ActionTypeRebase:  "rebase",
	models.ActionTypeAmend:   "amend",
	models.ActionTypeCommit:  "commit",
}

var hookTypeStrings = map[models.HookType]string{
	models.HookTypeUnknown:   "unknown",
	models.PostRewriteAmend:  "post-rewrite.amend",
	models.PostRewriteRebase: "post-rewrite.rebase",
	models.PostCommit:        "post-commit",
}

// Read obsolescence map file
func ReadObsolescenceMap(repo *git.Repository, filepath string) *models.ObsolescenceMap {
	contents := utils.ReadFile(filepath)
	if contents == "" {
		return &models.ObsolescenceMap{}
	}

	obsmap := models.ObsolescenceMap{}
	for _, line := range strings.Split(contents, "\n") {
		// Check for the start of a new action.
		if strings.HasPrefix(line, "action") {
			lineParts := strings.Fields(line)
			obsmap.Actions = append(obsmap.Actions, models.ObsolescenceAction{
				ActionType: actionTypeFromString(lineParts[1]),
			})
			continue
		}

		// Line does not indicate the start of a new action. Append an entry to
		// the latest action.
		lastAction := obsmap.Actions[len(obsmap.Actions)-1]
		lastAction.Entries = append(lastAction.Entries,
			obsolescenceEntryFromLine(repo, line))
		obsmap.Actions[len(obsmap.Actions)-1] = lastAction
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

func actionTypeFromString(value string) models.ActionType {
	for actionType, str := range actionTypeStrings {
		if str == value {
			return actionType
		}
	}
	return models.ActionTypeUnknown
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
	utils.OverwriteFile(filepath, obsolescenceMapString(obsmap))
}

func LastObsolescenceActionType(repo *git.Repository, filepath string) models.ActionType {
	obsmap := ReadObsolescenceMap(repo, filepath)
	return obsmap.Actions[len(obsmap.Actions)-1].ActionType
}

func SetLastObsolescenceActionType(repo *git.Repository, filepath string, actionType models.ActionType) {
	obsmap := ReadObsolescenceMap(repo, filepath)
	obsmap.Actions[len(obsmap.Actions)-1].ActionType = actionType
	WriteObsolescenceMap(obsmap, filepath)
}

func AppendObsolescenceAction(repo *git.Repository, filepath string, ActionType models.ActionType) {
	obsmap := ReadObsolescenceMap(repo, filepath)

	obsmap.Actions = append(obsmap.Actions, models.ObsolescenceAction{
		ActionType: ActionType,
	})
	WriteObsolescenceMap(obsmap, filepath)
}

// Append entries to obsolescence map file under the last ObsolescenceAction.
func AppendEntriesToLastObsolescenceAction(repo *git.Repository, filepath string, entries ...models.ObsolescenceEntry) error {
	obsmap := ReadObsolescenceMap(repo, filepath)
	if len(obsmap.Actions) < 1 {
		return errors.New("cannot append entry to obsolete map without actions")
	}

	lastAction := obsmap.Actions[len(obsmap.Actions)-1]
	lastAction.Entries = append(lastAction.Entries, entries...)
	obsmap.Actions[len(obsmap.Actions)-1] = lastAction

	WriteObsolescenceMap(obsmap, filepath)
	return nil
}

func obsolescenceMapString(obsmap *models.ObsolescenceMap) string {
	output := []string{}

	for _, action := range obsmap.Actions {
		actionHeader := fmt.Sprintf("action %s",
			actionTypeStrings[action.ActionType])
		output = append(output, actionHeader)

		for _, entry := range action.Entries {
			entryString := fmt.Sprintf("%s %s %s",
				entry.Commit.Id().String(),
				entry.Obsoleter.Id().String(),
				hookTypeStrings[entry.HookType])
			output = append(output, entryString)
		}
	}

	return strings.Join(output, "\n")
}
