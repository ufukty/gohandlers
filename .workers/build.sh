#!/bin/bash

set -ve

(
  cd "$(mktemp -d)"
  git clone --branch v0.16.1 --depth 1 https://github.com/ufukty/kask .
  make install
)

git fetch --tags --quiet

if test "$WORKERS_CI_BRANCH" = "main"; then
  git checkout "$(git tag --list 'v*' --sort '-version:refname' | head -n 1)"
fi

~/bin/kask build -in docs -out docs-build -domain "https://gohandlers.ufukty.com" -v -cfw
