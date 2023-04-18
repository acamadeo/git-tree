#!/bin/bash

append-branch() {
    git checkout -b $1
    echo $1 > $1.txt
    git add .
    git commit -m "Add $1"
}

switch-branch() {
    git checkout -b $1
}

reset() {
    rm *
    rm -rf .git/
}

# Perform `git-tree rebase -s c -d a` successively
# ================================================
# Tree:
#   a - b - c - d - e

# Construct the initial tree
git init

append-branch a
append-branch b
append-branch c
append-branch d
append-branch e

# Start at the first branch to rebase. Make a copy of the branch pointer, then
# rebase it to the destination.
git checkout c
git checkout -b c2
git rebase --onto a b c

# Do the same thing for the descedant of `c`.
git checkout d
git checkout -b d2
git rebase --onto c c2 d

# Do the same thing for the descendant of `d`.
git checkout e
git checkout -b e2
git rebase --onto d d2 e

# Delete the temporary branches
git branch -D c2 d2 e2
