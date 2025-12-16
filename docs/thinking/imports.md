# Imports & Modules

## 2025-04-27

We want ways to share and import function definitions between scripts, while avoiding dependency/third party issues like Python has.

Probably lots of lessons to be drawn from Cargo/Go.

We should be very judicious with this feature though -- a selling point of Rad is "batteries included" and we should make it as easy as possible to run scripts.

### Built-ins: No Modules Required

One key decision: we keep built-in functions available everywhere. No `import std` or official modules. Everything (`read_file`, `parse_json`, etc.) is just there.

This keeps the "batteries included" feel and means simple scripts stay simple.

### Built-in Shadowing and Aliasing

What happens if a user defines their own `read_file`?

```rad
read_file = fn(path):
    print("My custom read!")
    // ...

read_file("test.txt")  // calls the custom function, not the built-in
```

I think we should allow this. Built-ins are essentially pre-defined variables in the global scope. If you redefine one, you shadow it within that scope. This is consistent with how dynamic languages work.

An LSP feature that warns about shadowing a built-in identifier is the right solution here -- non-intrusive, helps users avoid mistakes without restricting the language.

Since built-ins are just values assigned to names, aliasing works naturally:

```rad
my_read_file = read_file   // alias the built-in
read_file = fn(path):
    // now I have a custom read_file
    // but I can still call the original via my_read_file

my_read_file("fallback.txt")  // calls the original built-in
```

This is useful for wrapping built-ins or passing them around as first-class values.

### Relative Imports

Start with relative imports only. Must use `./`, `../`, etc. -- no package registry, no external deps, just the ability to split your own code across files.

#### Syntax

Settled on Python-style `import`/`from`:

```rad
import "./utils.rad"                              // imports to `utils` namespace
import u "./utils.rad"                            // imports to `u` namespace (alias)
from "./config.rad" import api_key, timeout       // imports specific symbols into current scope
from "./config.rad" import api_key as key         // imports specific symbol with alias
```

Resolution is strictly relative (`./` or `../` required), and the `.rad` extension is mandatory. This keeps things explicit and avoids ambiguity.

#### Privacy

The `_` prefix enforces privacy -- `_`-prefixed symbols cannot be imported:

```rad
// in utils.rad
_internal_helper = fn():
    // implementation detail

public_func = fn():
    _internal_helper()
    // ...
```

```rad
// in main.rad
from "./utils.rad" import public_func        // works
from "./utils.rad" import _internal_helper   // ERROR: cannot import private symbol
```

This is simple, doesn't require new keywords, and is familiar from Python.

### Future: External Dependencies

Relative paths (`./`, `../`) refer to local files. Non-relative paths could be reserved for external dependencies:

```rad
import "./local/utils.rad"                    // local file (works now)
import "coollib"                              // future: external package
import "github.com/user/repo/pkg"             // future: external package
```

The path format provides a clear syntactic distinction between local project code and external libraries without introducing new keywords.

#### How This Would Work

When `rad` encounters `import "coollib"`:

1. Consult a manifest file to find the source/version for `"coollib"`
2. Ensure the correct version is available locally (downloading/updating if necessary)
3. Load the package from that location, using the same namespacing/`from` logic as relative imports

The manifest file (e.g., `Radfile`, `rad.toml`, `rad_packages.json`) would live in the project root and declare external dependencies with their sources and versions (Git URLs, potentially a future registry).

Packages could be stored in a project-local `_rad_deps/` directory or a shared cache -- TBD.

This is heavily inspired by Go modules. Not implementing now, but designing relative imports to not conflict with this future direction.

### Summary

1. **Rich global built-ins**: `read_file`, `parse_json`, etc. available everywhere, no `import` needed. Shadowable by users (with LSP warnings). Functions are first-class values.
2. **Local code sharing**: Strictly via relative imports (`import "./..."`, `from "./..." import ...`). Requires `.rad` extension. `_` prefix marks symbols as private/unimportable.
3. **Future external dependencies**: Plan to use non-relative paths (`import "packagename"`) to trigger lookup via a manifest file and `rad`'s future package management tooling.

## 2025-12-16

Spawned by https://github.com/amterp/rad/issues/74, I'm just re-reading the above previous thinking session and I just wanna point out some things I'm not sure about or disagree on:

`import u "./utils.rad"` vs. `import "./utils.rad as u"

^ `as u` would be more consistent with syntax for aliasing individual `from` imports.

> the `.rad` extension is mandatory. This keeps things explicit and avoids ambiguity.

Not sure I agree with this. Seems better if we don't force users to name their files a certain way, somewhat arbitrarily?
