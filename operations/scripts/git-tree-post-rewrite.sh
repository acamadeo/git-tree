#!/usr/bin/env bash

set -e -u
command="$1"

lines=$(</dev/stdin)
git-tree obsolete "$command" "$lines"
