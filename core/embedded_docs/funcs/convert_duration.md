# convert_duration

Converts a numeric value in the specified unit. Supports `nanos`, `micros`, `millis`, `seconds`, `minutes`, `hours`, and
`days`.

```rad
convert_duration(_value: int|float, _unit: ["nanos", "micros", "millis", "seconds", "minutes", "hours", "days"]) -> error|{ "nanos": int, "micros": float, "millis": float, "seconds": float, "minutes": float, "hours": float, "days": float }
```

```rad
convert_duration(90, "seconds").minutes   // -> 1.5
convert_duration(1, "days").hours         // -> 24.0
convert_duration(1.5, "hours").minutes    // -> 90.0
convert_duration(1500, "millis").seconds  // -> 1.5
```
