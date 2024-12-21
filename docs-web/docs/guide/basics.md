---
title: Basics
---

This section of the guide will rapidly cover the basics of RSL. RSL shares a lot of conventions and syntax with popular languages like Python, so if you're familiar with programming, this will be quick & easy.

## Variables & Assignment

To create a variable, you can do it through assignment. Let's use a string example:

```rsl
name = "Alice"
```

You can re-assign variables at any time:

```rsl
name = "bob"
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

### Strings

Strings can be delimited in three ways. The most standard are double quotes (`"text"`) or single quotes(`'text'`). The third is backticks (``` `text` ```).

Single and double quotes behave the same way. For example:

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

### int

### float

### bool

### list

### map

- TBC
    - incl dot.syntax 

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
