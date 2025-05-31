---
title: Defer & Errdefer
---

With defer statements, you can specify blocks of code to be run before the script exits.

You may wish to use them to clean up or undo operations before exiting.

## Defer

`defer` blocks always run before the script exits, regardless of if it's due to an error or not.

Here is an example:

```rad title="defer.rsl"
$!`mv notes.txt notes-tmp.txt`
defer:
    $!`mv notes-tmp.txt notes.txt`
    print("Moved back!")

$!`echo "hi!" >> notes.txt`
$!`cat notes.txt`
```

Let's say we already have a file `notes.txt` containing some text. In this script, we take the following steps, largely by invoking [shell commands](./shell-commands.md):

1. Rename `notes.txt` to `notes-tmp.txt`
2. Define a `defer` block which will undo the rename, and then print `Moved back!`.
3. Create a *new* `notes.txt` with the contents `hi!`
4. Print the contents `notes.txt` to show what it contains.

When run:

```shell
rad defer.rl
```

<div class="result">
```
⚡️ Running: mv notes.txt notes-tmp.txt
⚡️ Running: echo "hi!" >> notes.txt
⚡️ Running: cat notes.txt
hi!
⚡️ Running: mv notes-tmp.txt notes.txt
Moved back!
```
</div>

Note that despite the `Moved back!` print statement appearing *earlier* in the script, it only gets run at the end due to being in a `defer` block.

## Errdefer

Sometimes, you only want certain deferred statements to run in the event of a failure.
This is useful when your script is working toward a critical step that, once executed, should not be rolled back.
However, if the script fails before reaching that step, rollback actions may still be necessary.

Below is an example of a version-bumping script. Using `sed`, this script replaces the version in a file called `VERSION`, stages the file with git,
and commits it. However, if there's a failure in between the `sed` and `commit` steps, then we want to undo earlier steps as a cleanup, in order to
make the script *atomic* i.e. it either succeeds entirely or does nothing, leaving no intermediary state changes behind. We accomplish this through `errdefer` blocks. 

```rad title="bump.rsl"
args:
    version string

path = "VERSION"

$!`sed -i '' "s/Version = .*/Version = {version}/" {path}`
errdefer:
    print("Undoing bump...")
    $!`git checkout -- {path}`

if false:  // failure simulation point 1
    print("Oh no! ERROR!")
    exit(1)

$!`git add {path}`
errdefer:
    print("Resetting {path}...")
    _, _ = $!`git reset {path}`

if false:  // failure simulation point 2
    print("Bah! ERROR!")
    exit(1)

$!`git commit -m "Bump version to {version}"`
print("Done!")
```

We include a couple of "failure points". We can set their condition to `true` to have them simulate an error, as the exit code of '1' indicates failure.

Before we do that though, we can see an example of this script working correctly. Let's say we define our `VERSION` file in the same directory as the script as follows:

```txt title="VERSION"
Version = 1
```

If we execute our script, we get the following output:

```shell
rad bump.rl 2
```

<div class="result">
```
⚡️ Running: sed -i '' "s/Version = .*/Version = 2/" VERSION
⚡️ Running: git add VERSION
⚡️ Running: git commit -m "Bump version to 2"
[main 6ce2ebb] Bump version to 2
 1 file changed, 1 insertion(+), 1 deletion(-)
Done!
```
</div>

We can see the series of commands get run as we expect, including output from git. Notice that none of our `errdefer` blocks ran, because there were no failures.

**Now let's say that we activate failure point 1** by setting its condition to `true`. This means that, after performing the `sed` command, but before we `git add`, the script exits, and we trigger just the first `errdefer` block to 'reset' the `VERSION` file.

```shell
rad bump.rl 3
```

<div class="result">
```
⚡️ Running: sed -i '' "s/Version = .*/Version = 3/" VERSION
Oh no! ERROR!
Undoing bump...
⚡️ Running: git checkout -- VERSION
```
</div>

If you run this locally, you should see with `git status` that there are no changes to the `VERSION` file, thanks to our `errdefer` block rolling back the `sed` replacement.

Next let's try deactivating failure point 1 again and enabling failure point 2, and running our script again. This time, we can expect the `git add` to run, and our failure will occur after, but before the `git commit`.

```shell
rad bump.rl 3
```

<div class="result">
```
⚡️ Running: sed -i '' "s/Version = .*/Version = 3/" VERSION
⚡️ Running: git add VERSION
Bah! ERROR!
Resetting VERSION...
⚡️ Running: git reset VERSION
Undoing bump...
⚡️ Running: git checkout -- VERSION
```
</div>

Here we see a very important detail about defer blocks that applies both to `defer` and `errdefer` - if you have multiple, they run in LIFO (last in, first out) order. In other words, the defer blocks defined *later* run *first*.

This is typically desirable, as this example demonstrates. After we `git commit`, we need to first `git reset`, otherwise the `git checkout` to undo the bump won't work. Thanks to LIFO, our `git reset` runs first and all is good.

## Errors in defer blocks

If a script exits successfully and has multiple `defer` blocks, and the first one to run encounters an error, **the remaining defer blocks still run**. This also applies to `errdefer`. However, the script will exit with a non-0 error code.

`errdefer` blocks are only triggered if the *main script* fails. If the main script runs successfully, but a `defer` block then errors, that does *not* trigger `errdefer` blocks to run.

## Summary

- Use `defer` and `errdefer` blocks to run operations after your main script ends.
- They can commonly be used for clean up operations and making your scripts atomic.
- Defer blocks run in LIFO order - last in, first out.

[//]: # (todo next? or end guide?)
