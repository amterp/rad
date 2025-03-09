# Contributing

Thanks for your interest in contributing to rad! It's much appreciated - this document will help you get started.

## Reporting Issues

- Please first check if the feature request or bug report has already been posted.
- If not, feel free to raise a new issue.
- If it's a bug report, it'd be very handy if you could include a few things:
  1. What version of rad you're using (`rad -v`).
  2. What happens, and the behavior you expect.
  3. Replication steps, ideally with an RSL script which replicates the issue.

## Contributing PRs

### Setup

Clone the repo and `cd` into it.

```shell
git clone https://github.com/amterp/rad
cd rad
```

The project currently uses Go 1.23. Ensure you've got the necessary `go` CLI tooling installed.

For example, to check (1.24+ is okay):

```
> go version
go version go1.24.1 darwin/arm64
```

One of the Makefile steps includes an automatic go-imports fixer. It requires `goimports`:

```shell
go install golang.org/x/tools/cmd/goimports@latest
```

Now, invoke the Makefile:

```shell
make all
```

It should format, build, and run tests. If it all passes, you should be good to go! Let me know if you have issues.

If you're using GoLand, the repo includes a few run configurations that may be helpful:

- **Rad**: Runs rad with arguments (handy for debugging).
- **Tests**: Runs all the tests.
- **make all**: Runs make all from the IDE.

### Submitting PRs

1. Fork the repo, create a feature branch, and commit your changes.
2. Push to your fork and open a PR.
3. If your PR isn't getting attention, please ping me on it!

- Please aim to respect the [code style & conventions](#code-style--conventions).
- Include tests, preferably comprehensive ones.
- If your changes impact user documentation, consider updating it.
  - If you're not comfortable writing user docs, feel free to leave it out. I can follow up on it :) 

### Code Style & Conventions

I started this project with very little Go knowledge, and so have definitely broken many idioms and conventions.

Just follow standard Go practices - you'll see this broken in many places in the existing code, so don't
blindly use it for examples, generally speaking :^) .

That said, here are some specific callouts:

- Aim for self-documenting code. Good variable names, smaller functions with descriptive names, etc.
- Use comments judiciously - convey intent and "why" of code, if it's not already obvious.
- Your commit messages are also good sources of information: include breakdowns of decisions you made, motivations, etc.
  - Ideally, our `git blame` will be a reliable source of information documenting why the code is the way it is.
- Try to keep commits small. If you can separate conceptually-unrelated changes into commits that each compile & pass tests, that's ideal!
- Run `make format` before making commits.
- `core` is unfortunately a big folder and package - untangling it into smaller packages at this point is a little tricky.
  - If you can, with new code, try to package it appropriately.

### Code Pointers

- [`main.go`](./main.go) is our entry point.
- [`core/runner.go`](./core/runner.go) contains logic for parsing arguments, reading the input script, and executing it.
- [`core/interpreter.go`](./core/interpreter.go) is the meat of where we step through instructions.
  - Specifically, it is given the tree sitter concrete syntax tree (CST), and we step through it to execute it.
- [`core/global.go`](./core/global.go) contains some global state & variables.
  - A lot of things in here are abstractions that enable us to swap in implementations for testing.
- [`core/funcs.go`](./core/funcs.go) defines most of our inbuilt functions.
- [`core/testing`](./core/testing) is where we define the bulk of our tests.
  - They tend to be end-to-end tests which define some RSL code, execute it, and assert against stdout/stderr from rad.
- [`core/type_rsl_value.go`](./core/type_rsl_value.go) defines a `RslValue` struct type.
  - It represents runtime variables when rad is interpreting a script, and gets passed around a lot.

## Questions?

Feel free to contact me directly or post your question [here](https://github.com/amterp/rad/discussions)!
