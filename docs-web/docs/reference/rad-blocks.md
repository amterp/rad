---
title: Rad Blocks
---

## `rad` block

```rad
rad url:
    fields Name, Birthdate, Height
    Name:
        map fn(n) truncate(n, 20)
    if sort_by_height:
        sort Height, Name, Birthdate
    else:
        sort
```

## `request` block

```rad
request url:
    fields Name, Birthdate, Height
```

## `display` block

```rad
display:
    fields Name, Birthdate, Height
```

## Colors

Valid colors:

`plain, black, red, green, yellow, blue, magenta, cyan, white, orange, pink`
