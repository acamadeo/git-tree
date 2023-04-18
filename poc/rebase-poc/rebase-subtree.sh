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

# Rebasing a tree of branches
#   `git-tree rebase -s c -d a`
# =============================
# Tree:
#   a - b - c - d ┬ e
#                 └ f - g

# Construct the initial tree
git init

append-branch a
append-branch b
append-branch c
append-branch d
append-branch e
switch-branch d
append-branch f
append-branch g

# Rebase first branch onto a (using segment method).
git checkout c
git checkout -b c2
git rebase --onto a b c

# Do the same with `d` as it is common to both branches of the subtree.
git checkout d
git checkout -b d2
git rebase --onto c c2 d

# DO DEPTH-FIRST SEARCH, as it allows users to address merge conflicts as they
# propagate through descendants.
#  - Start with rebasing branch `e`
git checkout e
git checkout -b e2
git rebase --onto d d2 e

#  - Then go down the `f` side
git checkout f
git checkout -b f2
git rebase --onto d d2 f

git checkout g
git checkout -b g2
git rebase --onto f f2 g

# Delete the temporary branches
git branch -D c2 d2 e2 f2 g2
