---
title: Args
---

This section covers syntax for defining arguments that your script can accept.

## Arg Block

RSL takes a declarative approach to arguments.
You simply declare what arguments your script accepts, and let RSL take care of the rest, including parsing user input.

Arguments are declared as part of an **args block**.

Here's an example for a script which prints an input word N number of times:

```rsl
args:
    word string
    repeats int
    
for _ in range(repeats):
    print(word)
```

This script defines two mandatory arguments: `word` that is expected to be a string, and `repeats` which is expected to be an integer.

RSL will take care of generating the help string and ensuring correct input is provided to the script when invoked.
For example, it will reject 


[//]: # (- todo)
[//]: # (  - all args are positional)
[//]: # (  - all also have flags)
