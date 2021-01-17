#!/bin/bash

set -e -o pipefail

if [ $# -ne 1 ]; then
  echo 'please use from Makefile'
  exit 1
fi

if [ "$GITHUB_TOKEN" = "" ]; then
  echo 'please set $GITHUB_TOKEN'
  exit 1
fi

LATEST_VERSION=$(curl -H "Authorization: token ${GITHUB_TOKEN}" https://api.github.com/repos/ktr0731/evans/releases | jq -r '.[0].tag_name')
VERSION=$1

if [ "$LATEST_VERSION" = "" ] || [ "$VERSION" = "$LATEST_VERSION" ]; then
  echo 'same version'
  exit 0
fi

git tag "$VERSION"
git tag "v$VERSION" "$VERSION" # tagging another one which has "v" prefix for Go Modules.
git push origin "$VERSION"
git push origin "v$VERSION"

goreleaser --skip-validate
