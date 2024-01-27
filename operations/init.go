package operations

import (
	"embed"
	"fmt"
	"os"
	"regexp"

	gitutil "github.com/acamadeo/git-tree/git"
	"github.com/acamadeo/git-tree/models"
	"github.com/acamadeo/git-tree/store"
	"github.com/acamadeo/git-tree/utils"
	git "github.com/libgit2/git2go/v34"
)

const preRebaseFilename = "scripts/git-tree-pre-rebase.sh"
const postRewriteFilename = "scripts/git-tree-post-rewrite.sh"
const preCommitFilename = "scripts/git-tree-pre-commit.sh"
const postCommitFilename = "scripts/git-tree-post-commit.sh"

//go:embed scripts/*
var gitHookScripts embed.FS

// Initialize git-tree for the given repository.
//
// Track the specified branches, or all branches if none were specified.
func Init(repo *git.Repository, branches ...*git.Branch) error {
	// If there were no branches passed in, initialize with all local branches.
	if len(branches) == 0 {
		branches = gitutil.AllLocalBranches(repo)
	}

	// Create the root branch as the most-common ancestor of the provided
	// branches.
	rootBranch, err := createRootBranch(repo, branches)
	if err != nil {
		return fmt.Errorf("Could not create temporary root branch: %s.", err.Error())
	}

	// Construct a branch map from the branches and store the branch map in our
	// file.
	branchMap := models.BranchMapFromRepo(repo, rootBranch, branches)
	store.WriteBranchMap(branchMap, store.BranchMapPath(repo.Path()))

	// Install `post-commit` and `post-rewrite` git-hooks.
	installGitHooks(repo)

	return nil
}

func createRootBranch(repo *git.Repository, branches []*git.Branch) (*git.Branch, error) {
	// Find the commit that will serve as the root of the git-tree. Create a new
	// branch pointed to this commit.
	rootOid := gitutil.MergeBaseMany_Branches(repo, branches...)
	rootCommit, _ := repo.LookupCommit(rootOid)
	return repo.CreateBranch(store.GitTreeRootBranch, rootCommit, false)
}

func installGitHooks(repo *git.Repository) {
	// `pre-rebase` hook
	hookFile := repo.Path() + "hooks/pre-rebase"
	destFilename := repo.Path() + "hooks/git-tree-pre-rebase.sh"
	installGitHook(hookFile, preRebaseFilename, destFilename)

	// `post-rewrite` hook
	hookFile = repo.Path() + "hooks/post-rewrite"
	destFilename = repo.Path() + "hooks/git-tree-post-rewrite.sh"
	installGitHook(hookFile, postRewriteFilename, destFilename)

	// `pre-commit` hook
	hookFile = repo.Path() + "hooks/pre-commit"
	destFilename = repo.Path() + "hooks/git-tree-pre-commit.sh"
	installGitHook(hookFile, preCommitFilename, destFilename)

	// `post-commit` hook
	hookFile = repo.Path() + "hooks/post-commit"
	destFilename = repo.Path() + "hooks/git-tree-post-commit.sh"
	installGitHook(hookFile, postCommitFilename, destFilename)
}

func installGitHook(hookFile string, sourceFilename string, destFilename string) {
	// Copy `scripts/git-tree-post-{}.sh` into `.git/hooks/`.
	sourceFile, _ := gitHookScripts.ReadFile(sourceFilename)
	utils.OverwriteFile(destFilename, string(sourceFile))

	// Call `git-tree-post-{}.sh` in the `post-{}` hook.
	contents := utils.ReadFile(hookFile)
	if matched, _ := regexp.MatchString(`^#!.*`, contents); !matched {
		utils.PrependToFile(hookFile, "#!/bin/bash")
	}
	utils.AppendToFile(hookFile, fmt.Sprintf(`%s "$@"`, destFilename))

	// Mark `git-tree-post-{}.sh` and `.git/hooks/post-{}` as executable.
	os.Chmod(destFilename, 0755)
	os.Chmod(hookFile, 0755)
}
