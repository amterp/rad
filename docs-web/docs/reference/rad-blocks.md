---
title: Rad Blocks
---

## `rad` block

```rsl
rad url:
    fields Name, Birthdate, Height
    Name:
        map n -> truncate(n, 20)
    if sort_by_height:
        sort Height, Name, Birthdate
    else:
        sort
```

## `request` block

```rsl
request url:
    fields Name, Birthdate, Height
```

## `display` block

```rsl
display:
    fields Name, Birthdate, Height
```

## Colors

Valid colors:

`plain, black, red, green, yellow, blue, magenta, cyan, white, orange, pink`
