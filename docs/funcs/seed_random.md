# seed_random

Seeds the random number generator used by [`rand`](#rand) and [`rand_int`](#rand_int).

## Signature

`seed_random(_seed: int) -> void`

## Examples

```rad
seed_random(42)
rand()              // -> Same sequence every time with seed 42
rand_int(10)        // -> Same sequence every time with seed 42
```

## Category

random
