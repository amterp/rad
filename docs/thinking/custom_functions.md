# Custom Functions & Lambdas

## 2025-04-27

Some thinking over on [stashes.md](./stashes.md).

TLDR of the syntax I concluded on there:

```rad
normalize = fn(x) x.trim().lower()
 
normalize(mystring)

normalize = fn(x):
    out = x.trim().lower()
    return out

provide = fn() 5
provide()  // returns 5

multiply = fn(x, y) x * y

mylist.map(fn(x) x.upper())
mylist.map(upper)  // technically, it'd need to redefine all my built-ins as function vars, so they can be passed this way
```

---

Related topic: [imports.md](./imports.md).

## 2025-06-03

More 'complete' non-lambda function definitions

Minimal signatures allowed:

```rad
fn dosomething(n):
    return n + 5
```

But can add if you want. They *do* get runtime validated (and statically analyzed, best effort):

```rad
fn dosomething(n: int) -> int:
    return n + 5
```

Can mix typing and no typing:

```rad
fn dosomething(n: int, name) -> int|float:
    if name == "alex":
        return n / 5
    return n + 5
```

The above *requires* two args, tho. Below does not, but makes name default to null.

```rad
fn dosomething(n: int, name: string?) -> int|float:
    if name == "alex":
        return n / 5
    return n + 5
```

Can also give non-null default value:

```rad
fn dosomething(n: int, name: string = "alice") -> int|float:
    if name == "alex":
        return n / 5
    return n + 5
```

The following will fail if name is alex, because we'll try to return a string when int|float only allowed.
Not super sure how I feel about that 'or' notation for the type system. Alternatives?

```rad
fn dosomething(n: int, name: string = "alice") -> int|float:
    if name == "alex":
        return "boo"
    return n + 5
```

The following *always* requires a nullable string to be returned, can be used as an error.
Note we wrap multiple return types in () in order to avoid parsing ambiguities.

```rad
fn dosomething(n: int, name: string = "alice") -> (int|float, string):
    if name == "alex":
        return 0, "alex not allowed"
    return n + 5, null
```

Can make it optional by adding question mark, this will default it to null if unspecified

```rad
fn dosomething(n: int, name: string = "alice") -> (int|float, string?):
    if name == "alex":
        return 0, "alex not allowed"
    return n + 5  // error defaults to null
```

If you add an exclamation point, then if the string is non-null, but not assigned to a variable, then we will panic.
It's kinda like an exception going uncaught, but if it's assigned, it's caught.

```rad
fn dosomething(n: int, name: string = "alice") -> (int|float, string?)!:
    if name == "alex":
        return 0, "alex not allowed"
    return n + 5  // error defaults to null
    
//

a, b = dosomething(2, "alex")  // no panic, 'b' defined
a = dosomething(2, "alex")  // panic
```

A little unsure about the exclamation point. If we had a return typing `-> string!` people might reasonably read that to simply mean
"a non-null string", rather than "will panic if you don't assign this". What about two exclamations i.e. `-> string!!` ?
That might have a similar problem, but I would argue `!` is a more common operator to say 'non-null' and so using it
for something else is more egregious than using `!!` which I don't think is nearly as universal?

We can use parentheses to apply optionality to unions: `(int|float)?`

Another thing: varying number of returns

```rad
fn dosomething(n: int, name: string = "alice") ->( int|float, string?):
    if name == "alex":
        return 0, "alex not allowed"
    return n + 5  // error defaults to null
```

We saw this before, this is allowed due to the optional string. If we had no typing of the return, then when an incorrect number of values are assigned, we default them to null.
I don't think there's a practical alternative to this?

```rad
fn dosomething(n: int, name: string = "alice"):
    if name == "alex":
        return 0, "alex not allowed"
    return n + 5
    
a, b, c = dosomething(2, "bob")  // a defined, b and c are null
```

However, if we do define the return typing, and someone assigns too many variables, then we fail (and try to statically analyze against it):

```rad
fn dosomething(n: int, name: string = "alice") -> (int, string?):
    if name == "alex":
        return 0, "alex not allowed"
    return n + 5
    
a, b, c = dosomething(2, "bob")  // error, because 'c' won't be defined, according to typing
```

Below, you can make a parameter nullable with minimal other typing:

```rad
fn dosomething(n: int, name?):
    if name == "alex":
        return 0, "alex not allowed"
    return n + 5
    
dosomething(4)        // valid
dosomething(4, "bob") // valid
dosomething(4, 5)     // valid
```

With multiple return, how do chained function calls work? e.g.

```rad
myfoo(returns2things())
```

How does `myfoo` receive the 2 return values of the inner function? Do they simply go into 1st/2nd arg? What if we wanted just the 1st or 2nd return value to
get passed into `myfoo`? Should we offer a syntax for that? Or require users to just assign `returns2things` to variables first, and then invoke `myfoo` with them?

If you have a function with untyped return values, and it sometimes returns 2 things, and sometimes 3, and you do this:

```rad
myfoo(return2or3things(), return1thing())
```

then what happens? if `return2or3things` was typed, then we'd know it can return 3 things, and so if it only returns 2,
then we could make the third `myfoo` arg `null` (similar to how we do assignments), before passing in the `return1thing` as the 4th arg.
But without typing on `return2or3things`, we wouldn't know to do this, so the arg position that `return1thing` goes into might depend on what 
`return2or3things` returns on runtime? This sounds quite bad. Maybe we just disallow ambiguously nesting functions with untyped return types, like this?

Another thing: we also want to allow var args.

```rad
fn dosomething(n: int, names: ...string):
    pass
```

When invoking, can spread:

```rad
names = [ ... ]
dosomething(5, ...names)
```

I think that works?

We also want to allow defining func params and say "params beyond this point are only named, not positional". For that:

```rad
fn dosomething(n: int, ..., name, age):
    pass
```

I think this would work? You cannot pass `name` or `age` positionally, only named.
This is inspired by how Python does it (using `*` instead), but I don't know exactly why Python is doing it -- I know it was
implemented later, but so perhaps it was designed this way because they were constrained by backwards compatibility? If you
had a clean slate, is this still the best way to do it?

Another tangent back to the `string?!` error syntax, maybe we drop this `!` idea and simply allow `error`, which is basically equivalent to `string?!` would otherwise be.

## 2025-06-04

Reddit discussion: https://www.reddit.com/r/ProgrammingLanguages/comments/1l2nbpj/feedback_idea_for_error_handling/

- Skepticism on Go-like approach, misunderstanding about how Go-like (or not) Rad's proposed approach is. 
  - Communicate more clearly the difference from Go, maybe don't compare at all.
- Proposal contributes to learning curve. Approach is alien to newcomers.
- Complicates nested function calls.
- Strong support for a `Result`-like union as an alternative.
  - Would need pattern matching. Recommended check out OCaml for example.
- How does propagation work? Good question.
  - Propagate with question mark? Look at how that works in other languages.
  - Some alternatives:

```
a, err = myfoo()
err?
// or does exclamation point make more sense?
err!
```

```
a = myfoo()!
// these alternatives probably also require the enclosing function to declare 'error' in its return signature
// I guess we only allow one error in the signature? probably as the last output?
```

- Error structure customization. I proposed just a string.
  - Customizability suggested potentially? Probably not.
  - Perhaps a canonical structure e.g. map with fields "msg", "stacktrace"
- Clarification on 'CLI language'. Invokes Rad being a bash/oil/fish-like language for some, which it's not.
  - ?? How better communicate to avoid this misunderstanding?

### Union Approach

This document has proposed a non-union approach above, where the error is separated. Let's consider a union approach and what that could look like.

Points of inspiration: Zig, Rust

```
fn myfoo(op: string, n: int) -> float|error:
    if op == "add":
        return n + 5
    if op == "divide":
        return n / 5
    return error("Invalid op: {op}")

a = myfoo()

// 'a' is technically a union type. it's either a float or an error.

// could propagate like this maybe? treats error in a special way
try a

// more likely, you do this
a = try myfoo()

// this will leave with 'a' definitely being non-error. it can still be null if our return signature is -> float?|error
// what I don't like about this though is that, we're moving away from this idea of 'just panic if error is unhandled'
// and having that be the default behavior. 'try' should be opt-out, not opt-in.

// would it be confusing to sorta flip what 'try' does? rather than propagating, it *disables* propagating? i.e.
a = try myfoo()

// allows 'a' to be an error? if we hadn't written 'try', it would propagate it up? think more.

// in terms of handling with pattern matching

a = try myfoo()
myfloat = switch type_of(a):
    case "float" -> a
    case "error" -> 0

// this reuses switch. it sorta is like defaulting to 0 if there's an error. it's not safe tho, could typo the cases.

myfloat = match a:
    float -> a
    error -> 0
    
// don't love that 'match' term. here we're sorta implying it works for just types, and if it doesn't then what's
// point of 'switch'? What if did commit to that tho? cut out 'case' keyword, and be more clever on what we will receive?

myfloat = switch a:
    float -> a
    error -> 0
    "example" -> -1
    
// silly example but sorta demonstrates ^. You can switch on types, and then specific values for equality. Only cornor
// where this could cause issues is if we ever wanted to actually allow storing references to types as variables. e.g.

myvar = float
switch myvar:
    float -> ...
    
// here, should myvar match on that case or not? It's *equal* to float, but it *is not a float*,
// but rather a reference to the type 'float'.

// in any case, it'd be nice to have a 'catch'-like syntax.

a = myfoo() catch 0

// worried it's not that self-explanatory, and zig is not popular enough that i can reasonably lean on that to
// lower the learning curve. 
```

Things to think for next time:

- Flip `try` from Zig to opt-out of propagating? Alternative keyword?
- 'catch' equivalent for rad?
- Switch statements on types?
- try to conclude: union type, or separate 'error' return?
