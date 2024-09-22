#!/usr/bin/env bash

rsl="\
---
Bumps the version in rad, creates a commit, tags it, and optionally pushes it

Bumps the version in rad, creates a commit, tags it, and optionally pushes it
This will trigger a GitHub action to create a homebrew-rad PR with the new version
---
args:
    new_version string # The new release version to bump to
    push p bool # Whether or not a push should also be performed"

new_version=
push=
eval "$(rad --SHELL --STDIN "$0" "$@" <<< "$rsl")"

# Build & test
./push.sh --no-push || exit 1

# Update Version in ./core/cobra_root.go
sed -i '' "s/Version: \".*\"/Version: \"$new_version\"/" ./core/cobra_root.go || exit 1
git add ./core/cobra_root.go
git commit -m "Bump version to $new_version"
git tag -a "$new_version" -m "Bump version to $new_version" || exit 1

if [ "$push" = true ]; then
    ./push.sh || exit 1
else
    echo "Tagged, not pushing..."
    echo "Run 'git push origin main --tags' to push the tag"
fi
