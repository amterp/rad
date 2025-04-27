# Imports & Modules

We probably want ways to share and import function definitions between scripts, while avoiding dependency/third party hell like Python has.

Probably lots of lessons to be drawn from Cargo/Go.

We should be very judicious in this feature though - a selling point of RSL/rad is "batteries included" and we should make it as easy as possible to run scripts.

## Design

We keep built-in functions: no "official" modules. It's all available and accessible.

Start off with "relative imports". Must use `./`, `../`, etc.

Summary provided after back-n-forths with LLM:

---

**1. Built-in Shadowing and Aliasing:**

- **Accepted Behavior:** If a user defines `read_file = fn(...) ...`, they shadow the built-in `read_file` within that scope. This is consistent with dynamic languages and treats built-ins as pre-defined variables in the global scope.
- **LSP Hints:** An LSP feature that warns about shadowing a built-in identifier is the perfect non-intrusive way to help users avoid mistakes without restricting the language.
- **Aliasing:** Since built-ins are just values assigned to names in the global scope, `my_read_file = read_file` works naturally for creating aliases or passing built-ins around.

**2. Syntax for Relative Imports:**

- We're settled on the following for importing from *local relative* `.rsl` files:
    - `import "./utils.rsl"` (imports to `utils` namespace)
    - `import u "./utils.rsl"` (imports to `u` namespace)
    - `from "./config.rsl" import api_key, timeout` (imports specific symbols into current scope)
    - `from "./config.rsl" import api_key as key` (imports specific symbol with alias)
- Resolution is strictly relative (`./` or `../` required), `.rsl` extension mandatory.
- `_` prefix enforces privacy (cannot be imported).

**3. Future Package Management Strategy (Based on Your Idea):**

- **Concept:** This is a great way to delineate!
    - **Relative Paths (`./`, `../`):** Always refer to local files relative to the importer. Works *now*.
    - **Non-Relative Paths (`import "coollib"`, `import "github.com/user/repo/pkg"`):** Reserved for future use. When implemented, these would signal a request for an *external dependency*.
- **Implications for the Future:**
    - The `rad` CLI tool would need a package management aspect (fetch, store, resolve versions).
    - A manifest file (e.g., `Radfile`, `rsl.toml`, `rsl_packages.json`) would be required in the project root to declare these external dependencies and their sources/versions (e.g., Git URLs, potentially a future registry).
    - When `rad` encounters `import "coollib"`, it would:
        1.  Consult the manifest file to find the source/version for `"coollib"`.
        2.  Ensure the correct version is available locally (downloading/updating if necessary, perhaps into a project-local `_rsl_deps/` directory or a shared cache).
        3.  Load the package from that location, using the same namespacing/`from` logic as relative imports.
- **Benefit:** This provides a clear syntactic distinction between local project code and external libraries *without* introducing new keywords for imports. The *meaning* of the import string changes based on its format.

**Summary of Final Design:**

1. **Rich Global Built-ins:** `read_file`, `exec`, etc., available everywhere, no `import` needed. Shadowable by users (with LSP warnings). Functions are first-class values.
2. **Local Code Sharing:** Strictly via relative imports (`import "./..."`, `from "./..." import ...`). Requires `.rsl` extension. `_` prefix marks symbols as private/unimportable.
3. **Future External Dependencies:** Plan to use non-relative paths (`import "packagename"`) to trigger lookup via a manifest file and `rad`'s future package management tooling.

---