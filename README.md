# Rad - Request And Display

A tool for easily writing JSON API query scripts.

## What problem does this solve?

- Many backends expose a JSON / REST API.
- Many of these contain useful information that people would like to query and view, often ad hoc.
- It'd therefore be useful to have easy, quick-to-use CLI scripts that fetch and display this information.
- You can write these scripts/queries using various existing tools, but they each have their downsides. See [Alternatives](#alternatives).
- What we want is a flexible, easy, and *efficient* way to express:
  1. What kind of query we want to run, including parameterization
  2. What information we want to *extract* from the response
  3. How we want to view this data, or display it to user

## How does it solve it?

- Rad comes with a domain-specific language called RSL (Rad Scripting Language).
- RSL is purpose-built for this: to efficiently express the queries, what data to extract, and define how it should be displayed.
- `rad` is a command-line tool for running and managing these scripts.
- When invoked on a script, `rad` will interpret the script, validate and pass user-supplied args to the script, and execute it.
  - The script tells `rad` what arguments it expects.
- `rad` helps the user manage their scripts, enabling them to build an organized repertoire of RSL queries, right at their fingertips.

## Examples

### Minimal example

```
args:
    repo string # The repo to query. Format: user/project
    limit int = 20 # The max commits to return.
    
url = "https://api.github.com/repos/{repo}/commits?per_page={limit}"

Time = json[].commit.author.date
Author = json[].commit.author.name
SHA = json[].sha

rad url:
    Time, Author, SHA
    sort Time desc, Author, SHA
```

Example invocation (let's call the script `commits.rad`):

```
> rad commits spf13/cobra 3

Time                   Author                 SHA
2024-07-28T16:18:07Z   Gabe Cook              756ba6dad61458cbbf7abecfc502d230574c57d2
2024-07-16T23:36:29Z   Sebastiaan van Stijn   371ae25d2c82e519feb48c82d142e6a696fd06dd
2024-06-01T10:31:11Z   Ville Skyttä           e94f6d0dd9a5e5738dca6bce03c4b1207ffbc0ec
```

1. This script (let's call it `commits`) takes a repo string and an optional limit (defaults to 20) as args.
   - The `#` comments are read by Rad and used to generate helpful docs / usage strings for the script.
2. It uses string interpolation to resolve the url we will query, based on the supplied args.
3. It defines the fields to extract from the JSON response.
4. It executes the query, extracting the specified fields, and displays the resulting data as a table, sorted first by time (descending), then author, then SHA (both ascending).
- We keep this example somewhat minimal - there are RSL features we could use to improve this, but it's kept simple here.
- Some alternative valid invocations for this example:
  - `rad commits <repo>`
  - `rad commits --repo <repo> --limit <limit>`
  - `rad commits --limit <limit> --repo <repo>`
  - `rad commits <repo> --limit <limit>`

## Alternatives

- **bash**
  - Bash scripts, using a combination of `curl`, `jq`, and/or `column`, are an excellent choice.
  - Without Rad, Bash is what I'd be using, as I did before creating Rad.
  - As much as I like this toolset, bash is (imo) not syntactically friendly and simple things can be deceivingly laborious to do.
    - Crucially, arg parsing is decent for simple cases, but more complex ones quickly get unwieldy. For the sorts of scripts Rad targets, that's important.
  - There's also a bit of a learning curve. Bash can be intimidating to devs that haven't used it a lot, as can `jq` syntax.
- **Python, Ruby, JavaScript, Rust, etc**
  - General purpose and very flexible. However, this can also mean more code is required to get the behavior you want.
  - Not at a bad choice, but managing dev environments/installations and sharing scripts in a reliable way can be onerous.  
- **HTTPie**
  - This eases some of the difficulties with using `curl` and `jq` in bash, but does not help with the rest, e.g. arg parsing, more complex bash logic, etc.

## Why Rad?

### Rad Scripting Language (RSL)

- Rad (and its accompanying language RSL) are *efficient*. What does this mean?
- The syntax is designed so that *every* line gets you closer to your goal of querying and displaying JSON.
- *Every* line is doing heavy lifting; it's dense with meaning. Equivalent code might be several lines in other languages.
- Think of it like this: for every line of RSL you write, Rad saves you from writing multiple lines in another language. That work has been shifted into the design and building of Rad, and what Rad is doing behind-the-scenes with the scripts you write.
- It allows you to, in fewer (and simpler) lines, express your intent much more directly.

### Easily shareable

- Everything you need to write JSON query scripts is built into Rad and RSL. You don't need to download dependencies, and if a Rad script runs on your machine, you can be confident it will run on others' machines too.

### Script management

- Rad comes with CLI tools for managing your scripts, making them quick and easy to search, access, and use.

### Framework designed for scripts

- Rad encourages the documentation of scripts by having it built into the language. There's syntax for documenting the overall script as well as its args.
- The declaration of args themselves also provides useful information that Rad leverages to generate helpful usage strings for your scripts, to make them user-friendly.

## Why *not* Rad?

- Rad aims to make writing 95% of your scripts better and easier. However, the last 5% may involve bespoke, complex logic, better suited for general-purpose programming languages such as Python or Bash.
- You can get quite far with Rad, as RSL provides utilities for the most common things devs want, but it will inevitably be lacking something that a more general-purpose programming language would offer.
- The hope is that these complex scripts are few and far between, so that Rad can make your life easier 95% of the time, and require you to pull out the heavy-duty 'general' tooling only once in a while.
