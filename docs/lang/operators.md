# Operators

This document describes Avenir operators and their typing rules.

## `+` (addition and string concatenation)

`+` is resolved by the type checker and supports exactly two forms:

- Numeric addition: `int + int`, `float + float`, or mixed `int`/`float` (result is `float`)
- String concatenation: `string + string` (result is `string`)

No implicit conversions are performed. In particular:

- `string + int` is a compile-time error
- `int + string` is a compile-time error

Examples:

```avenir
var a | string = "hello " + "world";
var n | int = 1 + 2;
var f | float = 1 + 2.5;
```

If you need to format non-strings, use string interpolation:

```avenir
print("sum = ${1 + 2}");
```
