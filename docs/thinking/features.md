# Feature Scratchpad

## 2024-08-15

```
"""
Hottest locations in city.
"""

args:
    city string # The city to check the weather for.
    api string = "apple" # Weather API to query.
    hideLocation bool = false # If true, don't display the Location column
    
apiKey = env("WEATHER_API_KEY")
username = env("WEATHER_USERNAME", "not set")

url = switch api:
    case "apple", "aapl": "https://appleweather.com/"
    case "banana": "https://bananaweather.com/"
    case "coconut":
        if city == "Toronto":
            "https://coconutweather.ca/"
        else:
            "https://coconutweather.com/"

Location = json...
Temp = json... 

rad url:
    if hideLocation:
        fields Temp
    else:
        fields Location, Temp
    
    sort Temp desc, Location  // Location ignored? Or should also if-else this
```

- env
- switch expression
- if stmt in rad block
- fields keyword in rad block

---

```
Name = json.results[].name
Passengers = json.results[].passengers
Cost = json.results[].cost_in_credits
rad url:
    Name, Passengers, Cost
    sort Passengers desc, Cost, Name
    table markdown
    limit limit
    Name:
        max_width 16
        color green
        
// INSTEAD

json.results[]:
    Name = .name
    Passengers = .passengers
    Cost = .cost_in_credits

// this queries and extracts the fields
query url:
    Name, Passengers, Cost
    // can add headers, auth, etc here

display table:
    Name, Passengers, Cost
    sort Cost desc, Name, Passengers
    
// OR (for output)

for i, Name in Names:
    header(Name, color=yellow)
    print("P = {passengers[i]}, C = {costs[i]}"
    
// OR

rad url:
    Name, Passengers, Cost
```

- json fields block
- query block
- output block
- index from for-each loop
- indexing into json fields

---

These things are true:

1. We want the most common use-case to be minimal. Request - Extract - Display, in one motion. No repeating yourself.

```
rad url:
    Name, Passengers, Cost
    sort Cost desc, Name, Passengers
    header "Accept: application/json"
```

2. But, we want to support doing the query, performing some processing, and *then* displaying

```
apiKey = env("FX_API_KEY")
usdaud = json.fx[].usdaud

query exchangeUrl:
    usdaud
    header "Authentication: Bearer {apiKey}"

query url:
    Name, Passengers, usdCosts
    header "Accept: application/json"
    
AudCost float[]

for usdCost in usdCosts:
    AudCost += usdCost * usdaud
    
table:
    Name, Passengers, AudCost
    sort AudCost desc, Name, Passengers
```

- ^ suggests some more features:
  - floats
  - empty variable declaration
  - appending to array with +=
  - multiplication (implicitly, division)
  - table block
- I think we simply suggest the 'simple and terse' and the 'flexible but verbose' approaches.
  - We offer a block which does the most common use-case: query, extract, display in table
  - We then offer 'split' blocks that do only a *part* of those things, and may thus need some things repeated e.g. fields

---

Distilled down:

- env
    - yes, needs a little more thinking through, but doesn't need action -- functions already planned
- switch expression
    - ! yes, action, I think we should replace choice blocks with this
    - ! FIRST think about interpolation matching!
- if stmt in rad block
    - ! yes, action, want to ensure ebnf reflects this before implementing either
- fields keyword in rad block
  - ! yes, action. It's a minimal sacrifice towards:
    - being more self-descriptive
    - allowing fields to be defined *anywhere* in the rad block, including as part of if-statements
- json fields block
  - yes, but don't need action. I think there's a place for both types of json field definitions, so this can wait.
- query block
  - ! yes, create ebnf for it
- output block
  - ! yes, create ebnf for it. Let's call it 'table' though, for now.
- index from for-each loop
  - ! yes, create ebnf for it, before for loops get implemented
- indexing into json fields
  - ! yes, create ebnf for it
- floats
  - ! yes, create ebnf for it, add lexing/parsing
- empty variable declaration
  - ! yes, I don't see an alternative. Create ebnf for it before we're too far gone
- appending to array with +=
  - ! yes, but also support for other types e.g. strings, ints, floats.
- multiplication (implicitly, division)
  - ! yes, i thought i didn't need it, but on reflection, it'd be so frustrating to not be able to do this randomly

## 2024-08-16

```
args:
    city string? # The city to check the weather for.
    country string? # The country to check the weather for.
    
    at_least 1: city, country

baseUrl = "https://appleweather.com/api/weather"
    
queryParam = switch:
    case: "city={city}"
    case: "country={country}"
    
url = "{baseUrl}?{queryParam}"

rad url:
    ...
```

- `at_least` / `at_most` / `exactly` syntax
- switch stmts without discriminator

---

- What about choice resource replacements? Old:

```
// urls.radr
fruitUrls = resource choice:
  "https://apple.com" a
  "https://banana.com" b
...
args:
  fruit string

url = choice fruit from "urls" on fruitUrls
```

- New? :

```
// urls.radr
urls = resource switch:
  case "a": "https://apple.com"
  case "b": "https://banana.com"
...
args:
  fruit string

url = switch fruit on "urls" fruitUrls
```

---

- multi switch expressions

```
fruit, url = switch inputFruit:
    case "apple, "a": "apple", "https://apple.com" 
    case "banana", "b": "banana", "https://banana.com" 
```

## 2024-10-21

- More Python-like f-strings
  - require 'f' before quotations? do I really need that?
  - allow expressions e.g. `"Hello, {upper(name)}"`
  - also formatting `"{:.2f}"`
    - explore alternatives, both to f prefix and to this formatting, maybe newer language have nicer syntax
      - quick search suggests Python is maybe best 
- Allow invoking bash commands with `ls -l`.
  - Have syntax for failures? like amber?
  - Allow interpolation in it e.g. `cd {myDir}`
