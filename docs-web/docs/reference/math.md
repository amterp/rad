---
title: Math
---

## Modulo

RSL has a modulo operator. It can be used to get the remainder after performing integer division.

```rsl
print(5 % 3)
```

<div class="result">
```
2
```
</div>

This also works for floats, or mixes between floats and ints.

```rsl
print(5.6 % 4.1)
print(5 % 4.5)
```

<div class="result">
```
1.5
0.5
```
</div>

Negative numbers is a somewhat complex topic, and different languages handle them differently. Here is a brief overview:

| Result of -11 % 7 | Approach           | Result takes sign of... | Example Languages                                      |
|-------------------|--------------------|-------------------------|--------------------------------------------------------|
| -4                | Truncated Division | Numerator (dividend)    | **RSL**, C, C++, Java, JavaScript, Go, Rust, Swift, C# |
| 3                 | Floored Division   | Denominator (divisor)   | Python, Ruby, R                                        |

Notice RSL behaves differently from Python, and instead follows the behavior of most other major languages.
