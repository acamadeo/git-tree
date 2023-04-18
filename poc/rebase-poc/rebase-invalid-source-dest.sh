#!/bin/bash

append-branch() {
    git checkout -b $1
    echo $1 > $1.txt
    git add .
    git commit -m "Add $1"
}

switch-branch() {
    git checkout $1
}

reset() {
    rm *
    rm -rf .git/
}

# Invalid source and dest rebases
# ===============================
# Tree:
#   a - b - c - d ┬ e
#                 └ f - g

# Errors:
# `git-tree rebase -s a -d a`
#    ERROR: source and destination cannot be the same
# `git-tree rebase -s a -d b`
#    ERROR: source cannot be an ancestor of destination
# `git-tree rebase -s a -d e`
#    ERROR: source cannot be an ancestor of destination
# `git-tree rebase -s c -d g`
#    ERROR: source cannot be an ancestor of destination
# `git-tree rebase -s b -d a`
#    There is nothing to rebase
