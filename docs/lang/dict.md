# Dictionaries (`dict`)

`dict` is Avenir's first-class dictionary type. Keys are always `string`, and
values are statically typed.

## Syntax

Dictionary literals use braces with `key: value` entries:

```avenir
var user | dict<any> = {
    name: "Alex",
    "age": 30,
    isAdmin: true,
};
```

Keys can be:

- identifiers (`name`, `isAdmin`)
- string literals (`"age"`, `'role'`)

Trailing commas are optional.

## Types

Dictionary types are written as `dict<T>` where `T` is the value type:

```avenir
var scores | dict<int> = { alice: 10, bob: 12 };
var meta   | dict<any> = { env: "dev", retries: 3 };
```

All values in the literal must be assignable to `T`. Mixed value types require
an explicit union type:

```avenir
var mixed | dict< <int|string> > = { a: 1, b: "two" };
```

## Access

Dot access reads a key:

```avenir
print(user.name);
```

Index access reads a key dynamically:

```avenir
print(user["age"]);
```

Missing keys with index access throw a runtime error. Use `dict.get()` when a
key may be missing.

## Built-in Methods

| Method | Parameters | Returns | Notes |
| --- | --- | --- | --- |
| `length()` | — | `int` | Number of entries |
| `keys()` | — | `list<string>` | Order not guaranteed |
| `values()` | — | `list<T>` | Order not guaranteed |
| `has(key)` | `string` | `bool` | Presence check |
| `get(key)` | `string` | `T?` | `none` if missing |
| `set(key, value)` | `string`, `T` | `void` | Mutates in place |
| `remove(key)` | `string` | `bool` | Returns whether key existed |

## Dicts vs Structs

Structs have a fixed schema and named fields; dictionaries are dynamic and
keyed by strings. Use structs for fixed, known shapes and dictionaries for
dynamic data.

## Notes

- Dicts are backed by a hash map; iteration order is not guaranteed.
- `dict.set` mutates the dictionary in place.
