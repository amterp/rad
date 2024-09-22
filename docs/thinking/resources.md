# Resources

## Pick

```
// options is an array of values
// this launches a 'picking' mode in CLI where the user narrows options and selects one
out = pick("Pick a website", options)
```

```
// here, 'filter' is a value used for picking from options. It's almost like 'prefilling' what the user might enter
// in the 1-ary case. Although, it should probably just narrow down all the options completely and still start the user
// with no input. fuzzy filter.
out = pick("Pick a website", options, filter)
```

## Json Approach

Resource file

`./website.json`
```json
{
  "options": [
    {
      "return": ["gitlab.com", "GitLab"],
      "match": ["gl", "lab"]
    },
    {
      "return": ["github.com", "github"],
      "match": ["gh", "hub"]
    }
  ]
}
```

Using in RSL

```
// Usually it expects a string array. However, if given a single string, treat as resource to look up?
out = pick("website.json", filter, prompt="Pick a website")
```
