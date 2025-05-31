---
title: Resources
---

Depending on your script, you may wish to look up values at some point, such as URL endpoints depending on your input. Rad offers a function for this, `pick_from_resource`, but before we dive into it, we'll first cover a couple of related functions.

## `pick`

[`pick`](../reference/functions.md#pick) is an in-built function which allows you to choose one value from a list of inputs, using some filter. If the filter has several matches, Rad will enter an interactive mode which allows the user to pick a single value to continue with.

You can try it yourself with this example:

```rad
options = ["chicken burger", "chicken sandwich", "ham sandwich"]
output = pick(options, "sandwich")
print("You chose: {output}")
```

When you initially run this, the `sandwich` filter should exclude `chicken burger` and ask you to select between two remaining options:

```
┃ Pick an option
┃ > chicken sandwich
┃   ham sandwich
```

After you choose, you get the final output:

```
You chose: chicken sandwich
```

## `pick_kv`

A similar function is [`pick_kv`](../reference/functions.md#pick_kv). However, instead of the values you're filtering and picking between also being the *output*, `pick_kv` performs the filtering/picking on a list of *keys*, each which map to a value that will get output from the function if its associated key is picked. For example:

```rad
keys = ["chicken burger", "chicken sandwich", "ham sandwich"]
values = ["CHICKEN", "CHICKEN", "HAM"]

output = pick_kv(keys, values)

print("We'll need {output}!")
```

In this example, we leave out the filter, as it's optional, which will launch us into an interactive select between all the key values:

```
┃ Pick an option
┃   chicken burger
┃   chicken sandwich
┃ > ham sandwich
```

If we pick this third option, this is the final output of the script:

```
We'll need HAM!
```

Notice that the function did not output the *key* `ham sandwich` that was selected, but instead the *value* `HAM` that it mapped to.

## `pick_from_resource`

Now we'll look at actually using what this section is about - resources. [`pick_from_resource`](../reference/functions.md#pick_from_resource) allows you to pre-define a resource file (using JSON) which contains a range of key-value pairs. When invoked, it will behave similarly to the two previous `pick` functions i.e. it lets you apply an optional filter, and will launch into an interactive picking mode to narrow down a single choice, if needed.

Let's do a simple example. As mentioned, a resource file is simply a JSON file. We'll create an example where we look up a url based on user input:

```json title="websites.json"
{
  "options": [
    {
      "keys": ["gl", "lab"],
      "values": ["gitlab.com", "GitLab"]
    },
    {
      "keys": ["gh", "hub"],
      "values": ["github.com", "GitHub"]
    }
  ]
}
```

You may see some similarity here to what we did with [`pick_kv`](#pick_kv). We're defining two options: one which can get matched by either `gl` or `lab`, and one which gets matched by `gh` or `hub`. In the first case, if chosen, `pick_from_resource` will return *two* values: `gitlab.com` and `GitLab`. Similarly it will return `github.com` and `GitHub` for the latter.

Let's create an Rad script to use this resource:

```rad title="example.rad"
args:
    website string = ""

url, name = pick_from_resource("./resources/websites.json", website)
print("url: {url}, name: {name}")
```

Note that the first argument to `pick_from_resource` is a path to a resource file. This path is *relative to the script's path*.
This allows you to store your resources with your scripts. In this example, we'll place our files like so:

```
.
├── example.rad
└── resources
    └── websites.json
```

This means that it doesn't matter where on your computer you invoke your script from including if it's on your PATH - the script will consistently look in the same spot for resource files.

TBC

[//]: # (todo make pick_from_resource interactively select!)

## Summary

- `pick` and `pick_kv` are built-in functions that allow users to select one option from many, allowing for an optional filter.
- `pick_from_resource` is similar, but uses a pre-defined resource file to define the options.
- The resource file is defined in JSON.
- The resource file path can be defined relative to the script's path.

## Next

The shell offers a ton of useful utilities, and Rad allows you to leverage them from within your scripts.

We'll look at that in the next section: [Shell Commands](./shell-commands.md).
