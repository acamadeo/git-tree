package common

import "path/filepath"

type GitTreeFile int

const (
	BranchMap GitTreeFile = iota
	ObsoleteMap
	RebaseInProgress
	RebaseSource
	RebaseDest
	RebaseTemporaryBranches
)

var gitTreeFileNames = map[GitTreeFile]string{
	BranchMap:               "branches",
	ObsoleteMap:             "obsmap",
	RebaseInProgress:        "rebasing",
	RebaseSource:            "rebasing-source",
	RebaseDest:              "rebasing-dest",
	RebaseTemporaryBranches: "rebasing-temps",
}

const GitTreeRootBranch = "git-tree-root"

const GitTreeSubdir = "tree"

func GitTreeFilePath(gitPath string, fileType GitTreeFile) string {
	fileName := gitTreeFileNames[fileType]
	return filepath.Join(gitPath, GitTreeSubdir, fileName)
}

// -------------------------------------------------------------------------- \
// Ease-of-use functions                                                      |
// -------------------------------------------------------------------------- /

func GitTreeSubdirPath(gitPath string) string {
	return filepath.Join(gitPath, GitTreeSubdir)
}

func BranchMapPath(gitPath string) string {
	return GitTreeFilePath(gitPath, BranchMap)
}

func ObsoleteMapPath(gitPath string) string {
	return GitTreeFilePath(gitPath, ObsoleteMap)
}

func RebasingPath(gitPath string) string {
	return GitTreeFilePath(gitPath, RebaseInProgress)
}

func RebasingSourcePath(gitPath string) string {
	return GitTreeFilePath(gitPath, RebaseSource)
}

func RebasingDestPath(gitPath string) string {
	return GitTreeFilePath(gitPath, RebaseDest)
}

func RebasingTempsPath(gitPath string) string {
	return GitTreeFilePath(gitPath, RebaseTemporaryBranches)
}
