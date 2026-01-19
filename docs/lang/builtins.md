# Built-in Functions and Methods

Avenir provides built-in functions and methods for common operations.
Builtins throw runtime errors when they fail; use `try / catch` to handle
errors from I/O and conversions.

## Built-in Functions

| Function | Parameters | Returns | Errors |
| --- | --- | --- | --- |
| `print` | `value | any` | `any` | I/O failures |
| `input` | — | `string` | I/O failures |
| `len` | `value | list<any> \| bytes` | `int` | non-list/bytes |
| `typeOf` | `value | any` | `string` | invalid runtime value |
| `toInt` | `value | string` | `int` | invalid integer |
| `error` | `message | string` | `error` | — |
| `errorMessage` | `e | error` | `string` | — |
| `fromString` | `s | string` | `bytes` | — |

### `print(value | any) | any`

Prints a value to standard output and returns the same value.

```avenir
print("Hello, World!");
print(42);
```

### `input() | string`

Reads a line from standard input. Leading/trailing newline characters are trimmed.
Returns an empty string on EOF.

```avenir
var name | string = input();
```

### `len(value | any) | int`

Returns the length of a list or bytes value. Throws a runtime error if the
argument is not a list or bytes.

```avenir
var length | int = len([1, 2, 3]);  // Returns 3
var size | int = len(b"bytes");     // Returns 5
```

### `typeOf(value | any) | string`

Returns the canonical Avenir type name of a runtime value.
For interface-typed values, `typeOf()` returns the concrete runtime type.
Lists are rendered as `list<T...>`, and optionals are rendered as `T?` (or `any?` when the inner type is unknown).

```avenir
typeOf(10);            // "int"
typeOf(1.5);           // "float"
typeOf("hello");       // "string"
typeOf(true);          // "bool"
typeOf([]);            // "list<any>"
typeOf([1, 2]);        // "list<int>"
typeOf(fromString("abc")); // "bytes"
typeOf(error("x"));    // "error"
typeOf(Point{x = 1, y = 2}); // "Point"
typeOf(some(1));       // "int?"
typeOf(none);          // "any?"
```

### `toInt(value | string) | int`

Converts a decimal string (e.g. `"123"`, `"-42"`) to an integer.
Throws a runtime `error` if the string is not a valid integer.

```avenir
var ok | int = toInt("123");
try {
    var bad | int = toInt("hello");
} catch (e | error) {
    print(errorMessage(e));
}
```

### `error(message | string) | error`

Creates an error value.

```avenir
var e | error = error("something went wrong");
```

### `errorMessage(e | error) | string`

Extracts the message from an error value.

```avenir
var msg | string = errorMessage(e);
```

### `fromString(s | string) | bytes`

Converts a string to bytes.

```avenir
var data | bytes = fromString("hello");
```

## List Methods

Lists have the following methods:

> Note: list methods return new lists rather than mutating in place, unless
> otherwise noted.

| Method | Parameters | Returns | Notes |
| --- | --- | --- | --- |
| `append` | `element | any` | `list<any>` | Returns new list |
| `length` | — | `int` | Element count |
| `get` | `index | int` | `any` | Throws on out-of-bounds |
| `contains` | `element | any` | `bool` | Deep equality |
| `indexOf` | `element | any` | `int` | `-1` if not found |
| `slice` | `start | int`, `end | int` | `list<any>` | End exclusive |
| `reverse` | — | `list<any>` | Returns new list |
| `copy` | — | `list<any>` | Shallow copy |
| `pop` | — | `any` | Returns last element |
| `insert` | `index | int`, `element | any` | `list<any>` | Returns new list |
| `removeAt` | `index | int` | `list<any>` | Returns new list |
| `clear` | — | `list<any>` | Empty list |
| `isEmpty` | — | `bool` | — |
| `map` | `fn | fun(any) | any` | `list<any>` | Calls function per element |
| `filter` | `fn | fun(any) | bool` | `list<any>` | Calls predicate per element |
| `reduce` | `initial | any`, `reducer | fun(any, any) | any` | `any` | Accumulator |

### `append(element | any) | list<any>`

Appends an element to the list, returning a new list.

```avenir
var list | list<int> = [1, 2];
list = list.append(3);  // [1, 2, 3]
```

### `length() | int`

Returns the length of the list.

```avenir
var len | int = list.length();
```

### `get(index | int) | any`

Returns the element at the given index. Throws if the index is out of bounds.

```avenir
var item | int = list.get(0);
```

### `contains(element | any) | bool`

Checks if the list contains an element.

```avenir
var found | bool = list.contains(42);
```

### `indexOf(element | any) | int`

Returns the index of the first occurrence of an element, or -1 if not found.

```avenir
var idx | int = list.indexOf(42);
```

### `slice(start | int, end | int) | list<any>`

Returns a slice of the list from `start` (inclusive) to `end` (exclusive).

```avenir
var sub | list<int> = list.slice(1, 3);
```

### `reverse() | list<any>`

Returns a reversed copy of the list.

```avenir
var reversed | list<int> = list.reverse();
```

### `copy() | list<any>`

Returns a copy of the list.

```avenir
var copied | list<int> = list.copy();
```

### `pop() | any`

Returns the last element of the list. Throws on empty lists.

```avenir
var last | int = list.pop();
```

### `insert(index | int, element | any) | list<any>`

Inserts an element at the given index, returning a new list. Throws if the
index is out of bounds.

```avenir
list = list.insert(1, 42);
```

### `removeAt(index | int) | list<any>`

Removes the element at the given index, returning a new list. Throws if the
index is out of bounds.

```avenir
list = list.removeAt(1);
```

### `clear() | list<any>`

Returns an empty list of the same type.

```avenir
list = list.clear();
```

### `isEmpty() | bool`

Checks if the list is empty.

```avenir
var empty | bool = list.isEmpty();
```

### `map(fn | fun (any) | any) | list<any>`

Applies a function to each element, returning a new list.
Throws a runtime error if `fn` is not a function or if the function throws.

```avenir
var doubled | list<int> = list.map(fun (x | int) | int {
    return x * 2;
});
```

### `filter(fn | fun (any) | bool) | list<any>`

Filters elements using a predicate function, returning a new list.
Throws a runtime error if `fn` is not a function or if the function throws.

```avenir
var evens | list<int> = list.filter(fun (x | int) | bool {
    return x % 2 == 0;
});
```

### `reduce(initial | any, reducer | fun (any, any) | any) | any`

Reduces the list to a single value using a reducer function.
Throws a runtime error if `reducer` is not a function or if the function throws.

```avenir
var sum | int = list.reduce(0, fun (acc | int, x | int) | int {
    return acc + x;
});
```

## String Methods

Strings have the following methods:

| Method | Parameters | Returns | Notes |
| --- | --- | --- | --- |
| `length` | — | `int` | Byte length (UTF-8) |
| `toUpperCase` | — | `string` | — |
| `toLowerCase` | — | `string` | — |
| `trim` | — | `string` | Removes leading/trailing whitespace |
| `trimLeft` | — | `string` | Removes leading whitespace |
| `trimRight` | — | `string` | Removes trailing whitespace |
| `contains` | `substr | string` | `bool` | — |
| `startsWith` | `prefix | string` | `bool` | — |
| `endsWith` | `suffix | string` | `bool` | — |
| `replace` | `old | string`, `new | string` | `string` | Replaces all occurrences |
| `split` | `sep | string` | `list<string>` | — |
| `indexOf` | `substr | string` | `int` | `-1` if not found |
| `lastIndexOf` | `substr | string` | `int` | `-1` if not found |

### `length() | int`

Returns the byte length of the string (UTF-8 byte count).

```avenir
var len | int = str.length();
```

### `toUpperCase() | string`

Returns an uppercase copy of the string.

```avenir
var upper | string = str.toUpperCase();
```

### `toLowerCase() | string`

Returns a lowercase copy of the string.

```avenir
var lower | string = str.toLowerCase();
```

### `trim() | string`

Returns a copy with leading and trailing whitespace removed.

```avenir
var trimmed | string = str.trim();
```

### `trimLeft() | string`

Returns a copy with leading whitespace removed.

```avenir
var trimmed | string = str.trimLeft();
```

### `trimRight() | string`

Returns a copy with trailing whitespace removed.

```avenir
var trimmed | string = str.trimRight();
```

### `contains(substr | string) | bool`

Checks if the string contains a substring.

```avenir
var found | bool = str.contains("hello");
```

### `startsWith(prefix | string) | bool`

Checks if the string starts with a prefix.

```avenir
var matches | bool = str.startsWith("http");
```

### `endsWith(suffix | string) | bool`

Checks if the string ends with a suffix.

```avenir
var matches | bool = str.endsWith(".av");
```

### `replace(old | string, new | string) | string`

Replaces all occurrences of `old` with `new`.

```avenir
var replaced | string = str.replace("old", "new");
```

### `split(sep | string) | list<string>`

Splits the string by a separator, returning a list of strings.

```avenir
var parts | list<string> = str.split(",");
```

### `indexOf(substr | string) | int`

Returns the index of the first occurrence of a substring, or -1 if not found.

```avenir
var idx | int = str.indexOf("hello");
```

### `lastIndexOf(substr | string) | int`

Returns the index of the last occurrence of a substring, or -1 if not found.

```avenir
var idx | int = str.lastIndexOf("hello");
```

## Bytes Methods

Bytes have the following methods:

| Method | Parameters | Returns | Notes |
| --- | --- | --- | --- |
| `length` | — | `int` | Byte count |
| `append` | `b | bytes` | `bytes` | Returns new bytes |
| `concat` | `b | bytes` | `bytes` | Returns new bytes |
| `slice` | `start | int`, `end | int` | `bytes` | End exclusive |
| `toString` | — | `string` | UTF-8 decode, errors on invalid |

### `length() | int`

Returns the length of the bytes value.

```avenir
var len | int = data.length();
```

### `append(b | bytes) | bytes`

Appends bytes to the value, returning a new bytes value.

```avenir
data = data.append(b"more");
```

### `concat(b | bytes) | bytes`

Concatenates two bytes values, returning a new bytes value.

```avenir
var combined | bytes = data.concat(b"more");
```

### `slice(start | int, end | int) | bytes`

Returns a slice of bytes from `start` (inclusive) to `end` (exclusive).

```avenir
var sub | bytes = data.slice(0, 10);
```

### `toString() | string`

Converts bytes to a UTF-8 string. Throws if the bytes are not valid UTF-8.

```avenir
var str | string = data.toString();
```

## Dict Methods

Dicts have the following methods:

| Method | Parameters | Returns | Notes |
| --- | --- | --- | --- |
| `length` | — | `int` | Entry count |
| `keys` | — | `list<string>` | Order not guaranteed |
| `values` | — | `list<any>` | Order not guaranteed |
| `has` | `key | string` | `bool` | Presence check |
| `get` | `key | string` | `any?` | `none` if missing |
| `set` | `key | string`, `value | any` | `void` | Mutates in place |
| `remove` | `key | string` | `bool` | Returns whether key existed |

### `length() | int`

Returns the number of entries in the dictionary.

### `keys() | list<string>`

Returns the dictionary keys. Order is not guaranteed.

### `values() | list<any>`

Returns the dictionary values. Order is not guaranteed.

### `has(key | string) | bool`

Checks if a key is present in the dictionary.

### `get(key | string) | any?`

Returns the value as an optional. Returns `none` if the key is missing.

### `set(key | string, value | any) | void`

Inserts or updates a key. Mutates the dictionary in place.

### `remove(key | string) | bool`

Removes a key and returns whether it existed.
