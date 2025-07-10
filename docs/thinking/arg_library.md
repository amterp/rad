# Arg Library

I've hit quite a few frustrations during my time developing rad as I've used cobra, and then just pflag, and hand-rolling my own arg parsing for parts as well. 

## 2025-07-10

What spawns this particular thought is wanting var arg behavior, but for flags.

Unix/Posix standards mandate 0 or 1 argument associated with each flag. 2 or more is not allowed.

Posix-compliancy is nice, but honestly not a requirement I want to enforce from Rad. That's on users. And sometimes, I want to write scripts which are not Posix compliant, because the ergonomics of the alternative are just that appealing.

For example

```
mycmd --foo aaa bbb ccc --bar ddd
```

Here, I want `foo` to be `[aaa, bbb, ccc]`, and `bar` to be `[ddd]` (assuming they're both var args).

One thought is to actually use Rad syntax to define how you want the flags parsed, and let Rad power the args engine.

Probably more realistically, we move Rad to entirely using our new arg framework, including teaching the framework to do our enum constraints, regex constraints, etc.

Some things to handle

- various types of positional args
- for positional args, vararg must be last
- var arg flags
- the `--` ending convention to end all flags and now interpret only as positional args
- negative ints
- constraints e.g. relational, range, regex, enum, etc
- int shorthand flags can be repeated to increment number e.g. `-vvv` makes `v == 3`
- optionals/defaults
- interactive mode -- missing flags are asked for
- usage string generation
- joined shorthands possible e.g. `-eva`
  - how combine if they're not all bools? following arg goes to the *last* flag?
- useful errors
- inspect pflag -- what does it offer we need to offer too?

Names ideas:

- amterp/rad-args
- amterp/rad-arglib
