# std.json

`std.json` provides JSON parsing and serialization with a clean mapping to
native Avenir values.

## JSON Mapping

- JSON object → `dict<any>`
- JSON array → `list<any>`
- string → `string`
- number → `int` or `float` (integers become `int`, decimals/exponents become `float`)
- boolean → `bool`
- null → `none` (type `any?`)

## API

| Function | Parameters | Returns | Errors |
| --- | --- | --- | --- |
| `parse` | `text | string` | `any` | invalid JSON |
| `stringify` | `value | any` | `string` | unsupported values |

### Typed Helpers

| Function | Parameters | Returns | Errors |
| --- | --- | --- | --- |
| `asDict` | `value | any` | `dict<any>` | wrong type |
| `asList` | `value | any` | `list<any>` | wrong type |
| `asString` | `value | any` | `string` | wrong type |
| `asInt` | `value | any` | `int` | wrong type |
| `asBool` | `value | any` | `bool` | wrong type |

These throw a JSON error when the value does not match the expected type.

### Optional Lookups

Use built-in dict helpers for optional lookups:

```avenir
var value | any? = obj.get("key");
```

## Examples

### Parsing

```avenir
import std.json;

fun main() | void {
    try {
        var data | any = json.parse("{ \"name\": \"Alex\", \"age\": 30 }");
        var obj | dict<any> = json.asDict(data);
        print(obj.name);
        print(obj.age);
    } catch (e | error) {
        print("JSON error: " + errorMessage(e));
    }
}
```

### Stringify

```avenir
import std.json;

fun main() | void {
    var user | dict<any> = { name: "Alex", age: 30, isAdmin: false };
    var text | string = json.stringify(user);
    print(text);
}
```

## Error Handling

All parse and stringify errors are runtime errors and can be caught with
`try / catch`.

### Error Helpers

`std.json.errors` exposes small helpers used by the std library:

| Function | Parameters | Returns | Notes |
| --- | --- | --- | --- |
| `typeError` | `expected | string`, `actual | string` | `error` | Type mismatch helper |
| `keyError` | `key | string` | `error` | Missing key helper |

## Limitations

- Unsupported values (closures, structs, bytes, errors, etc.) cause stringify
  errors.
- Dictionary output uses deterministic key ordering, not insertion order.
