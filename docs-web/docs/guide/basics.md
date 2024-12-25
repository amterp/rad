---
title: Basics
---

This section of the guide will rapidly cover the basics of RSL. RSL shares a lot of conventions and syntax with popular languages like Python, so if you're familiar with programming, this will be quick & easy.

## Variables & Assignment

To create a variable, you can do it through assignment. Let's use a string example:

```rsl
name = "Alice"
```

You can re-assign variables. Types don't need to stay the same:

```rsl
name = 2
```

!!! warning "You **cannot create** multiple variables this way on one line."

    The following is illegal syntax

    ```rsl
    a, b = "one", "two"
    ```

    instead, declare each variable on one line.

    ```rsl
    a = "one"
    b = "two"
    ```

## Data Types

[//]: # (todo for number types, should we advertise precision/bits?)

RSL's data types closely mirror those of JSON. Namely: strings, ints, floats, bools, lists, and maps.

### Strings

Strings can be delimited in three ways. The most standard are double quotes (`"text"`) or single quotes(`'text'`). The third is backticks (``` `text` ```).

Double and single quotes behave the same way. For example:

```rsl
greeting = "Hello\nWorld!"
print(greeting)
```

<div class="result">
```
Hello
World!
```
</div>

```rsl
greeting = 'Hello\nWorld!'
print(greeting)
```

<div class="result">
```
Hello
World!
```
</div>

Backtick-delimited strings behave a little differently. Characters like `\n` don't get escaped and so print as-is.

```rsl
greeting = `Hello\nWorld!`
print(greeting)
```

<div class="result">
```
Hello\nWorld!
```
</div>

Use backtick strings when you want the contents to remain closer to their 'raw' form.

!!! tip "Why 3 different delimiters?"

    Having 3 different delimiters is particularly useful when you want your string to contain one (or more) of the delimiter characters.

    For example, if you want a double quote in your string, you *can* use double quote delimiters and escape them:

    ```rsl
    "She said \"Goodbye\""
    ```

    However, this can be finicky and hard to read. Instead, you can pick one of the other two delimiters, for example:

    ```rsl
    'She said "Goodbye"'
    `She said "Goodbye"`
    ```
    
    We'll cover this again later, but as a note, backticks can be particularly useful in
    [shell commands](../guide/shell_commands.md), as shell/bash commands may include single or double quotes, and backticks
    save us from having to escape them.

### int

RSL has ints. There's nothing unusual about them. Example:

```rsl
team_size = 20
celsius = -5
```

Note that if you divide two ints, you will get back a [float](#float).

```rsl
liters = 10
people = 4
print("This is a float:", liters / people)
```

<div class="result">
```
This is a float: 2.5
```
</div>

### float

The other number type is float:

```rsl
length_meters = 2.68
```

If you want to define a whole number as a float, simply include a decimal place:

```rsl
years = 20.0
```

### bool

RSL uses lowercase `true` / `false`:

```rsl
is_running = true
is_tired = false
```

### list

RSL has two collection types: lists and maps. First, let's look at lists. These are also sometimes referred to as 'arrays' in other languages.

```rsl
names = ["alice", "bob", "charlie"]
```

Lists you define can contain any types:

```rsl
mixed = ["alice", true, 50, -2.4]
```

They can also be nested:

```rsl
nested = ["alice", [1, ["very nested", "ahhh"]]]
```

Indexing and slicing works very similarly to Python. If we assume the 3 variables above exist, you can index with both positive and negative indexes:

```rsl
print(names[0])
print(mixed[-1])  // grab last element in the list
print(nested[1][1][0])
```

<div class="result">
```
alice
-2.4
very nested
```
</div>

And also you can slice:

```rsl
numbers = [10, 20, 30, 40, 50]
print(numbers[1:3])
print(numbers[2:])
print(numbers[:-1])
```

<div class="result">
```
[20, 30]
[30, 40, 50]
[10, 20, 30, 40]
```
</div>

[//]: # (todo cover modifying indices!)
[//]: # (todo maybe this is assuming too much knowledge)

### map

The last type, and second of two collection types, is a 'map'. These may also be referred to as 'hashmap' or 'dictionary' in other languages.

```rsl
scores = { "alice": 25, "bob": 17, "charlie": 36 }
```

Like lists, they can contain mixed types for values, and can nest. However, **keys must be strings.**

```rsl
mixed_map = { "alice": "accountant", "mylist": ["London", 25] }
nested_map = { "error": { "msg": "Request failed!", "code": 400 } }
```

If we take the above example, values can then be accessed in two ways. First is the square bracket lookup:

```rsl
print(mixed_map["alice"])
print(nested_map["error"]["msg"])
```

<div class="result">
```
accountant
Request failed!
```
</div>

Alternatively, you can use a dot syntax. Note this second way only works for keys with no spaces in the name.

```rsl
print(mixed_map.alice)
print(nested_map.error.msg)
```

<div class="result">
```
accountant
Request failed!
```
</div>

You can modify maps using either syntax:

```rsl title="Using brackets"
mymap = { "alice": 30 }

mymap["alice"] = 40
print(mymap)

mymap.alice = 50
print(mymap)
```

<div class="result">
```
{ alice: 40 }
{ alice: 50 }
```
</div>

You can also add keys this way:

```rsl
mymap = { "alice": 30 }
mymap["bob"] = 31
mymap.charlie = 32
print(mymap)
```

<div class="result">
```
{ alice: 30, bob: 31, charlie: 32 }
```
</div>

## Operators

- TBC
    - arithmetic
    - comparison
    - logical
    - concat
    - ternary

## Control Flow

- TBC
    - if
        - truthy/falsy 
    - for
    - switch
