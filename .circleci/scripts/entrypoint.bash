#!/usr/bin/env bash

if [ $# -ne 1 ]; then
  echo 'please use from Makefile'
  exit 1
fi

if [ "$GITHUB_TOKEN" = "" ]; then
  echo 'please set $GITHUB_TOKEN'
  exit 1
fi

LATEST_VERSION=$(curl -H "Authorization: token ${GITHUB_TOKEN}" https://api.github.com/repos/ktr0731/evans/releases | jq '.[0].tag_name')

if [ "$VERSION" = "$LATEST_VERSION" ]; then
  echo 'same version'
  exit
fi

bash .circleci/scripts/build-assets.bash "$VERSION"
bash .circleci/scripts/create-new-release.bash "$VERSION"
bash .circleci/scripts/update-brew-formula.bash "$VERSION"
