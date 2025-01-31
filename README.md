# 🤙 Rad - Request And Display

A powerful command-line tool and domain-specific language for effortlessly querying and displaying JSON API data. Simplifies the process of writing, managing, and sharing API query scripts.

## [Documentation](https://amterp.github.io/rad/)

Docs are still a work in progress, but there's enough to get you started!

Feel free to check out the [**Getting Started**](https://amterp.github.io/rad/guide/getting-started/) guide :)

## Installation

### macOS

```shell
brew tap amterp/rad
brew install rad
```

Other than building from source, Rad is not available for other platforms/package managers (yet).

## Status 📊

⚠️ **Rad is still in early development!** ⚠️

Rad is a working CLI tool and interpreter that can run useful RSL scripts.

It's complete enough to be useful, but please don't be surprised when major parts of it change, or you hit bugs and rough edges. That said, please do give it a try, I'd love to hear your experience and any feedback :)

Below is a quick glimpse of major items that've been implemented and are missing.

### What's being worked on 🚧

- LSP language server ([RLS](./rsl-language-server))
- [Tree sitter implementation](https://github.com/amterp/tree-sitter-rsl)
- [Visual Studio Code extension](./vsc-extension)

### What's planned 🌱 

- Many more language features
- Polished syntax error feedback
- `rad` script management features & helpers
- JetBrains IDE plugin

## What problem does this solve? 🎯

Many backend services expose JSON/REST APIs containing valuable information that users often need to query and view ad hoc (e.g. DevOps).
While various tools exist for this purpose, they often come with drawbacks such as complex syntax, steep learning curves, or the need for extensive setup.
What's needed is a flexible, easy, and efficient way to:

1. Define and parameterize queries
2. Extract specific information from API responses
3. Display the data in a user-friendly format

## How does Rad solve it? 🛠️

- Rad comes with a domain-specific language called RSL (Rad Scripting Language).
- RSL is purpose-built for this problem: to efficiently express what to query, the data to extract, and how to display it.
- `rad` is a command-line tool which runs these scripts, handling argument parsing, query execution, and result display + formatting.

## Minimal Example 🌟

```
args:
    repo string    # The repo to query. Format: user/project
    limit int = 20 # The max commits to return.
    
url = "https://api.github.com/repos/{repo}/commits?per_page={limit}"

Time = json[].commit.author.date
Author = json[].commit.author.name
SHA = json[].sha

rad url:
    fields Time, Author, SHA
```

Example invocation (let's call the script `commits.rad`):

```
> rad commits spf13/cobra 3

Time                   Author                 SHA
2024-07-28T16:18:07Z   Gabe Cook              756ba6dad61458cbbf7abecfc502d230574c57d2
2024-07-16T23:36:29Z   Sebastiaan van Stijn   371ae25d2c82e519feb48c82d142e6a696fd06dd
2024-06-01T10:31:11Z   Ville Skyttä           e94f6d0dd9a5e5738dca6bce03c4b1207ffbc0ec
```

1. This script takes two args: a repo string and an optional limit (defaults to 20).
  - The `#` comments are read by Rad and used to generate helpful docs / usage strings for the script.
2. It uses string interpolation to resolve the url we will query, based on the supplied args.
3. It defines the fields to extract from the JSON response.
4. It executes the query, extracting the specified fields, and displays the resulting data as a table.
- We keep this example somewhat minimal - there are RSL features we could use to improve this, but it's kept simple here.
- Some alternative valid invocations for this example:
  - `rad commits.rad <repo>`
  - `rad commits.rad --repo <repo> --limit <limit>`
  - `rad commits.rad --limit <limit> --repo <repo>`
  - `rad commits.rad <repo> --limit <limit>`

## Alternatives 📚

- **Bash**
  - Bash, using a combination of `curl`, `jq`, and/or `column`, is an excellent choice.
  - Bash is the primary tool I'd use, outside of Rad.
  - But, as much as I like this toolset, Bash is (in my opinion) not syntactically friendly and simple things can be deceivingly laborious to do.
    - Crucially, arg parsing is decent for simple cases, but more complex ones quickly get unwieldy. For the sorts of scripts Rad targets, that's important.
  - There's also a bit of a learning curve. Bash can be intimidating to devs that haven't used it a lot, as can `jq` syntax.
- **Python, Ruby, JavaScript, Rust, etc**
  - General purpose and very flexible, which can be great. However, this can also mean significantly more code is required to get the behavior you want.
  - Managing dev environments/installations and sharing scripts in a reliable way can be onerous, depending on the specific language.
- **HTTPie**
  - This eases some of the difficulties with using `curl` and `jq` in Bash, but does not help with the rest, e.g. arg parsing, more complex bash logic, etc.

## Why Rad? 🚀

### Rad Scripting Language (RSL)

- Rad (and its accompanying language RSL) allows you to be *efficient* in writing your scripts. What does this mean?
- The syntax is designed so that *every* line gets you closer to your goal of querying and displaying JSON.
- *Every* line is doing heavy lifting; it's dense with meaning.
- Think of it like this: for every line of RSL you write, Rad saves you from writing several equivalent lines in another language. That saved work has been shifted away from you and into the design and building of Rad, and what Rad is doing behind-the-scenes with the scripts you write.
- It allows you to, in fewer (and simpler) lines, express your intent much more directly.
- It does all this while staying simple and easy to learn. You can be writing great scripts within the first hour of getting started.

### Easily shareable

- Everything you need to write JSON query scripts is built into Rad and RSL. You don't need to download dependencies, and if a Rad script runs on your machine, you can be confident it will run on others' machines too.

### Encourages self-documenting scripts

- Rad encourages the documentation of scripts by having it built into the language. There's syntax for documenting the overall script as well as its args.
- The declaration of args themselves also provides useful information that Rad leverages to generate helpful usage strings for your scripts, such as types or constraints, to make them user-friendly.
- RSL's syntax is designed to be self-explanatory and readable to anyone. No arcane use of symbols or inscrutable keywords.

## Why *not* Rad? ⚠️

- Rad aims to make writing 95% of your scripts better and easier. However, the last 5% may involve bespoke, complex logic, better suited for general-purpose programming languages such as Python or Bash.
- You can get quite far with Rad, as RSL provides utilities for the most common things devs would want, but it will inevitably be lacking something that a more general-purpose programming language would offer.
- The bet is that these complex scripts are few and far between, so that Rad can make your life easier 95% of the time, and require you to pull out the heavy-duty 'general' tooling only once in a while.
