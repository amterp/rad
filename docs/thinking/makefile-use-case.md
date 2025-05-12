# Makefile Use Case

I've used RSL quite a lot of times to make 'dev' scripts which sort of act like programmable Makefiles. Example from the one in this repo:

```
#!/usr/bin/env rad
---
Facilitates developing rad.
---
args:
    version v int = 0 # 1 to bump patch, 2 for minor and 3 for major.
    docs d bool       # Enable to deploy docs.
    amend a bool      # Enable to amend the last commit with the version bump instead of creating a new commit.
    push p bool       # Enable to push local commits.

get_branch = fn():
    _, branch  = $!`echo -n $(git branch --show-current)`
    return branch

if version:
    $!`make all`

if docs:
    $!`mkdocs gh-deploy -f ./docs-web/mkdocs.yml`

if version:
    // ... truncated, resolve 'version' via git tags

    $!`go test ./core/testing -count 10`

    clean_before = unsafe $`git status --porcelain | grep -q .`

    $!`sed -i '' "s/Version = \".*\"/Version = \"{version}\"/" ./core/version.go`

    $!`git add .`
    $!`git diff HEAD --compact-summary`

    if confirm("Commit, tag & push? ({version}) [y/n] > "):
        branch = get_branch()

        if clean_before:
            // implied only our sed version bump made a diff

            if amend:
                $!`git commit --amend --no-edit`
            else:
                $!`git commit -m "Bump version to {version}"`
        else:
            $!`git commit`
        $!`git tag -a "{version}" -m "Bump version to {version}"`
        confirm $!`git push origin {branch} --tags`

if push:
    $!`make all`
    branch = get_branch()
    $!`git push origin {branch} --tags`

print(green("✅ Done!"))
```

Can see that they mostly consist of toggleable steps. For example, these two might be equivalent for this rad version vs. make:

```
./dev -v 1 -d
make docs VERSION=1
```

Not sure I've got that arg right for make, ultimately rad has the advantage there, so not much benefit in lingering on it.

What make *does* have on rad is the ability to omit the `./`.
In order for rad to omit this, we'd need a command on the user's PATH which invokes the dev script.

Couple of potential options:

1. Allow users to put a 'passthrough' script on their path which can invoke a local script with passthrough args. 
2. `rad` somehow offers in-built, tailored functionality to automatically invoke e.g. `Make.rsl` files it sees in your local repo.

I straight up don't think option 2 is a good idea. Feels like tailoring Rad to a specific use case that I've found useful,
but is pretty clearly out of scope as a built-in feature (imo). It would lead to complexity and there's probably no great way
to design the API for it without sacrificing its main use case.

Option 1 might be realistic, on the other hand. The idea is that the user puts a very simply script on their path,
let's call it `dev`, and it looks something like this (theoretical syntax):

```
args:
    all string...

$!`./dev {join(all, " ")}`
```

This `dev` script takes its args, stores them as a string array e.g. `[ "-v", "1", "-d" ]` and invokes the local `./dev` script
with those args, thus equivalent to `./dev -v 1 -d`.

One trick is that we probably want to ensure all the global Rad flags don't get consumed by the PATH `dev` script, and
instead *also* get passed through to the local `./dev` script for use.

The vararg `string...` approach is probably not the way to go, here. It has a potential use case in the language, but I'm not sure
that involves making global flags not operate on the immediate script... Maybe. A couple of ideas:

1. No args, instead "get_args()" function + disable macros

```
---
@disable_args_validation
@disable_global_flags
---

`./dev {get_args()[1:].join(" ")}`
```

Use a macro to disable any args checking logic in Rad so that unrecognized arguments passed to it don't result in an error,
and also to disable global flags so they instead get treated as regular flags.
Also then use a new function `get_args()` to get the invocation line (including script name in index 0 probably, though
maybe we should rename the func in that case), and then pass the args onto the local `./dev`.

2. Vararg + disable all global flags

```
---
@disable_global_flags
---
args:
    all string...
```

Not sure this would actually work. Arguably:

`myscript --all hello there`

should define `all` as `[ "hello", "there" ]`, so varargs might not be the way to go. In this scenario, you could still
use `get_args` though, so the `all string...` just serves as a way to absorb args? A little hacky, don't like that.

---

So back on 1. since that seems like the best idea so far... If we allow `@disable_args_validation` as a macro, how do args work?
Like if you have a required string, but it doesn't get passed? I'd need the ability to not error and simply define it as null?

Maybe it makes more sense to change `@disable_args_validation` to simply `@disable_args_block` which forbids an args block entirely,
and implies that the user wants to do their own thing using `get_args`.
