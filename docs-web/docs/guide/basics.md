---
title: Basics
---

This section of the guide will rapidly cover the basics of RSL.
RSL shares a lot of conventions and syntax with popular languages like Python,
so if you're familiar with programming, this will be pretty straightforward.

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

Strings can be delimited in three ways:

1. Double quotes: `"text"`
2. Single quotes: `'text'`
3. Backticks: ``` `text` ```

All three behave the same way. To demonstrate:

```rsl
greeting = "Hello!"
print(greeting)

greeting = 'Hello!'
print(greeting)

greeting = `Hello!`
print(greeting)
```

<div class="result">
```
Hello!
Hello!
Hello!
```
</div>

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

Strings can include special characters such as `\n` for new lines and `\t` for tabs.

```rsl
print("Hello\tdear\nreader!")
```

<div class="result">
```
Hello	dear
reader!
```
</div>

RSL also supports **raw strings**. Raw strings do not allow any escaping (including the delimiter used to create them).
Use them when you want your contents to remain as "raw" and unprocessed as possible.

To use them, just prefix the delimiter of your choice (single/double quotes or backticks) with `r`.

```rsl
text = r"Hello\tdear\nreader!"
print(text)
```

<div class="result">
```
Hello\tdear\nreader!
```
</div>

Notice the printed string is exactly as written in code - neither the tab nor newline character were replaced.

Again, you could use any of the three delimiters for raw strings:

```rsl
text = r"Hello\tdear\nreader!"
text = r'Hello\tdear\nreader!'
text = r`Hello\tdear\nreader!`
```

!!! tip "Raw strings for file paths"

    Raw strings can be quite handy for file paths, especially Windows-style ones that use backslashes:

    ```rsl
    path = r"C:\Users\Documents\notes.txt"
    ```

[//]: # (todo MULTILINE STRINGS)

!!! info "You cannot escape the raw string's own delimiter"

    RSL raw strings behave more like their equivalent in Go than Python.
    In Python, you can escape the delimiter used to make the raw string i.e. `r"quote: \"!"`. If printed, this will
    display as `quote: \"!` i.e. the escape character is also printed. There are lots of discussions online about this
    somewhat odd behavior, which is why RSL (and Go) opted to instead keep the rules very simple and not allow escaping
    in raw strings of any kind.
    
    Instead, if you try the same thing in RSL, you will get an error because the quote following `\` will close the
    string, leaving a dangling `!"` at the end, which is invalid syntax.

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
mixed_map = { 
  "alice": "accountant",
  "mylist": [ "London", 25 ],
}

nested_map = {
  "error": {
    "msg": "Request failed!",
    "code": 400,
  }
}
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

RSL offers operators similar to many other languages. Below sections very quickly demonstrate.

### Arithmetic

RSL follows the standard order of operations for operators `() , + , - , * , / , %`:

1. Parentheses
2. Multiplication, Division, Modulo
3. Addition, Subtraction

```rsl
print(1 + 4 / 2)    // 3
print(2.5 * 3 - 1)  // 6.5
print((4 + 5) * 2)  // 18
print(5 % 3)        // 2
```

Dividing two integers will result in a floating point number.

```rsl
print(5 / 2)  // 2.5
```

### Comparison

Comparisons return bools that can be used in e.g. [if statements](#if-statements).

String comparison is done based on contents.

```rsl
print("alice" == "alice")  // true
print("alice" == "bob")    // false
print("alice" != "bob")    // true
print("alice" == "Alice")  // false
```

Numbers can also be compared with the standard comparators `> , >= , < , <= , ==`.

```rsl
print(2 >= 2)  // true
print(2 > 2)   // false
print(2 <= 2)  // true
print(2 < 2)   // false
print(2 == 2)  // false
```

You cannot use these operators (outside of `==`) to compare non-numbers such as strings:

```rsl
print("alice" > "bob")  // error
```

But you *can* check them for equality (will always return false, except ints and floats that are equal):

```rsl
print(2 == "alice")  // false
print(2 == 2.0)      // true
```

[//]: # (todo when we add collection equality, document here)

!!! info "Difference From Python on `True == 1` and `False == 0`"

    In Python, `False == 0` and `True == 1` are true, because under the hood, False is really int 0 and True is really int 1,
    hence they're equal. That's not the case in RSL. In RSL, **any two values of different types are not equal**.

    The reasoning stems from [truthy/falsy-ness](#truthyfalsy). In Python, both `1` and `2` are truthy. But only `1` equals `True`.
    RSL avoids this oddity of making `1` special by instead making any two different types not equal (except ints/floats).

[//]: # (todo move this ^ note to the reference section? is it really necessary in this basics section?)

### Logical

RSL uses `and` and `or` for binary logical operations.

```rsl
print(false and false)  // false
print(false and true)   // false
print(true  and false)  // false
print(true  and true)   // true

print(false or  false)  // false
print(true  or  false)  // true
print(false or true)    // true
print(true  or  true)   // true
```

And it uses `not` for logical negation.

```rsl
print(not true)   // false
print(not false)  // true
```

### Concatenation

You can concatenate strings with `+`. 

```rsl
first = "Alice"
last = "Bobson"
print(first + last)
```

<div class="result">
```
Alice Bobson
```
</div>

You cannot concatenate a string and a non-string. First convert the non-string into a string.

This can be done in several ways, the easiest is probably via [string interpolation](../reference/strings.md#string-interpolation):

[//]: # (todo that might change after str func gets added)

```rsl
a = 5
text = "Number: "
print(text + "${a}")
```

<div class="result">
```
Number: 5
```
</div>

### Compound Operators

```rsl
a = 3
a += 2   // a is now 5
a -= 1   // a is now 4
a *= 3   // a is now 12
a %= 10  // a is now 2
a /= 4   // a is now 0.5
```

RSL does not support `++` or `--` syntax.

[//]: # (todo that might change...)

### Ternary

RSL supports `? :` style ternary operators. 

`<condition> ? <true case> : <false case>`

```rsl
a = 5
b = a > 0 ? "larger than 0" : "less than 0"
print(b)
```

<div class="result">
```
larger than 0
```
</div>

## Truthy/Falsy

TBC

## Control Flow

### If Statements

- TBC
    - if
        - truthy/falsy 
    - for
    - switch

## Converting Types

- TBC
  - parsing
  - casting (once implemented)
