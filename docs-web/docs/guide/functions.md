---
title: Functions
---

Rad offers a range of built-in functions to help you write your scripts, and also allows you to define your own.
In this section, we'll take a look at the syntax and a few examples.
For a complete list of built-in functions, see the [reference](../reference/functions.md).

## Syntax

The syntax for invoking functions is pretty standard. Here's a script with some examples:

```rad
names = ["Bob", "Charlie", "Alice"]
num_people = len(names)
print("There are {num_people} people.")

sorted_names = sort(names)
print(sorted_names)
```

<div class="result">
```
There are 3 people.
[ "Alice", "Bob", "Charlie" ]
```
</div>

This example uses three different built-in functions [`len`](../reference/functions.md#len), [`print`](../reference/functions.md#print), and [`sort`](../reference/functions.md#sort).

### UFCS

Rad supports a syntax called Uniform Function Call Syntax (UFCS) that lets you call functions using dot notation. This means you can write:

```rad
upper("hello")

// ... is the same as ...

"hello".upper()
```

Both styles work identically, it's just syntactic sugar, but the dot notation really shines when you're chaining multiple function calls together.

Compare these two approaches:

```rad
// Traditional nested calls - hard to read
result = upper(trim(text))

// UFCS chaining - reads left to right
result = text.trim().upper()
```

The chained version is much more readable - you can follow the data flow naturally from left to right.

UFCS works with any function where its first parameter matches the type you're calling it on.

!!! tip "Encouraged Style"

    You're encouraged to use UFCS, especially when it helps you avoid nested function calls.

## Function Arguments

Rad functions can accept arguments in several different ways. Let's explore each pattern.

### Positional Arguments

Most functions accept arguments by position - you pass values in a specific order. Here are a few examples:

```rad
// Single argument
num = abs(-5)
print(num)  // 5

// Multiple arguments
text = "hello world".replace("world", "Rad")
print(text)  // hello Rad
```

Many functions also have **optional parameters with defaults**. For example, [`join`](../reference/functions.md#join) combines list items into a string:

```rad
numbers = [1, 2, 3]

// Just the list - uses default separator ""
print(numbers.join())

// Custom separator
print(numbers.join("... "))

// Separator and prefix
print(numbers.join("... ", "Counting: "))

// Separator, prefix, and suffix
print(numbers.join("... ", "Counting: ", "!"))
```

<div class="result">
```
123
1... 2... 3
Counting: 1... 2... 3
Counting: 1... 2... 3!
```
</div>

The function signature for `join` shows these optional parameters: `join(list, sep="", prefix="", suffix="")`. You can provide as many or as few as you need.

!!! tip "Example using join for url query params"

    The `prefix` parameter is handy for generating URL query params:

    ```rad
    url = "https://api.github.com/repos/amterp/rad/commits"
    query_params = ["path=README.md", "per_page=5"]
    url += query_params.join("&", "?")
    print(url)
    ```

    This produces: [`https://api.github.com/repos/amterp/rad/commits?path=README.md&per_page=5`](https://api.github.com/repos/amterp/rad/commits?path=README.md&per_page=5)

### Named Arguments

Some functions accept **named arguments** that you pass using `name=value` syntax. Named arguments always come after positional arguments and are typically optional.

A good example is [`http_post`](../reference/functions.md#http-functions), which performs HTTP POST requests:

```rad
// Just the URL (simplest form)
response = url.http_post()

// With custom headers
my_headers = {
    "Authorization": "Bearer {token}",
}
response = url.http_post(headers=my_headers)

// With both headers and a body
response = url.http_post(headers=my_headers, body=data)
```

Named arguments make it clear what each value represents, especially when a function has many optional parameters.

### Variadic Arguments

Some functions accept **unlimited arguments**. For example, [`zip`](../reference/functions.md#zip) can combine any number of lists:

```rad
names = ["alice", "bob", "charlie"]
ages = [30, 40, 25]
cities = ["NYC", "LA", "Chicago"]
scores = [100, 90, 85]

// Combine 2 lists
pairs = zip(names, ages)
print(pairs)

// Combine 3 lists
triples = zip(names, ages, cities)
print(triples)

// Combine 4 lists (or more!)
quads = zip(names, ages, cities, scores)
print(quads)
```

<div class="result">
```
[ [ "alice", 30 ], [ "bob", 40 ], [ "charlie", 25 ] ]
[ [ "alice", 30, "NYC" ], [ "bob", 40, "LA" ], [ "charlie", 25, "Chicago" ] ]
[ [ "alice", 30, "NYC", 100 ], [ "bob", 40, "LA", 90 ], [ "charlie", 25, "Chicago", 85 ] ]
```
</div>

Variadic functions can also have named arguments. For example, `zip` accepts `strict=true` for ensuring that all lists have the same length.

### Mixed Patterns

Some functions combine multiple argument patterns. For example, [`pick`](../reference/functions.md#pick) takes positional arguments and a named argument:

```rad
options = ["vim", "emacs", "nano"]
editor = pick(options, prompt="Choose your editor")
```

When in doubt about how to call a function, check the [Functions Reference](../reference/functions.md) for complete signature details.

## Custom Functions

Rad lets you define your own functions using the `fn` keyword.
You can create either **named functions** that you reference by name, or **lambdas** (anonymous functions) that you assign to variables or pass as arguments.

### Named Functions

Named functions include the function name as part of the definition, making them easy to call from anywhere in your code.

#### Basic Definition

Here's a simple function that adds two numbers:

```rad
fn add(x, y):
    return x + y

result = add(5, 3)
print(result)  // 8
```

Functions use the `return` keyword to send values back. If your function body is a single expression, you can use a more concise syntax:

```rad
fn add(x, y) x + y

result = add(5, 3)
print(result)  // 8
```

Notice there's no colon (`:`) after the parameters in the single-line form, and no `return` keyword is needed.

#### Multiple Return Values

Functions can return multiple values at once using comma separation:

```rad
fn get_coords():
    x = 10
    y = 20
    return x, y  // equivalent to 'return [x, y]'

x_pos, y_pos = get_coords()
print("Position: ({x_pos}, {y_pos})")
```

<div class="result">
```
Position: (10, 20)
```
</div>

This uses **destructuring** (covered in [Basics](basics.md#destructuring)) to unpack the returned values into separate variables.

#### Type Annotations

You can optionally add type annotations to function parameters and return values:

```rad
fn calculate_area(width: int, height: int) -> int:
    return width * height

area = calculate_area(5, 10)
print(area)  // 50
```

There are three benefits to using these.

1. They serve as documentation (self-documenting code).
2. They are validated at runtime i.e. the above function will error early if a string is passed into `calculate_area`.
3. They help Rad's static analysis tools reason about your code, making them more useful.

They are covered in detail in a later section: [Type Annotations](./type-annotations.md).

#### Hoisting

Named functions have some special scoping rules worth knowing:

**At the root level, functions are hoisted** - you can call them before they're defined:

```rad
result = multiply(4, 5)
print(result)  // 20

fn multiply(a, b):
    return a * b
```

This works because Rad processes all root-level function definitions before executing the script.

**Inside blocks, functions are NOT hoisted** - you must define them before calling:

```rad
if true:
    print(helper())  // Error - can't call before definition

    fn helper():
        return "I'm a helper!"
    
    print(helper())  // This is okay!
```

### Lambdas

Sometimes you need a quick function without giving it a name. That's where **lambdas** come in - they're anonymous functions you can assign to variables or pass as arguments.

Lambdas use the same `fn` keyword, but without a name:

```rad
// Single-line lambdas
double = fn(x) x * 2
add = fn(x, y) x + y

print(double(5))  // 10
print(add(3, 4))  // 7
```

For multi-line logic, use the block style with a colon:

```rad
calculate = fn(x):
    result = x * 2 + 10
    return result

print(calculate(5))  // 20
```

[//]: # (TODO closures behave poorly, fix and write here i.e. they should statically capture values!)

Lambdas are particularly useful for defining once-off operations and passing them as arguments.
For example, with [`map`](../reference/functions.md#map) and [`filter`](../reference/functions.md#filter):

```rad
numbers = [1, 2, 3, 4, 5]
doubled = numbers.map(fn(x) x * 2)
print(doubled)  // [2, 4, 6, 8, 10]

evens = numbers.filter(fn(x) x % 2 == 0)
print(evens)  // [2, 4]
```

!!! tip "Named Functions vs Lambdas"

    - Use **named functions** (`fn add(x, y):`) for reusable logic that you'll call from multiple places
    - Use **lambdas** (`fn(x) x * 2`) for one-off operations or callbacks

## Reference

There are a lot of built-in functions. If you want to see what's available and how to use them, refer to the [reference](../reference/functions.md).

## Summary

- Function invocation syntax is similar to Python, Java, and other familiar languages
- **UFCS** (Uniform Function Call Syntax) lets you chain functions using dot notation: `text.trim().upper()`
- Functions can accept arguments in several ways:
    - **Positional**: passed by order, may have defaults (`join(list)`, `join(list, "|")`)
    - **Named**: passed with `name=value` syntax (`http_get(url, headers=my_headers)`)
    - **Variadic**: accept unlimited arguments (`zip(list1, list2, list3, ...)`)
    - **Mixed**: combinations of the above patterns
- You can define **named functions** with `fn name():` for reusable logic
    - Supports block style (with `:`) 
    - Supports single-line style (without `:` e.g. `fn add(x, y) x + y`)
    - Can return multiple values using comma separation
    - Are **hoisted** at the root of scripts (can be called before definition)
    - Can have optional **type annotations**
- **Lambdas** are anonymous functions: `double = fn(x) x * 2`
    - Useful for one-off operations and callbacks - functions like `map()` and `filter()`

## Next

We've already covered the [Basics of strings](./basics.md#str),
but there are some more advanced string concepts which are worth covering, such as formatting in string interpolations,
raw strings, etc.

We'll cover these in the next section: [Strings (Advanced)](./strings-advanced.md)
