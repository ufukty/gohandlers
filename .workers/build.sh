#!/bin/bash

set -ve

if test "$WORKERS_CI_BRANCH" = "main"; then
  git fetch --tags --quiet
  git checkout "$(git tag --list 'v*' --sort '-version:refname' | head -n 1)"
fi

GOBIN="$PWD" go install go.ufukty.com/kask@v0.17.2
./kask build -in docs -out docs-build -domain "https://gohandlers.ufukty.com" -v -cfw
