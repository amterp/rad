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

RSL has 6 data types: strings, ints, floats, bools, lists, and maps. 

### string

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
    [shell commands](../guide/shell-commands.md), as shell/bash commands may include single or double quotes, and backticks
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

Strings also support **interpolation**. String interpolation allows you to write expressions *inside your string* that will be evaluated and replaced for the final string. We'll cover this in more depth in a future section, but to give a very simple example:

```rsl
name = "Alice"
print("Hi, {name}!")
```

<div class="result">
```
Hi, Alice!
```
</div>

Anything encapsulated in a `{}` gets treated as an expression. Here, the expression is just the identifier `name`, which gets evaluated and substituted, giving us the final `Hi, Alice!` string.

Those are the basics for strings - we'll cover additional string concepts in a future section, [Strings (Advanced)](./strings-advanced.md).

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

    The reasoning stems from [truthy/falsy-ness](#truthy-falsy). In Python, both `1` and `2` are truthy. But only `1` equals `True`.
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
print(text + "{a}")
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

### Increment / Decrement

You can quickly increment and decrement ints and floats using `++` and `--` syntax.

```rsl
a = 2
a++
print(a)

b = 2.5
b--
print(b)
```

<div class="result">
```
3
1.5
```
</div>

The increment and decrement operators produce *statements*, not *expressions*.
This means that `a++` does not *return* anything, and so cannot be used inside e.g. a conditional.

For example, the following two uses are invalid, because `a++` doesn't return a value:

```rsl
a = 5
if a++ > 0:  // invalid, nothing for > to evaluate against on the left side
  ...
  
b = a++  // also invalid because a++ doesn't return any value
```

Because of that, there's also no reason to support *pre*-incrementing, and so `++a` or `--a` are invalid statements.

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

## Control Flow

### If Statements

RSL employs very standard if statements.

You are **not** required to wrap conditions in parentheses `()`.

```rsl
if units == "metric":
  print("That's 10 meters.")
else if units == "imperial":
  print("That's almost 33 feet.")
else:
  print("I don't know what measurement system!")
```

### For Loops

RSL allows "for each" loops for iterating through collections such as lists.

```rsl
names = ["Alice", "Bob", "Charlie"]
for name in names:
  print("Hi {name}!")
```

<div class="result">
```
Hi Alice!
Hi Bob!
Hi Charlie!
```
</div>

You can also iterate through a range of numbers using the [`range`](../reference/functions.md#range) function, which returns a list of numbers within some specified range.

```rsl
for i in range(5):
  print(i)
```

<div class="result">
```
0
1
2
3
4
```
</div>

You can also invoke `range` with a starting value i.e. `range(start, end)` and with a step value i.e. `range(start, end, step)`.

If you want to iterate through a list while also having a variable for the item's index, you can do that by adding
in an additional variable after the `for`. The first variable will be the index, and the second the item.

```rsl
names = ["Alice", "Bob", "Charlie"]
for i, name in names:
  print("{name} is at index {i}")
```

<div class="result">
```
Alice is at index 0
Bob is at index 1
Charlie is at index 2
```
</div>

When iterating through a map, if you have one variable in the loop, then that variable will be the key:

```rsl
colors = { "alice": "blue", "bob": "green" }
for k in colors:
  print(k)
```

<div class="result">
```
alice
bob
```
</div>

If you have two, then the first will be the key, and the second will be the value.

```rsl
colors = { "alice": "blue", "bob": "green" }
for k, v in colors:
  print(k, v)
```

<div class="result">
```
alice blue
bob green
```
</div>

[//]: # (todo i, k, v when implemented)

- TBC
  - switch

## Truthy / Falsy

RSL supports truthy/falsy logic.

For those unfamiliar, this means that, instead of writing the following (as an example):

```rsl
if len(my_list) > 0:
    print("My list has elements!")
```

you can write

```rsl
if my_list:
    print("My list has elements!")
```

Essentially, you can use any type as a condition, and it will resolve to true or false depending on the value.

The following table shows which values return false for each type. **All other values resolve to true.**

| Type   | Falsy | Description   |
|--------|-------|---------------|
| string | `""`  | Empty strings |
| int    | `0`   | Zero          |
| float  | `0.0` | Zero          |
| list   | `[]`  | Empty lists   |
| map    | `{}`  | Empty maps    |

!!! note ""

    Note that a string which is all whitespace e.g. `" "` is truthy.

## Converting Types

- TBC
  - parsing
  - casting (once implemented)

## Summary

- We rapidly covered many basic topics such as assignment, data types, operators, and control flow.
- RSL has 6 data types that map from JSON: strings, ints, floats, bools, lists, and maps.
- RSL has operators such as `+ , - , * , / , %`. For bool logic, it uses `or` and `and`.
- RSL uses a "for-each" variety `for` loop. You always loop through items in a collection (or string).
    - If you want to increment through a number range, use the `range` function to generate you a list of ints.
- RSL offers truthy/falsy logic for more concise conditional expressions.

## Next

Good job on getting through the basics of the language! 

Next, let's dive into one of the areas of RSL that make it shine: [Args](./args.md).
