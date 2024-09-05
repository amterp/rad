# switch stmts

## current plan

```
args:
    base string

finalBase, title = switch base:
    case "github", "gh": "api.github", "Github"
    case "gitlab": "gitlab", "Gitlab"

url = switch:
    case: "https://{finalBase}.com/repos/{repo}/commits?per_page={limit}"
    case: "https://{finalBase}.com/repos/{owner}/{project}/commits?per_page={limit}"
```

i.e. single-line cases

## eventually

```
finalBase, title = switch base:
    case "github", "gh":
        yield "api.github", "Github"
    case "gitlab":
        print("Jokes, Gitlab not supported yet!")
        yield "api.github", "Github"
```

