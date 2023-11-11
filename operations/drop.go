package operations

import (
	"fmt"
	"os"

	"github.com/acamadeo/git-tree/common"
	"github.com/acamadeo/git-tree/store"
	git "github.com/libgit2/git2go/v34"
)

// Remove git-tree tracking for the given repository.
func Drop(repo *git.Repository) error {
	// Read the branch map file.
	branchMapPath := common.BranchMapPath(repo.Path())
	branchMap := store.ReadBranchMap(repo, branchMapPath)

	// Delete the root branch created by `git-tree init`.
	if err := branchMap.Root.Delete(); err != nil {
		return fmt.Errorf("Could not delete root branch: %s.", err.Error())
	}

	// Delete local git-tree storage (i.e. the branch map and obsolescence map
	// files).
	gitTreePath := common.GitTreeSubdirPath(repo.Path())
	if err := os.RemoveAll(gitTreePath); err != nil {
		return fmt.Errorf("Could not delete git-tree files: %s.", err.Error())
	}

	// Remove `post-commit` and `post-rewrite` git-hooks for git-tree.
	uninstallGitHooks(repo)

	return nil
}

func uninstallGitHooks(repo *git.Repository) {
	// `pre-rebase` hook
	hookFile := repo.Path() + "hooks/pre-rebase"
	hookImplFilename := repo.Path() + "hooks/git-tree-pre-rebase.sh"
	uninstallGitHook(hookFile, hookImplFilename)

	// `post-rewrite` hook
	hookFile = repo.Path() + "hooks/post-rewrite"
	hookImplFilename = repo.Path() + "hooks/git-tree-post-rewrite.sh"
	uninstallGitHook(hookFile, hookImplFilename)

	// `pre-commit` hook
	hookFile = repo.Path() + "hooks/pre-commit"
	hookImplFilename = repo.Path() + "hooks/git-tree-pre-commit.sh"
	uninstallGitHook(hookFile, hookImplFilename)

	// `post-commit` hook
	hookFile = repo.Path() + "hooks/post-commit"
	hookImplFilename = repo.Path() + "hooks/git-tree-post-commit.sh"
	uninstallGitHook(hookFile, hookImplFilename)
}

func uninstallGitHook(hookFile string, hookImplFilename string) {
	// Delete `.git/hooks/git-tree-post-{}.sh`.
	os.Remove(hookImplFilename)

	// Delete call to `git-tree-post-{}.sh` in the `post-{}` hook.
	scriptCall := fmt.Sprintf(`%s "$@"`, hookImplFilename)
	store.DeleteLineInFile(hookFile, scriptCall)
}
