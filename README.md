# Git Tree

Tool to more easily manage trees of dependent Git branches. Key features:

 - Rebase Git branches *and their descendants* onto one another.
 - Evolve changes in a branch to all descendant branches.

## Installation

```sh
git clone https://github.com/acamadeo/git-tree.git ~/.git-tree
~/.git-tree/install
```

# Building from source

TODO: Add instructions on setting up the required dependencies (i.e. libgit2).

Run the following command to build the CLI:

```shell
go build -tags static,system_libgit2 .
```

# Running tests

To run tests across all packages, issue the following:

```shell
go test -tags static,system_libgit2 ./...  -v
```
