package operations

import (
	"embed"
	"fmt"
	"os"
	"regexp"

	"github.com/acamadeo/git-tree/common"
	gitutil "github.com/acamadeo/git-tree/git"
	"github.com/acamadeo/git-tree/models"
	"github.com/acamadeo/git-tree/store"
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
	store.WriteBranchMap(branchMap, common.BranchMapPath(repo.Path()))

	// Install `post-commit` and `post-rewrite` git-hooks.
	installGitHooks(repo)

	return nil
}

func createRootBranch(repo *git.Repository, branches []*git.Branch) (*git.Branch, error) {
	var rootOid *git.Oid
	if len(branches) == 1 {
		rootOid = branches[0].Target()
	} else {
		// Find the commit that will serve as the root of the git-tree. Create a new
		// branch pointed to this commit.
		rootOid, _ = repo.MergeBaseMany(branchOids(branches))
	}

	rootCommit, _ := repo.LookupCommit(rootOid)
	return repo.CreateBranch(common.GitTreeRootBranch, rootCommit, false)
}

func branchOids(branches []*git.Branch) []*git.Oid {
	oids := []*git.Oid{}
	for _, branch := range branches {
		oids = append(oids, branch.Target())
	}
	return oids
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
	store.OverwriteFile(destFilename, string(sourceFile))

	// Call `git-tree-post-{}.sh` in the `post-{}` hook.
	contents := store.ReadFile(hookFile)
	contents = addPrefixIfNoPattern(contents, `^#!.*`, "#!/bin/bash\n")
	contents += fmt.Sprintf(`%s "$@"`, destFilename)
	store.OverwriteFile(hookFile, contents)

	// Mark `git-tree-post-{}.sh` and `.git/hooks/post-{}` as executable.
	os.Chmod(destFilename, 0755)
	os.Chmod(hookFile, 0755)
}

// Prepend `contents` with `prefix` if `contents` does not contain the regex
// `pattern`.
func addPrefixIfNoPattern(contents string, pattern string, prefix string) string {
	matched, _ := regexp.MatchString(pattern, contents)
	if !matched {
		return prefix + contents
	}
	return contents
}
