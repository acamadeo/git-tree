package common

import "path/filepath"

const GitTreeRootBranch = "git-tree-root"

const GitTreeSubdir = "tree"
const GitTreeBranchMap = "tree/branches"
const GitTreeObsMap = "tree/obsmap"

func GitTreeSubdirPath(gitPath string) string {
	return GitTreeFilePath(gitPath, GitTreeSubdir)
}

func BranchMapPath(gitPath string) string {
	return GitTreeFilePath(gitPath, GitTreeBranchMap)
}

func GitTreeFilePath(gitPath string, filename string) string {
	return filepath.Join(gitPath, filename)
}
