# Dictionaries (`dict`)

`dict` is Avenir's first-class generic dictionary type with the full form
`dict<K, V>` where `K` is the key type and `V` is the value type. For backward
compatibility, `dict<V>` is shorthand for `dict<string, V>`.

## Syntax

Dictionary literals use braces with `key: value` entries:

```avenir
var user | dict<any> = {
    name: "Alex",
    "age": 30,
    isAdmin: true,
};

// With explicit key type
var ids | dict<int, string> = {
    1001: "alice",
    1002: "bob",
};
```

Keys can be:

- identifiers (`name`, `isAdmin`) - only for string keys
- string literals (`"age"`, `'role'`) - only for string keys  
- expressions of type `K` - for generic key types

Trailing commas are optional.

## Types

Dictionary types are written as `dict<K, V>` where `K` is the key type and `V` is the value type:

```avenir
// String keys (shorthand form)
var scores | dict<int> = { alice: 10, bob: 12 };
var meta   | dict<any> = { env: "dev", retries: 3 };

// Explicit key type
var ids | dict<int, string> = { 1001: "alice", 1002: "bob" };
var matrix | dict<string, dict<int, float>> = {
    "row1": { 1: 1.1, 2: 1.2 },
    "row2": { 1: 2.1, 2: 2.2 },
};
```

All values in the literal must be assignable to `V`. Mixed value types require
an explicit union type:

```avenir
var mixed | dict<string, <int|string>> = { a: 1, b: "two" };
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

For a `dict<K, V>` (shorthand `dict<V>` uses `K = string`):

| Method | Parameters | Returns | Notes |
| --- | --- | --- | --- |
| `length()` | — | `int` | Number of entries |
| `keys()` | — | `list<K>` | Order not guaranteed |
| `values()` | — | `list<V>` | Order not guaranteed |
| `has(key)` | `K` | `bool` | Presence check |
| `get(key)` | `K` | `V?` | `none` if missing |
| `set(key, value)` | `K`, `V` | `void` | Mutates in place |
| `remove(key)` | `K` | `bool` | Returns whether key existed |

## Dicts vs Structs

Structs have a fixed schema and named fields; dictionaries are dynamic and
keyed by their key type `K`. Use structs for fixed, known shapes and dictionaries for
dynamic data or when you need non-string keys.

## Notes

- Dicts are backed by a hash map; iteration order is not guaranteed.
- `dict.set` mutates the dictionary in place.
- `dict<K, V>` is a built-in parametric type (not a user-defined generic type).
- For backward compatibility, `dict<V>` is equivalent to `dict<string, V>`.
