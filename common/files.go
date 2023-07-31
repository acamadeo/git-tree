package common

import "path/filepath"

const GitTreeRootBranch = "git-tree-root"

const GitTreeSubdir = "tree"
const GitTreeBranchMap = "tree/branches"
const GitTreeObsMap = "tree/obsmap"

const GitTreeRebasing = "tree/rebasing"
const GitTreeRebasingSource = "tree/rebasing-source"
const GitTreeRebasingDest = "tree/rebasing-dest"
const GitTreeRebasingTemps = "tree/rebasing-temps"

func GitTreeSubdirPath(gitPath string) string {
	return GitTreeFilePath(gitPath, GitTreeSubdir)
}

func BranchMapPath(gitPath string) string {
	return GitTreeFilePath(gitPath, GitTreeBranchMap)
}

func RebasingPath(gitPath string) string {
	return GitTreeFilePath(gitPath, GitTreeRebasing)
}

func RebasingSourcePath(gitPath string) string {
	return GitTreeFilePath(gitPath, GitTreeRebasingSource)
}

func RebasingDestPath(gitPath string) string {
	return GitTreeFilePath(gitPath, GitTreeRebasingDest)
}

func RebasingTempsPath(gitPath string) string {
	return GitTreeFilePath(gitPath, GitTreeRebasingTemps)
}

func GitTreeFilePath(gitPath string, filename string) string {
	return filepath.Join(gitPath, filename)
}
