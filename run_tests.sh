#!/usr/bin/env bash

# Define the list of invocations
invocations=(
    "./main ./tests/test1.rad samber/lo"
    "./main ./tests/test2.rad alice bobson"
    "./main ./tests/test3.rad tomnomnom/gron"
    "./main ./tests/test3.rad --repo tomnomnom/gron"
    "./main ./tests/test3.rad --repo tomnomnom/gron --limit 10"
    "./main ./tests/test3.rad --limit 10 --repo tomnomnom/gron"
    "./main ./tests/test4.rad alice,bob,charlie"
    "./main ./tests/test5.rad --repo samber/lo"
    "./main ./tests/test6.rad"
    "./main ./tests/test7.rad --arr11=2.1,2.2"
    "./main ./tests/test8.rad"
    "./main ./tests/test9.rad"
    "./main ./tests/test10.rad"
    "./main ./tests/test11.rad"
    "./main ./tests/test12.rad"
    "./main ./tests/test13.rad"
    "./main ./tests/test14.rad"
    "./main ./tests/test15.rad"
    "./main ./tests/test16.rad"
    "./main ./tests/test17.rad"
    "./tests/test18.sh --name alice"
    "./tests/test18.sh --help"
    "./main ./tests/test19.rad --help"
    "./main ./tests/date_functions.rad"
    "./main ./tests/replace_function.rad"
    "./main ./tests/join_function.rad"
    "./main ./tests/upper_lower_functions.rad"
    "./main ./tests/modify_var_in_block.rad"
    "./main ./tests/later_json_array.rad"
    "./main ./tests/single_quote_strings.rad"
    "./main ./tests/not_condition.rad"
)

go build main.go || exit 1

# Loop through the invocations
any_failed=false
for invocation in "${invocations[@]}"
do
    echo -n "$invocation - "

    # Run 'go run main.go' with the invocation and capture the exit status
    eval "$invocation" > /dev/null 2>&1
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
