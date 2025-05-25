# Bundling

Maybe relevant: [imports](./imports.md).

## Rad Script

If we have a conceptual script which is spread out across several files, this hurts shareability unless we build in some
functionality to rad to make managing this easier.

Let's say we have

```
- .
  - myscript.rsl
  - greet.rsl
  - admonish.rsl
```

where `myscript` is an entry script (see [commands](./commands.md#invoking-other-scripts)).

From here, we run `rad bundle --entry myscript.rsl --name myscript --output out`

The result:

```
- .
  - myscript.rsl
  - greet.rsl
  - admonish.rsl
  - out
    - myscript
    - myscript-lib
      - myscript.rsl
      - greet.rsl
      - admonish.rsl
```

Where `myscript` is a passthrough script to the entry script:

```
#!/usr/bin/env rad
---
@enable_args_block = 0
@enable_global_flags = 0
---

my_args = get_args()[1:].join(" ")
quiet $!`./myscript-lib/myscript {my_args}`
```

To share `myscript`, you'll then need to move `myscript` and `myscript-lib` together around.

That's not actually *great*, and you risk users misplacing them relative to each other.

Maybe it's possible to have a `myscript.radb` file i.e. 'rad bundle'.
This would really be a directory, or zip? idk how it'd work, we'd need to keep it very very performant for running,
we don't want to have to unzip every time it's run.

Rad would somehow need to know how to run the bundle, and if you put it on your path, you'd need to also have it be runnable (without writing `.radb` each time).

Maybe this is too finicky, still. Maybe we need to further abstract this away from users and allow rad's CLI to handle all the details of sharing, kinda like brew formulas.
You just deal with formulas, and brew takes care of installation, PATH, etc.

## Executable

Stubbed here, but this is referring specifically to bundling scripts such that they are a runnable executable that doesn't require rad to be installed on the running machine.

A big challenge here is that, if we want to keep rad interpreted, then this involves bundling in the Go runtime & GC, which makes 'executable scripts' >10 MB.
