package common

import "path/filepath"

const GitTreeRootBranch = "git-tree-root"

const GitTreeSubdir = "tree"
const GitTreeBranchMap = "tree/branches"
const GitTreeObsMap = "tree/obsmap"

const GitTreeRebasing = "tree/rebasing"

func GitTreeSubdirPath(gitPath string) string {
	return GitTreeFilePath(gitPath, GitTreeSubdir)
}

func BranchMapPath(gitPath string) string {
	return GitTreeFilePath(gitPath, GitTreeBranchMap)
}

func RebasingPath(gitPath string) string {
	return GitTreeFilePath(gitPath, GitTreeRebasing)
}

func GitTreeFilePath(gitPath string, filename string) string {
	return filepath.Join(gitPath, filename)
}
