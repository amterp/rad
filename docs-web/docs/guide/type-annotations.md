---
title: Type Annotations
---

As your scripts grow beyond quick one-offs, type annotations become increasingly valuable.
They help catch errors early, make your code self-documenting, and keep scripts maintainable as they evolve or get shared with others.

In the [Functions](./functions.md#type-annotations) section, we briefly introduced type annotations for function parameters and return values.
Now let's explore Rad's complete type system - from basic primitives to advanced types like unions, structs, and function signatures.

## The Basics

Type annotations let you declare what types of values your function parameters accept and what type of value your
function returns. The syntax follows a pattern you may recognize from TypeScript or Python's type hints:

```rad
fn calculate_area(width: int, height: int) -> int:
    return width * height

area = calculate_area(5, 10)
print(area)  // 50
```

Here, `width: int` and `height: int` specify that both parameters must be integers, and `-> int` declares that the function returns an integer.

### Why Use Type Annotations?

Type annotations provide three key benefits:

1. **Self-documenting code** - The function signature clearly communicates what types it expects and returns
2. **Runtime validation** - Rad checks types at runtime and produces helpful error messages when types don't match
3. **Tooling support** - IDEs and linters can provide better autocomplete and catch errors before you run your code

Let's see runtime validation in action:

```rad linenums="1"
fn greet(name: str) -> str:
    return "Hello, {name}!"

message = greet(42)  // Error!
print(message)
```

<div class="result">
```
Error at L4:17

message = greet(42)  // Error!
                ^^ Value '42' (int) is not compatible with expected type 'str'
```
</div>

The error message clearly identifies the problem - we passed an integer when the function expects a string.

### Basic Primitive Types

Rad supports the standard primitive types you'd expect: `str`, `int`, `float`, and `bool`

```rad
fn process_data(
    name: str,
    age: int,
    salary: float,
    is_active: bool
) -> str:
    status = is_active ? "active" : "inactive"
    return "{name} ({age}) earns ${salary:.2} - {status}"

result = process_data("Alice", 30, 75000.50, true)
print(result)
```

[//]: # (TODO ^^ should support trailing comma after is_active: bool)

<div class="result">
```
Alice (30) earns $75000.50 - active
```
</div>

### Special Types: void and null

Two additional types appear throughout Rad but work differently from the primitives:

**`void`**: Indicates a function returns nothing. Functions marked `-> void` don't return a value, and attempting to return a value is an error:

```rad
fn log_message(msg: str) -> void:
    print(msg)                 // OK

fn log_message(msg: str) -> void:
    return msg                 // Error: can't return values
```

**`null`**: The single value representing "no value" or "absence." Important: `null` is only a valid value for optional types marked with `?`:

```rad
fn get_name() -> str:
    return null          // Error: can't return null from str function

fn get_name() -> str?:
    return null          // OK: str? can return null
```

Think of `null` as belonging exclusively to optional types - it's the way to represent "this optional value is absent."

## Collection Types

Rad also lets you specify collection types, with their contents being either typed or untyped.

### Typed Lists

You can specify what type of values a list contains using the `<type>[]` syntax:

```rad
fn sum_numbers(nums: int[]) -> int:
    total = 0
    for num in nums:
        total += num
    return total

result = sum_numbers([1, 2, 3, 4, 5])
print(result)  // 15
```

The `int[]` annotation means "a list of integers". Like with other type annotations, if you try to pass a list
containing non-integers, you'll get a runtime error.

More examples with different types:

```rad
fn join_words(words: str[]) -> str:
    return words.join(" ")

fn average(numbers: float[]) -> float:
    return sum(numbers) / len(numbers)

sentence = join_words(["Hello", "from", "Rad"])
print(sentence)
```

<div class="result">
```
Hello from Rad
```
</div>

### Typed Maps

Maps can also be typed, specifying both key and value types using `{ <key type>: <value type> }` syntax:

```rad
fn count_words(text: str) -> { str: int }:
    words = text.split(" ")
    counts = {}
    for word in words:
        if word in counts:
            counts[word] += 1
        else:
            counts[word] = 1
    return counts

result = count_words("hello world hello")
print(result)
```

<div class="result">
```
{ "hello": 2, "world": 1 }
```
</div>

The `{ str: int }` annotation means "a map with string keys and integer values".

### Generic Collections

When you don't want to specify what's inside a collection, use the generic forms:

```rad
fn print_items(items: list) -> void:
    for item in items:
        print(item)

fn lookup(data: map, key: str) -> any:
    return data[key]
```

Here, `list` accepts a list with any types of values, and `map` accepts a map with any keys and values.

The `any` type means "any type of value" - it's the most permissive type and accepts strings, numbers, booleans, lists, maps, or any other value.

Generic collections are useful when your types are mixed or you don't wish to overcomplicate your type annotations unnecessarily.

### Nested Collections

Types can be nested for complex data structures:

```rad
fn organize_by_category(items: str[]) -> { str: str[] }:
    categories = {}
    for item in items:
        category = item[0].upper()  // First letter
        if category not in categories:
            categories[category] = []
        categories[category] += [item]
    return categories

items = ["apple", "banana", "apricot", "blueberry"]
result = organize_by_category(items)
print(result)
```

<div class="result">
```
{ "A": [ "apple", "apricot" ], "B": [ "banana", "blueberry" ] }
```
</div>

The return type `{ str: str[] }` describes a map where each key is a string and each value is a list of strings.

## Optional Types

Sometimes a parameter might not always be needed. Rad's optional type syntax with `?` makes parameters completely optional - you can pass a value, pass `null`, or omit the parameter entirely:

```rad
fn greet(name: str, title: str?) -> str:
    if title == null:
        return "Hello, {name}!"
    else:
        return "Hello, {title} {name}!"

print(greet("Alice", "Dr."))     // Pass a value
print(greet("Bob", null))        // Explicitly pass null
print(greet("Charlie"))          // Omit the parameter entirely
```

<div class="result">
```
Hello, Dr. Alice!
Hello, Bob!
Hello, Charlie!
```
</div>

The `str?` annotation means "an optional string parameter". When omitted or explicitly set to `null`, the parameter will be `null` inside the function. This makes it clear that the `title` parameter is optional and the function knows how to handle its absence.

Optional types work with any type:

```rad
fn find_user(id: int, users: map[]) -> map?:
    for user in users:
        if user["id"] == id:
            return user
    return null

users = [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]
user = find_user(1, users)
print(user)  // {"id": 1, "name": "Alice"}

missing = find_user(999, users)
print(missing)  // null
```

The `map?` return type indicates the function might return a map or might return null if no user is found.

## Defaults

Parameters can have default values, making them optional to provide when calling the function. This works whether or not the parameter is marked with `?`:

```rad
fn greet(name: str, greeting: str = "Hello") -> str:
    return "{greeting}, {name}!"

print(greet("Alice"))                // Uses default "Hello"
print(greet("Bob", "Hi"))            // Uses provided "Hi"
```

<div class="result">
```
Hello, Alice!
Hi, Bob!
```
</div>

The `greeting` parameter has a default value of `"Hello"`. When you omit it, the default is used. Note that `greeting` is not marked with `?` - it always has a string value, never `null`.

### Defaults & Optionals

When you combine defaults with optional types (`?`), you can choose whether the default should be `null` or something else:

```rad
fn format_price(amount: float, currency: str? = "USD") -> str:
    if currency == null:
        return "${amount:.2}"
    return "{amount:.2} {currency}"

print(format_price(19.99))           // Uses default "USD"
print(format_price(19.99, "EUR"))    // Uses provided "EUR"
print(format_price(19.99, null))     // Explicitly passes null
```

<div class="result">
```
19.99 USD
19.99 EUR
$19.99
```
</div>

With `str?` alone, omitting the parameter means it defaults to `null`. With `str? = "USD"`, you can provide a non-null default value, but callers can still explicitly pass `null` if they want.

## Union Types

Sometimes a function can accept or return multiple different types. Union types express this with the `|` operator:

```rad
fn to_string(val: int|float|str) -> str:
    return str(val)

print(to_string(42))
print(to_string(3.14))
print(to_string("hello"))
```

<div class="result">
```
42
3.14
hello
```
</div>

The `int|float|str` annotation means "accepts an int, float, or string" - any of these three types is valid.

### Error Union Types

A common union pattern in Rad is combining `error` with another type to represent operations that might fail:

```rad
fn divide(a: float, b: float) -> float|error:
    if b == 0:
        return error("Cannot divide by zero")
    return a / b

result = divide(10, 2)
print(result)  // 5
```

The `float|error` return type signals that this function returns either a float (on success) or an error value (on failure).

!!! note "Error Handling in Rad"

    Rad has a comprehensive error handling model. We'll cover error handling in detail in a later section: [Error Handling](./error-handling.md).

## Advanced Types

Rad's type system includes several advanced features for expressing complex data structures and constraints.

### Enum Types

When a value should be restricted to specific strings, use enum types:

```rad
fn set_log_level(level: ["debug", "info", "warn", "error"]) -> str:
    return "Log level set to: {level}"

print(set_log_level("info"))
// set_log_level("trace")  // Error: "trace" not in enum
```

<div class="result">
```
Log level set to: info
```
</div>

The `["debug", "info", "warn", "error"]` annotation restricts the parameter to exactly these four string values. Any other string will cause a runtime type error.

This is particularly useful for configuration options, status values, and other cases where only certain strings are valid:

```rad
fn create_connection(
    host: str,
    protocol: ["http", "https", "ws", "wss"] = "https"
) -> str:
    return "{protocol}://{host}"

url = create_connection("api.example.com")
print(url)
```

<div class="result">
```
https://api.example.com
```
</div>

### Structured Maps

For maps with specific named fields, use the struct syntax with quoted keys:

```rad
fn create_user(name: str, age: int, email: str) ->
        { "name": str, "age": int, "email": str, "id": int }:
    return {
        "name": name,
        "age": age,
        "email": email,
        "id": rand_int(1000, 9999)
    }

user = create_user("Alice", 30, "alice@example.com")
print(user)
```

<div class="result">
```
{ "name": "Alice", "age": 30, "email": "alice@example.com", "id": 7234 }
```
</div>

The `{ "name": str, "age": int, "email": str, "id": int }` annotation describes a map with exactly these four fields, each with a specific type. Notice the quoted keys - this distinguishes named fields from the typed map syntax we saw earlier.

#### Optional Fields in Structs

Fields can be marked as optional with `?`:

```rad
fn get_user_profile(id: int) ->
        { "name": str, "age": int, "bio"?: str, "avatar"?: str }:
    // Fetch user... in this example, we'll return mock data
    return {
        "name": "Bob",
        "age": 25,
        "bio": "Software engineer"
        // avatar field is omitted
    }

profile = get_user_profile(123)
print(profile)
```

<div class="result">
```
{ "name": "Bob", "age": 25, "bio": "Software engineer" }
```
</div>

The `"bio"?: str` and `"avatar"?: str` fields are optional - the map might or might not contain them.

#### Nested Structures

Struct types can be nested for complex data:

```rad
fn fetch_article() -> {
    "title": str,
    "author": { "name": str, "id": int },
    "metadata": { "views": int, "likes": int },
}:
    return {
        "title": "Getting Started with Rad",
        "author": {"name": "Alice", "id": 1},
        "metadata": {"views": 1234, "likes": 56}
    }

article = fetch_article()
print("Article: {article.title} by {article.author.name}")
print("Stats: {article.metadata.views} views, {article.metadata.likes} likes")
```

<div class="result">
```
Article: Getting Started with Rad by Alice
Stats: 1234 views, 56 likes
```
</div>

### Function Types

Functions themselves can be typed, which is especially useful when passing functions as parameters:

```rad
fn apply_to_list(items: str[], transform: fn(str) -> str) -> str[]:
    result = []
    for item in items:
        result += [transform(item)]
    return result

words = ["hello", "world"]
upper_words = apply_to_list(words, upper)
print(upper_words)
```

<div class="result">
```
[ "HELLO", "WORLD" ]
```
</div>

The `fn(str) -> str` annotation describes a function that takes a string parameter and returns a string.

Other examples of valid function type annotations:

```
fn() -> int
fn(str, str) -> str
fn(str[]) -> void
```

## Variadic and Named Parameters

Type annotations work seamlessly with Rad's parameter patterns, as seen earlier in [Functions](./functions.md#function-arguments).

### Variadic Parameters

When a function accepts unlimited arguments, you can type the variadic parameter:

```rad
fn sum_all(*numbers: int) -> int:
    total = 0
    for num in numbers:
        total += num
    return total

result = sum_all(1, 2, 3, 4, 5)
print(result)
```

<div class="result">
```
15
```
</div>

The `*numbers: int` annotation means "zero or more integer arguments". All arguments passed to this variadic parameter must be integers.

### Named-Only Parameters

Named-only parameters (those after `*`) can also be typed:

```rad
fn format_text(
    text: str,
    *,
    uppercase: bool = false,
    prefix: str = "",
    suffix: str = ""
) -> str:
    result = prefix + text + suffix
    return uppercase ? upper(result) : result

output = format_text("hello", uppercase=true, prefix=">>> ")
print(output)
```

<div class="result">
```
>>> HELLO
```
</div>

### Combining Everything

Here's a function that combines positional, variadic, and named-only parameters with types:

```rad
fn create_report(
    title: str,
    *data_points: int|float,
    *,
    format: ["text", "html", "json"] = "text",
    include_summary: bool = true
) -> str:
    total = sum(data_points)
    avg = total / len(data_points)

    report = "=== {title} ===\n"
    report += "Data: {data_points.join(', ')}\n"

    if include_summary:
        report += "Total: {total}, Average: {avg:.2}"

    return report

output = create_report(
    "Q4 Sales",
    100, 150, 200, 175,
    format="text",
    include_summary=true
)
print(output)
```

<div class="result">
```
=== Q4 Sales ===
Data: 100, 150, 200, 175
Total: 625, Average: 156.25
```
</div>

This example demonstrates:

- A required positional parameter (`title: str`)
- A typed variadic parameter accepting multiple numeric values (`*data_points: int|float`)
- Named-only parameters with enum and boolean types
- A clear, self-documenting function signature

## Summary

Type annotations are an optional but powerful tool for keeping your Rad scripts maintainable and self-documenting,
especially as they grow in complexity or get reused across projects.

**Key takeaways:**

- **Syntax**: Parameter types use `param: <type>`, return types use `-> <type>`
- **Benefits**: Self-documenting code, runtime validation, better tooling support
- **Primitive types**: `str`, `int`, `float`, `bool` for basic values
- **Special types**: `void` (function returns nothing), `null` (only valid for optional types marked with `?`)
- **Collection types**:
    - `T[]` for typed lists (e.g., `str[]`, `int[]`, `float[]`)
    - `{ <key type>: <value type> }` for typed maps (e.g., `{ str: int }`)
    - `list` and `map` for generic collections (any contents)
    - Nested collections like `int[][]` and `{ str: str[] }`
- **Optional types**: `T?` for nullable values (e.g., `str?`, `int?`)
- **Union types**: `T|U` for multiple acceptable types (e.g., `int|float`, `str|list`)
- **Advanced types**:
    - **Enums**: `["value1", "value2", "value3"]` for restricted string values
    - **Structs**: `{ "field1": type1, "field2"?: type2 }` for structured maps with named fields (quoted keys) and optional fields
    - **Function types**: `fn(<param_type>) -> <return_type>` for function parameters and variables
    - **Nested structures**: Complex combinations of the above
- **Special parameters**: Work with variadic (`*param: <type>`) and named-only parameters

Type annotations make your code clearer to both humans and tools, catching errors early and making your intentions explicit.

## Next

We've briefly seen `error|T` union types in this section - functions that return either a value or an error.

In the next section, we'll explore Rad's comprehensive error handling model in depth: [Error Handling](./error-handling.md).
