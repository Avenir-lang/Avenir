# Dictionaries (`dict`)

`dict` is Avenir's first-class generic dictionary type with the full form
`dict<K, V>` where `K` is the key type and `V` is the value type.

The shorthand `dict<V>` is equivalent to `dict<string, V>` (string keys by
default).

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

Dictionary types can be written in two forms:

**Short form** — `dict<V>` (key type defaults to `string`):

```avenir
var scores | dict<int> = { alice: 10, bob: 12 };
var meta   | dict<any> = { env: "dev", retries: 3 };
```

**Full form** — `dict<K, V>` (explicit key and value types):

```avenir
var scores | dict<string, int> = { math: 95, science: 88 };
```

All values in the literal must be assignable to `V`. Mixed value types require
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

For a `dict<K, V>`:

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
keyed by strings. Use structs for fixed, known shapes and dictionaries for
dynamic data.

## Notes

- Dicts are backed by a hash map; iteration order is not guaranteed.
- `dict.set` mutates the dictionary in place.
- The runtime currently supports only `string` keys. The `K` type parameter
  enables future key type expansion without syntax changes.
