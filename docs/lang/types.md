# Types

Avenir is a statically-typed language. All types are checked at compile time.

## Primitive Types

### `int`

64-bit signed integer.

```avenir
var x | int = 42;
```

### `float`

64-bit floating-point number.

```avenir
var x | float = 3.14;
```

### `string`

String of characters.

```avenir
var s | string = "Hello";
```

### `bool`

Boolean value (`true` or `false`).

```avenir
var b | bool = true;
```

### `bytes`

Byte sequence.

```avenir
var data | bytes = b"bytes";
```

### `void`

Represents no return value. Used for functions that don't return a value.

```avenir
fun print_hello() | void {
    print("Hello");
}
```

### `any`

Represents any type. Used for generic operations.

```avenir
var x | any = 42;
```

### `error`

Error type for exception handling.

```avenir
var e | error = error("something went wrong");
```

## Interfaces

Interfaces define a set of method signatures that types must implement. Avenir uses **structural typing**: a type automatically satisfies an interface if it has all required methods with matching signatures.

### Interface Declaration

Interfaces are declared using the `interface` keyword:

```avenir
interface Stringer {
    fun toString() | string
}
```

Interfaces can be public:

```avenir
pub interface Writer {
    fun write(data | string) | void
    fun flush() | void
}
```

### Structural Typing

A type satisfies an interface **implicitly** if it has all required methods. There is no `implements` keyword:

```avenir
interface Stringer {
    fun toString() | string
}

struct Point {
    x | int
    y | int
}

fun (p | Point).toString() | string {
    return "Point";
}

fun main() | void {
    var s | Stringer = Point{x = 1, y = 2};
    var str | string = s.toString();
}
```

### Interface Methods

Interface methods are declared with their full signature:

```avenir
interface Writer {
    fun write(data | string) | void
    fun flush() | void
}
```

Method signatures must match exactly:
- Method name must match
- Parameter count and types must match
- Return type must match

### Using Interfaces as Types

Interfaces can be used as types:

```avenir
fun printValue(x | Stringer) | void {
    print(x.toString());
}
```

### Method Compatibility

A method satisfies an interface method if:
- Name matches exactly
- Parameter count and order match
- Parameter types match exactly
- Return type matches exactly

Only **instance methods** can satisfy interfaces. Static methods do not satisfy interfaces.

### Built-in Type Satisfaction

Built-in types can satisfy interfaces if they have the required methods:

```avenir
interface Length {
    fun length() | int
}

fun main() | void {
    var l | Length = "hello";  // string has length() method
    var len | int = l.length();
}
```

### Visibility Rules

- Public interfaces can only require public methods
- Private methods cannot satisfy public interfaces across modules
- Within the same module, private methods can satisfy public interfaces

### Error Messages

When a type does not satisfy an interface, the compiler provides detailed error messages:

```
type Point does not satisfy interface Stringer: missing methods: toString() | string
```

or

```
type Point does not satisfy interface Stringer: incompatible method signatures: toString: expected () | string, got () | int
```

## Composite Types

### Lists

Lists are sequences of values. Lists can have homogeneous or heterogeneous element types:

```avenir
var numbers | list<int> = [1, 2, 3];
var mixed | list<int, string> = [1, "two", 3];
```

Lists use structural typing: two lists are equal if they have the same element types.

### Dicts

Dicts map string keys to values and are written as `dict<T>` where `T` is the
value type:

```avenir
var user | dict<any> = { name: "Alex", "age": 30 };
var scores | dict<int> = { alice: 10, bob: 12 };
```

Keys are always strings; values must be assignable to `T`.
Use `dict.get()` when a key may be missing; it returns an optional `T?`.

### Structs

User-defined types with named fields:

```avenir
struct Point {
    x | int
    y | int
}
```

Structs use nominal typing: two structs are equal only if they have the same name.

### Functions

Function types specify parameter and return types:

```avenir
fun (int) | int    // Function taking int, returning int
fun (int, string) | bool    // Function taking int and string, returning bool
```

### Optional Types

Optional types allow nullable values:

```avenir
var x | int? = some(42);
var y | int? = none;
```

Optional types are written as `T?` where `T` is the base type.
`none` is the only null-like value in the language.

### Union Types

Union types allow values that can be one of several types:

```avenir
var value | <int|string> = 42;
value = "hello";
```

Union types use angle brackets: `<T1|T2|...>`.

## Type System

### Structural vs Nominal Typing

- **Lists**: Structural typing (element types must match)
- **Structs**: Nominal typing (struct name must match)
- **Functions**: Structural typing (signature must match)

### Type Checking

All types are checked at compile time. Type errors are reported during compilation.

### Type Assignability

A value of type `T` can be assigned to a variable of type `U` if:

1. `T` and `U` are the same type
2. `T` is assignable to `U` (e.g., `int` can be assigned to `float` in some contexts)
3. `T` is a subtype of `U` (if applicable)

## Type Annotations

Types are specified using the pipe (`|`) separator:

```avenir
var name | type = value;
```

For function parameters and return types:

```avenir
fun name(param1 | type1, param2 | type2) | returnType {
    // ...
}
```

## Type Inference

Avenir requires explicit type annotations. Type inference is not currently supported.
