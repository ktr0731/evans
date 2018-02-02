#!/usr/bin/env bash

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
  exit 1
fi

git config --global user.email "$GIT_EMAIL"
git config --global user.name "$GIT_NAME"

bash .circleci/scripts/build-assets.bash "$VERSION" || exit 1
bash .circleci/scripts/create-new-release.bash "$VERSION" || exit 1
bash .circleci/scripts/update-brew-formula.bash "$VERSION" || exit 1
