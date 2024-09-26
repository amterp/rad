# Json Field Syntax

## Basics

```
// extracts as a non-array
a = json.id
```

```
// extracts as an array
a = json[].id
```

```
// extracts as a non-array
a = json[0].id
a = json.ids[1]
```

---

```json
{
  "ids": [1, 2, 3]
}
```

```
// extracts as a non-array, but the value itself is an array, so this leaves 'a' as an array
a = json.ids
```

```
// what would this do?
a = json.ids[]
// where the previous one extracts the 'array' as a 'non-array' and it therefore ends up as an array, maybe this one
// extracts the individual elements as non-arrays into an array, and so the end-result is the same.
```

---

```json
[
  {
    "name": "alice",
    "age": 35
  },
  {
    "name": "bob",
    "age": 25
  }
]
```

## Blob Extraction

```
// extracts 'a' as a string containing the json blob, potentially?
a = json

rad url:
    fields a
```

```
// extracts 'a' as a string '{ "name": "alice", "age": 35 }'
a = json[0]
```

## Request-Display

```
// extract blob
a = json
request url:
    fields a
    
// pass blob to 'display' block, extract name
b = json[].name
display a:
    fields b
```

```
// COULD'VE done this instead
b = json[].name
request url:
    fields b

// then display, i.e. json blob optional
display:
    fields b
```

## Len

```json
[
  {
    "name": "alice",
    "ids": [1, 2]
  },
  {
    "name": "bob",
    "ids": [3, 4, 5]
  }
]
```

```
// syntax tbd? but this would extract lengths of IDs so a = [2, 3]
a = json[].ids[LEN]

// since this is not an array, it returns a non-array value '2'
b = json[LEN]

// OR MAYBE WE DON'T DO THAT, instead
c = json[].name
d = json[].ids
request url:
    fields c, d
// now d = [[1, 2], [3, 4, 5]], i.e. array of arrays
lens int[] = []
for ids in d:
    lens += len(ids)
display:
    fields c, lens
    
// seems a like little bit of a shame to manually need to extract lengths? could offer utility
c = json[].name
d = json[].ids
request url:
    fields c, d
// honestly comprehensions like this would be pretty dope
lens = [len(l) for l in d]
// but otherwise (for example):
lens = lengths(d)
```

## Keys

```json
{
  "alice": {
    "age": 35,
    "surname": "abba"
  },
  "bob": {
    "age": 25,
    "surname": "billy"
  }
}
```

```
// Matches all keys there, so a = [alice, bob] 
a = json.*

// b = [35, 25]
b = json.*.age

// again syntax tbd, but c = 
c = json
d = flatten_keys(c) // d = ["alice.age", "alice.surname", "bob.age", "bob.surname"]
e = flatten_values(c) // e = [35, "abba", 25, "billy"] // todo this will require multi-type arrays
f, g = flatten(c) // f = d AND g = e 
```

## Conclusions

### 2024-09-26

- Add json field array index extraction support e.g. `json.ids[0]`, `json[1].name`
- Add list comprehensions
- Really consider array of arrays
- Really consider multi-type arrays
- LEN solved by list comprehensions and array of arrays
- FLATTEN solved by multi-type arrays
- Allow storing/extracting json blobs as strings
- 'request' seems like a sensible block to add
  - but is 'display' not the same as 'rad' but allowing json blob in place of url? why not just allow 'rad' that?
