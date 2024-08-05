# Rad - Request And Display

A tool for easily writing JSON API query scripts.

## What problem does this solve?

- Lots of backends expose a JSON / REST API.
- Many of these contain useful information that people would like to query and view, often ad hoc, for various reasons.
- It'd therefore be useful to have easy and quick-to-use CLI scripts that fetch this information and displays it.
- You can write these scripts/queries using various existing tools, but they each fall short in some way. See [Alternatives](#alternatives).
- What we want is a flexible, easy, and *effective* way to express:
  1. What kind of query we want to run, including parameterization
  2. What information we want to *extract* from the response
  3. How we want to view this data, or display it to user

## How does it solve it?

- Rad comes with a domain-specific language called RSL (Rad Scripting Language).
- RSL is designed specifically for this domain: to effectively express queries, the data to extract, and how to display it to the user.
- `rad` is a command-line tool for running and managing these scripts.
- When invoked on a script, `rad` will interpret the script, validate and pass user-supplied args to the script, and execute it.
  - The script tells `rad` what arguments it expects.
- `rad` helps the user manage their scripts, so they can build up an organized repertoire of RSL queries, runnable within a finger's reach. 

## Examples

### Minimal example

```
args:
    repo string # The repo to query. Format: user/project
    limit int = 20 # The max commits to return.
    
url = "https://api.github.com/repos/{repo}/commits?per_page={limit}"

Author = json[].commit.committer.name
Time = json[].commit.committer.date
Message = json[].commit.message

rad url:
    Time, Author, Message
    sort Time desc, Author, Message
```

Example invocation: `rad commit junegunn/fzf 5`

```
TODO
```

1. This script (let's call it `commits`) takes a repo string and an optional limit (defaults to 20) as args, this is declared at the top.
2. It uses string interpolation to resolve the url we will hit, based on the supplied args.
3. It defines the fields we'd like to extract from the JSON response.
4. It executes the query, extracting the supplied fields, and displaying the resulting data as a table, sorted first by time (descending), then author, then message, latter two in ascending order.
- We keep this example somewhat minimal - there are RSL features we could use to improve this, but it's kept simple here.
- Some alternative valid invocations:
  - `rad commit --repo junegunn/fzf --limit 5`
  - `rad commit --limit 5 --repo junegunn/fzf`
  - `rad commit junegunn/fzf --limit 5`

### More elaborate example

## Alternatives

- curl
- Python
- bash, jq
  - These are excellent.
- HTTPie

## Why Rad?
