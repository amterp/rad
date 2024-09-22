#!/usr/bin/env bash

rsl="\
---
Performs some pre-pushing steps such as build, test, and then pushes if all good.
---
args:
  branch b string = 'main' # The branch to push to
  no_push 'no-push' n bool = false # Skip the pushing step
"

branch=""
no_push=false
eval "$(rad --SHELL --STDIN "$0" "$@" <<< "$rsl")"

echo "Building..."
if ! go build main.go; then
    echo "Build failed ❌"
    exit 1
fi

# Run tests many times to ensure no flakiness
echo "✅ Testing..."
if ! go test ./core/testing -count 50; then
    echo "Tests failed ❌"
    exit 1
fi

if [ "$no_push" = true ]; then
    echo "Skipping push, done!"
    exit 0
fi

echo "✅ Pushing..."
git push origin "$branch" || exit 1
echo "✅ Pushed!"
