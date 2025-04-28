#!/bin/bash

set -ve

git fetch --tags --quiet
git checkout "$(git tag --list 'v*' | sort -Vr | head -n 1)"
go install ./cmd/kask

git checkout dev --
$(go env GOPATH)/bin/kask build -in docs -out docs-build -domain / -v
