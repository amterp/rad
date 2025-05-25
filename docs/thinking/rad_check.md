# Rad Check

## 2025-04-28

We want a way to statically validate/check scripts without running them. For example, something like:

```
> rad check <script>
> rad validate <script>
> rad lint <script>
> rad check <script> --lint
> rad check <script> --no-lint
> rad check <script> --errors
```

Some things it could check for:

- ERROR/MISSING nodes in the CST
- References to undefined references
- Invalid `range` constraint e.g. `[2, 1]`
- Invalid built-in (or custom) function calls (wrong args, etc)

Crucially, we could and **should** share the implementation with our LSP. Whether the LSP directly invokes `rad check`
or (more realistically) they both invoke shared Go code and simply action the results differently (LSP sends diagnostics,
`rad check` compiles a human-friendly report of issues; perhaps a `--json` flag can be used for program-friendly output).

## 2025-05-25

- `--format=json`
- Allow linting (hint severity)
