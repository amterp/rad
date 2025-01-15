---
title: Functions
---

RSL offers a range of built-in functions to help you write your scripts. In this section, we'll take a look at the syntax and a few examples.
For a complete list of available functions, see the [reference](../reference/functions.md).

## Syntax

The syntax for invoking functions is pretty standard. Here's a script with some examples:

```rsl
names = ["Bob", "Charlie", "Alice"]
num_people = len(names)
print("There are {num_people} people.")

sorted_names = sort(names)
print(sorted_names)
```

<div class="result">
```
There are 3 people.
[Alice, Bob, Charlie]
```
</div>

This example uses a few different functions:

- `len`
- `print`
- `sort`

## No Arguments

Some functions take no arguments. For example, `rand()` returns a random float between 0 and 1:

```rsl
random_float = rand()
print(random_float)
```

<div class="result">
```
0.8436881320514183
```
</div>

## Fixed Arguments

Some functions take a fixed number of arguments, such as `upper` and `lower` that always take one argument:

```rsl
text = "oh WOW!"
print(upper(text))
print(lower(text))
```

<div class="result">
```
OH WOW!
oh wow!
```
</div>

## Variadic Arguments

Some functions can take different numbers of arguments! For example `join`:

```rsl
numbers = [1, 2, 3]
print(join(numbers, "... "))

print(join(numbers, "... ", "Okay I'll count. "))

print(join(numbers, "... ", "Okay I'll count. ", "!"))
```

<div class="result">
```
1... 2... 3
Okay I'll count. 1... 2... 3
Okay I'll count. 1... 2... 3!
```
</div>

In this example, `join` is being invoked with all these valid variations:

- `join(list, joiner)`
- `join(list, joiner, prefix)`
- `join(list, joiner, prefix, suffix)`

!!! tip "Example using join for url query params"

    The second variation of `join` can be handy for generating the query params in a url. For example:

    ```rsl
    url = "https://api.github.com/repos/amterp/rad/commits"
    query_params = ["per_page=5", "path=README.md"]
    url += join(query_params, "&", "?")
    print(url)
    ```

    In this example, the final url will be the following valid URL utilizing those query params:
    [`https://api.github.com/repos/amterp/rad/commits?path=README.md&per_page=5`](https://api.github.com/repos/amterp/rad/commits?path=README.md&per_page=5)

## Named Arguments

Finally, some functions may also have named arguments.
An example of this is `http_post`. `http_post` (unsurprisingly) performs an HTTP POST request against an input url, usually with a body of some sort.
One variation of the function only takes that url:

```rsl
response = http_post(url, body)
```

However, if you wish to customize the headers on your HTTP request, you can do so:

```rsl
my_headers = {
    "Authorization": "Bearer {token}",
}
response = http_post(url, body, headers=my_headers)
```

In this example, the named arg `headers` expects a map. Named args are always optional. Required arguments cannot be specified as required arguments.

[//]: # (todo might be nice to be able to specify e.g. join suffix without needing to specify prefix?)

## Reference

There are a lot of built-in functions. If you just want to see what's available and how to use them, it's best to refer to the [reference](../reference/functions.md).

!!! note "RSL does not allow defining your own functions"

    RSL currently does not allow you to define your own functions, so any function you use is built-in.

## Learnings Summary

- Function invocation syntax is similar to most other C-like languages such as Python, Java, etc.
- Functions may take no arguments, a fixed number of arguments, a varying number of args, and/or named arguments.
- RSL does not allow you to define your own functions - all functions are built-in.

## Next

We've already covered the [Basics of strings](./basics.md#string),
but there are some more advanced string concepts which are worth covering, such as string interpolation, raw strings, etc.
We'll cover these in the next section: [Strings (Advanced)](./strings-advanced.md)!
