# Rad Language Syntax Reference

This document provides a comprehensive overview of Rad's syntax for quick reference in development sessions.

## Script Structure

### Basic Script Format

```rad
#!/usr/bin/env rad
---
Script description goes here.
Multi-line descriptions are supported.
---

// Optional arguments section
args:
    name str              # Required string argument
    count int = 5         # Optional with default value
    verbose v bool        # Boolean flag (can use short form)

// Script body
print("Hello {name}!")
```

### File Header for Automatic Help Generation

The file header automatically generates the help text shown when users run your script with `-h` or `--help`. This is a core Rad feature that eliminates the need to manually write help documentation.

```rad
---
Script description that appears in --help.

Detailed explanation of what the script does.
Can be multiple paragraphs.

@stash_id = my_script_data
@enable_global_options = true  
@enable_args_block = true
---
```

#### Special @ Macros

- `@stash_id` - Sets the stash identifier for the script
- `@enable_global_options` - Enable/disable global Rad options (default: true)
- `@enable_args_block` - Enable/disable argument parsing (default: true)

## Comments

```rad
// Code comments use double slash
// Multi-line code comments use multiple // lines
// like this

// In args blocks, argument descriptions use # for help text generation:
args:
    name str    # This comment appears in --help usage
    count int   # This also appears in --help
```

## Data Types

### Primitives

```rad
// Strings
name = "alice"
path = 'path/to/file'

// Numbers
age = 25                   // int
height = 5.9               // float
scientific = 1.23e4        // exponential notation
large_num = 1_234_567      // underscore separators for readability
float_with_sep = 123.456_789

// Booleans
is_valid = true
is_empty = false

// Null
value = null
```

### Collections

```rad
// Lists
numbers = [1, 2, 3, 4, 5]
mixed = [1, "hello", true, [1, 2]]
empty_list = []

// Maps/Objects
person = {"name": "alice", "age": 25}
nested = {"user": {"name": "bob", "roles": ["admin", "user"]}}
empty_map = {}
```

## Variables and Assignment

### Basic Assignment

```rad
name = "alice"
age = 25
```

### Multiple Assignment

```rad
a, b = 1, 2
x, y = some_function_returning_tuple()
```

### Compound Assignment

```rad
count = 5
count += 1              // count = 6
count -= 2              // count = 4
count *= 3              // count = 12
count /= 4              // count = 3
```

### Increment/Decrement

```rad
i = 0
i++                     // i = 1
i--                     // i = 0
```

## String Interpolation

```rad
name = "alice"
age = 25
message = "Hello {name}, you are {age} years old!"

// Works with expressions
total = "Result: {x + y}"

// Format specifiers
price = 123.456
formatted = "Price: {price:.2}"          // "Price: 123.46" (2 decimal places)
padded = "Name: {name:<10}"              // "Name: alice     " (left-aligned, padded to 10)
right_aligned = "Name: {name:>10}"       // "Name:      alice" (right-aligned)
```

## Ternary Operator

```rad
// Basic ternary syntax
result = condition ? value_if_true : value_if_false

// Examples
status = age >= 18 ? "adult" : "minor"
message = count == 1 ? "1 item" : "{count} items"
max_value = a > b ? a : b

// Can be chained
category = age < 13 ? "child" : age < 18 ? "teen" : "adult"
```

## Collection Access and Slicing

### Indexing

```rad
items = [10, 20, 30, 40, 50]
first = items[0]        // 10
last = items[-1]        // 50

person = {"name": "alice", "age": 25}
name = person["name"]   // "alice"
```

### Slicing

```rad
items = [10, 20, 30, 40, 50]
subset = items[1:3]     // [20, 30]
from_start = items[:3]  // [10, 20, 30]
to_end = items[2:]      // [30, 40, 50]
all_items = items[:]    // [10, 20, 30, 40, 50]

// Negative indices
end_items = items[-2:]  // [40, 50]
```

### String Slicing

```rad
text = "hello"
substring = text[1:4]   // "ell"
```

## Control Flow

### If Statements

```rad
if age >= 18:
    print("Adult")
else if age >= 13:
    print("Teen")
else:
    print("Child")

// Single condition
if is_valid:
    process_data()
```

### Switch Statements

```rad
// Expression switches (single values)
result = switch value:
    case "a" -> "Apple"
    case "b" -> "Banana"
    default -> "Unknown"

// Multiple cases
status = switch code:
    case 200, 201, 204 -> "Success"
    case 400, 401, 403 -> "Client Error"
    case 500, 502, 503 -> "Server Error"
    default -> "Unknown"

// Multi-assignment from switch
a, b = switch condition:
    case true -> 10, 20
    case false -> 30, 40

// Block switches (multiple statements)
switch value:
    case "a":
        print("Found A!")
        do_something()
    case "b", "c":
        print("Found B or C!")
        do_something_else()
    default:
        print("Unknown value")

// Mixed syntax in same switch with yield
result = switch value:
    case 1:
        complex_processing()
        yield calculated_result    // Return value from case block
    case 2 -> simple_result
    default -> fallback_value

// Yield multiple values
a, b = switch condition:
    case "pair":
        calculate_values()
        yield first_result, second_result
    case "single" -> default_value, 0
```

### Loops

#### For Loops

```rad
// Iterate over list
items = ["a", "b", "c"]
for item in items:
    print(item)

// Iterate with index
for idx, item in items:
    print(idx, item)

// Iterate with unpacking (for lists of lists)
data = [["alice", 25], ["bob", 30], ["charlie", 35]]
for idx, name, age in data:
    print(idx, name, age)

// Multiple variable unpacking with zip
names = ["alice", "bob", "charlie"]
ages = [25, 30, 35]
cities = ["NYC", "LA", "Chicago"]
scores = [100, 90, 85]
for idx, name, age, city, score in zip(names, ages, cities, scores):
    print(idx, name, age, city, score)

// Iterate over map keys
person = {"name": "alice", "age": 25}
for key in person:
    print(key, person[key])

// Range iteration
for i in range(5):      // 0, 1, 2, 3, 4
    print(i)
```

#### While Loops

```rad
// Basic while loop
count = 0
while count < 5:
    print(count)
    count++

// Infinite loop with break
while:
    if condition:
        break
    // do something

// Continue statement
while condition:
    if skip_condition:
        continue
    process_item()
```

## Additional Control Flow

### Delete Statement

```rad
// Delete variables or data structures
my_var = "test"
del my_var              // Remove variable from scope

my_list = [1, 2, 3, 4]
del my_list[0]         // Remove first element

my_map = {"a": 1, "b": 2}
del my_map["a"]        // Remove key "a"
```

### Return Statement

```rad
fn calculate(x, y):
    if x < 0:
        return error("negative values not allowed")
    result = x * y
    return result      // Return single value

fn get_coordinates():
    x = 10
    y = 20
    return x, y        // Return multiple values

fn early_exit():
    if condition:
        return         // Early return with no value
    do_more_work()
```

### Yield Statement

```rad
// Yield can only be used in switch case blocks to return values
// (Examples shown in Switch Statements section above)

// NOT valid - yield cannot be used in regular functions
// fn generate_values():
//     yield i        // This would be an error
```

## Functions

### Function Definition

```rad
// Single-line functions
double = fn(x) x * 2
add = fn(a, b) a + b
greet = fn() print("Hello!")

// Block functions
calculate = fn(x, y):
    result = x * y + 10
    return result

// Function with multiple return values
coords = fn(point):
    x = point["x"]
    y = point["y"]
    return x, y
```

### Function Calls

```rad
result = double(5)      // 10
sum_val = add(3, 4)     // 7

// UFCS (Uniform Function Call Syntax)
"hello".upper()         // "HELLO"
[1, 2, 3].len()        // 3

// Built-in function assignment
my_upper = upper
"test".my_upper()      // "TEST"
```

### Named Arguments

```rad
print("hello", "world", sep="|")    // hello|world
```

## List Comprehensions

```rad
// Basic comprehension
numbers = [1, 2, 3, 4, 5]
squares = [x * x for x in numbers]

// With function calls
words = ["hello", "world"]
uppers = [upper(word) for word in words]

// Side effects (returns empty list)
[print(x) for x in items]
```

## Argument Parsing

### Basic Arguments

```rad
args:
    name str                    # Required string
    age int                     # Required integer
    height float               # Required float
    verbose bool               # Required boolean
    
    // Optional arguments
    role str?                  # Optional string (null if not provided)
    count int = 10            # Optional with default
    debug d bool              # Short form flag
```

### Argument Constraints

```rad
args:
    status str
    age int
    email str
    username str?
    password str?
    
    status enum ["active", "inactive", "pending"]
    age range [0, 120]        // Inclusive range
    age range (0, 120]        // Exclusive start, inclusive end
    email regex ".*@.*\\..*"
    username requires password   // If username provided, password required
```

## Shell Commands

### Basic Shell Commands

```rad
// Execute shell command (critical - exits on failure)
$!`echo hello`

// Capture output
code, stdout, stderr = $!`ls -la`

// Command with variables
filename = "test.txt"
$!`cat {filename}`
```

### Advanced Shell Commands

```rad
// Unsafe shell command (doesn't exit on failure)
unsafe $`command_that_might_fail`

// Shell command with modifiers
quiet $!`silent_command`           // Suppress command output
confirm $!`dangerous_command`      // Ask for confirmation
unsafe quiet $!`risky_command`     // Multiple modifiers

// Checked shell command with error handling
$`potentially_failing_command`
fail:
    print("Command failed!")
    exit(1)

$`another_command`
recover:
    print("Command failed, but continuing")
    // Script continues after this block
```

## JSON Processing and Display Blocks

### JSON Path Definitions

```rad
// Define JSON field mappings
Name = json.results[].name
Email = json.results[].email
Age = json.results[].age

// JSON path with wildcards
AllFields = json.results[].*         // All fields in each result
DeepFields = json.data.*.value       // Wildcard in path
Indexed = json.items[0].name         // Specific index
```

### Rad Blocks

```rad
// Rad blocks do HTTP request + JSON extraction + print table (all-in-one)
url = "https://api.example.com/users"
Name = json.results[].name
Email = json.results[].email

rad url:
    fields Name, Email
    sort Name
    // Automatically prints formatted table after extraction
```

### Display Blocks

```rad
// Display assumes lists are already populated and prints data as table
names = ["alice", "bob", "charlie"]
ages = [25, 30, 35]

// Display with pre-populated lists (no data source)
display:
    fields names, ages
    sort ages desc, names

// OR display with data source (runs JSON extraction + prints table)
data = [
    {"name": "alice", "age": 25},
    {"name": "bob", "age": 30}
]

Name = json[].name
Age = json[].age
display data:
    fields Name, Age
    sort Age desc, Name
```

## Advanced Features

### String Escape Sequences

```rad
// Supported escape sequences in strings
text = "Quote: \"Hello\""          // Escaped double quote
single = 'It\'s working'           // Escaped single quote  
backtick = `Backtick: \`command\`` // Escaped backtick
newline = "Line 1\nLine 2"         // Newline
tab = "Column1\tColumn2"           // Tab
backslash = "Path\\to\\file"       // Literal backslash
brace = "Not interpolated: \{var}" // Escaped brace (literal {)
```

### Defer and Error Defer Blocks

```rad
// Defer block - runs before script ends regardless of success/failure
defer:
    print("This always runs before script exits")
    cleanup_resources()

// Error defer block - only runs before script ends if an error occurs
errdefer:
    print("This only runs if script exits with error")
    emergency_cleanup()

process_data()
// More code can run here
exit(0)  // defer blocks run just before this
```

### Request Blocks

```rad
// Request blocks run JSON extraction algorithm but don't print table
// They populate lists with extracted field data
url = "https://api.example.com/users"
Name = json.results[].name
Email = json.results[].email

request url:
    fields Name, Email
    sort Name

// After this block, Name and Email contain the extracted data
print(Name)  // ["alice", "bob", "charlie"]
print(Email) // ["alice@example.com", "bob@example.com", "charlie@example.com"]
```

### Advanced Function Features

```rad
// Vararg parameters
fn sum_all(*numbers: int):
    total = 0
    for num in numbers:
        total += num
    return total

result = sum_all(1, 2, 3, 4, 5)

// Named-only parameters (after *)
fn format_text(text: str, *, uppercase: bool = false, prefix: str = ""):
    result = prefix + text
    return uppercase ? upper(result) : result

formatted = format_text("hello", uppercase=true, prefix=">>> ")

// Complex type annotations  
fn process_data(
    input: list[str], 
    callback: fn(str) -> bool,
    options: {config: str, debug?: bool}
) -> error|{processed: int, failed: int}:
    // function implementation
    return {processed: 10, failed: 0}
```

### Map Dot Syntax

```rad
person = {"name": "alice", "details": {"age": 25}}
name = person.name              // Same as person["name"]
age = person.details.age        // Nested access
```

### Error Handling

```rad
// Catch errors from function calls
result = catch risky_function()  // Returns error message as string if error occurs

// Without catch, errors propagate and exit
result = risky_function()        // Will exit script if error occurs

// Example function that returns error
fn risky_function():
    return error("something went wrong")

// Catching in various contexts
error_msg = catch foo()                    // Direct catch
array_with_error = [catch foo()]          // Catch in list
map_with_error = {"result": catch foo()}  // Catch in map
```

### Deferred Execution

```rad
defer:
    cleanup_resources()

// Code that might fail
process_data()
// cleanup_resources() will run regardless
```

### Pass Statement

```rad
if condition:
    pass        // Do nothing, placeholder
else:
    do_something()
```

## Operators

### Arithmetic

```rad
a + b       // Addition
a - b       // Subtraction
a * b       // Multiplication
a / b       // Division
a % b       // Modulo
```

### Comparison

```rad
a == b      // Equal
a != b      // Not equal
a < b       // Less than
a <= b      // Less than or equal
a > b       // Greater than
a >= b      // Greater than or equal
```

### Logical

```rad
a and b     // Logical AND
a or b      // Logical OR
not a       // Logical NOT
```

### Membership

```rad
item in collection      // Check if item exists in collection
item not in collection  // Check if item doesn't exist
```

## Scoping and Variables

```rad
// Global scope
global_var = "global"

if true:
    // Can access global variables
    print(global_var)
    
    // Variables defined in blocks persist after the block
    block_var = "accessible"

// block_var IS accessible here
print(block_var)  // Works fine

// Variables defined in for loops also persist
for i in range(3):
    loop_var = "also accessible"

print(i)          // Last value: 2
print(loop_var)   // "also accessible"

// Function closures
outer_var = 10
fn_with_closure = fn(x):
    return x + outer_var    // Can access outer_var
```

## Built-in Types and Methods

### String Methods

```rad
text = "hello world"
text.upper()            // "HELLO WORLD"
text.lower()            // "hello world" 
text.split(" ")         // ["hello", "world"]
text.replace("hello", "hi")  // "hi world"
```

### List Methods

```rad
items = [3, 1, 4, 1, 5]
items.len()             // 5
items.sort()            // Sort in place
items.reverse()         // Reverse in place
```

### Type Checking

```rad
// Runtime type checking happens automatically
// Type errors will be caught during execution
```

## Type System

Rad has a dynamic type system with runtime type checking. Here are the type annotations used in function signatures and documentation:

### Basic Types

```rad
str           // String type
int           // Integer type  
float         // Float type
bool          // Boolean type
any           // Any type (dynamic)
void          // No return value
error         // Error type
```

### Collection Types

```rad
list          // List of any type
list[T]       // List of specific type T
str[]         // List of strings (shorthand for list[str])
any[]         // List of any type
map           // Map/object with any keys/values
map[K,V]      // Map with specific key/value types
```

### Optional Types

```rad
str?          // Optional string (can be null)
any?          // Optional any type  
int?          // Optional integer
```

### Union Types

```rad
int|float     // Either int or float
str|list      // Either string or list
error|str     // Either error or string (common for fallible operations)
```

### Enum Types

```rad
["option1", "option2"]        // Enum with specific string values
["auto", "seconds", "millis"] // Enum for time units
```

### Function Types

```rad
fn(any) -> any                     // Function taking any, returning any
fn(any, any) -> bool               // Function taking two any params, returning bool
fn() -> any                        // Function with no params, returning any
```

### Complex Return Types

```rad
// Map with specific structure
{ "exists": bool, "size"?: int }

// Map with optional fields (? suffix)
{ "content": str, "created"?: bool }

// Nested structures
{ "epoch": { "seconds": int, "millis": int } }
```

### Variadic Parameters

```rad
*_items: any          // Variable number of any type
*_others: list|str    // Variable number of lists or strings
```

### Named Parameters

```rad
// Required named parameter
func(*, required_param: str)

// Optional named parameter with default
func(*, optional_param: str = "default")

// Mixed positional and named
func(_pos: str, *, named: int = 5)
```

### Parameter Constraints

```rad
// Parameter with underscore prefix (positional-only)
_path: str

// Named parameter (after *)
sep: str = " "

// Optional with default value
end: str = "\n"
```

### Real Examples from Built-ins

```rad
print(*_items: any, *, sep: str = " ", end: str = "\n") -> void
range(_arg1: float|int, _arg2: float?|int?, _step: float|int = 1) -> list[float|int] 
zip(*_lists: list, *, fill: any?, strict: bool = false) -> error|list[]
read_file(_path: str, *, mode: ["text", "bytes"] = "text") -> error|{ "size_bytes": int, "content": str|[int] }
```

## Rad Code Style

### Argument Block Formatting

```rad
// Good: Group args together, then constraints with empty line separation
args:
    name str           # Required string argument
    count int = 5      # Optional with default
    verbose v bool     # Boolean flag with short form
    email str?         # Optional argument
    
    name regex "^[a-zA-Z]+$"
    count range [1, 100]
    email requires verbose
```

### Comment Alignment

```rad
// Good: Align comments within reason
args:
    name str              # User's full name
    age int               # Age in years  
    email str?            # Contact email
    very_long_param str   # Don't align with this one if it's much longer
    city str              # Align with the shorter ones instead

// Bad: Inconsistent alignment
args:
    name str      # User's full name
    age int            # Age in years
    email str?  # Contact email
```

### Shell Command Delimiters

```rad
// Good: Use backticks to avoid delimiter conflicts
result = $!`echo "Hello world"`
output = $!`grep 'pattern' file.txt`
status = $!`curl -H "Content-Type: application/json" api.example.com`

// Avoid: Quotes can conflict with shell command content
// result = $!"echo "Hello world""  // This breaks
```

### Variable Naming

```rad
// Good: Use snake_case for variables
user_name = "alice"
max_retry_count = 3
is_valid = true

// Good: Common abbreviations are fine for CLI scripting
msg = "Hello world"
req = http_get(url)
cfg = load_config()
args = get_args()

// Good: Use descriptive names when clarity matters
user_data = load_user_info()
processed_items = filter(items, is_active)

// Avoid: Unclear single letters and confusing abbreviations
x = load_user_info()  // Too generic
a, b = 1, 2          // Meaningless names
usr_nm = "alice"     // Unclear abbreviation
```

### Function Definition Style

```rad
// Good: Clear, descriptive function names
calculate_tax = fn(amount, rate):
    return amount * rate

validate_email = fn(email):
    return "@" in email and "." in email

// Good: Single-line for simple functions
double = fn(x) x * 2
is_even = fn(n) n % 2 == 0
```

### Control Flow Formatting

```rad
// Good: Consistent indentation and spacing
if user.is_admin:
    print("Admin access granted")
    log_admin_action(user.name)
else if user.is_moderator:
    print("Moderator access granted")
else:
    print("Regular user access")

// Good: Switch formatting
result = switch status:
    case "active" -> "Running"
    case "paused" -> "Waiting" 
    case "stopped" -> "Inactive"
    default -> "Unknown"
```

### Collection and Data Structure Style

```rad
// Good: Readable list formatting with trailing commas
users = [
    {"name": "alice", "role": "admin"},
    {"name": "bob", "role": "user"},
    {"name": "charlie", "role": "moderator"},
]

// Good: Multiline maps with trailing commas
config = {
    "host": "localhost",
    "port": 8080,
    "debug": true,
}

// Good: Simple lists on one line when short
colors = ["red", "green", "blue"]
numbers = [1, 2, 3, 4, 5]

// Good: Clear JSON path definitions
UserName = json.users[].name
UserEmail = json.users[].email  
UserRole = json.users[].role
```

### String Interpolation Style

```rad
// Good: Clear variable interpolation
message = "Hello {user_name}, you have {message_count} messages"
file_path = "{base_dir}/{filename}.txt"

// Good: Format specifiers for numbers
price = "Total: ${amount:.2}"
percentage = "Progress: {progress:.1}%"
```

### Error Handling Style

```rad
// Good: Clear error handling
user_data = catch load_user(user_id)
if user_data == null:
    print_err("Failed to load user data")
    exit(1)

// Good: Inline error handling for non-critical operations
backup_data = catch load_backup()
if type_of(backup_data) == "error":
    backup_data = {"version": "none", "error": backup_data.str()}
```

This syntax reference covers the core language constructs of Rad. For built-in functions and more advanced features, refer to the main documentation.