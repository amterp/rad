---
title: Args 
---

## Basics

```rsl
args:
    argVar "alias"? shorthand? type = default? # Description?
```

```rsl
args:
    name n string # A required arg 'name' which can be specified positionally or also with -n.
    is_employee "is-employee" bool # Variable for script is is_employee, but users will see it as is-employee.
    
if is_employee:
    print("{name} is an employee.")
else:
    print("{name} is not an employee.") 
```

```rsl
args:
    name string
    age_years "age-years" int
    height float # Height in meters
    is_employee "is-employee" e bool
    friends string[] # Specified as e.g. Alice,Bob
    nationality n string = "Australian" # Defaults to this if not specified.
```

Example usage:

```
script Charlie 30 -e --friends David,Eve -h 1.86
```

## Constraint Statements

### Enum

```rsl
args:
    name string
    name enum ["alice", "bob", "charlie"]
```

```
// valid!
myscript alice

// invalid, will print error
myscript david
```
