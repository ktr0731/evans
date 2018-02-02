#!/usr/bin/env bash

if [ $# -ne 1 ]; then
  echo 'please use from Makefile'
  exit 1
fi

if [ "$GIT_EMAIL" = "" ] || [ "$GIT_NAME" = "" ]; then
  echo 'please set $GIT_EMAIL and $GIT_NAME'
  exit 1
fi

VERSION=$1

git clone https://github.com/ktr0731/brew-evans brew

cd brew

sed -i -r "s/[0-9]+\.[0-9]+\.[0-9]+/$VERSION/" evans.rb
git add --all
git commit -m "bump ${VERSION}"
git push https://github.com/ktr0731/brew-evans master
