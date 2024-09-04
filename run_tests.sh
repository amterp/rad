#!/usr/bin/env bash

# Define the list of invocations
invocations=(
    "test1.rad samber/lo"
    "test2.rad alice bobson"
    "test3.rad tomnomnom/gron"
    "test3.rad --repo tomnomnom/gron"
    "test3.rad --repo tomnomnom/gron --limit 10"
    "test3.rad --limit 10 --repo tomnomnom/gron"
    "test4.rad alice,bob,charlie"
)

go build main.go || exit 1

# Loop through the invocations
any_failed=false
for invocation in "${invocations[@]}"
do
    cmd="./main ./tests/$invocation"
    echo -n "$cmd - "

    # Run 'go run main.go' with the invocation and capture the exit status
    eval "$cmd" > /dev/null 2>&1
    exit_status=$?

    # Check if the command succeeded (exit status 0) or failed (non-zero exit status)
    if [ $exit_status -eq 0 ]; then
        echo -e "\033[1;32mSuccess\033[0m"
    else
        echo -e "\033[1;31mFailed\033[0m"
        any_failed=true
    fi
done

if [ "$any_failed" = true ]; then
    exit 1
fi
