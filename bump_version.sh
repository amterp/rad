#!/usr/bin/env bash

rsl="\
---
Bumps the version in rad, creates a commit, tags it, and optionally pushes it
---
args:
    new_version string # The new version to bump to
    push p bool # Whether or not a push should also be performed"

go build main.go || exit 1
eval "$(./main --SHELL --STDIN "$0" "$@" <<< "$rsl")"

# Update Version in ./core/cobra_root.go
sed -i '' "s/Version: \".*\"/Version: \"$new_version\"/" ./core/cobra_root.go || exit 1
git add ./core/cobra_root.go
git commit -m "Bump version to $new_version"
git tag -a "$new_version" -m "Bump version to $new_version" || exit 1

if [ "$push" = true ]; then
    echo "Pushing..."
    git push origin main --tags || exit 1
else
    echo "Tagged, not pushing..."
    echo "Run 'git push origin main --tags' to push the tag"
fi
