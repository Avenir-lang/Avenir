# Structs

Structs are user-defined types with named fields. Structs use nominal typing: two structs are equal only if they have the same name.

## Struct Declaration

Structs are declared using the `struct` keyword:

```avenir
struct Point {
    x | int
    y | int
}
```

## Public Structs

Structs can be marked as public using the `pub` keyword:

```avenir
pub struct Point {
    x | int
    y | int
}
```

Public structs are accessible from other modules. Private structs (without `pub`) are only accessible within the same module.

## Mutable Structs

Structs can be marked as mutable using the `mut` keyword:

```avenir
mut struct Point {
    x | int
    y | int
}
```

Mutable structs allow field assignment. Immutable structs (without `mut`) do not allow field assignment.

## Fields

### Field Declaration

Fields are declared with a name and type:

```avenir
struct Point {
    x | int
    y | int
}
```

### Public Fields

Fields can be marked as public using the `pub` keyword:

```avenir
pub struct Point {
    pub x | int
    pub y | int
}
```

Public fields are accessible from other modules. Private fields (without `pub`) are only accessible within the same module or from methods of the same struct.

**Note**: Public fields can only be declared in public structs. Private structs cannot have public fields.

### Mutable Fields

Fields can be marked as mutable using the `mut` keyword:

```avenir
struct Point {
    mut x | int
    y | int
}
```

Mutable fields allow assignment even in immutable structs. Field-level `mut` overrides the struct-level mutability.

### Field Mutability Rules

Field assignment is allowed only if the field is mutable. Field mutability is computed as follows:

1. If the struct is `mut struct`, all fields are mutable by default
2. If a field is marked with `mut`, it is mutable regardless of struct mutability
3. Otherwise, the field is immutable

Examples:

```avenir
// Immutable struct, immutable field
struct Point {
    x | int
}
// Point{x = 10}.x = 20;  // Error: cannot assign to immutable field

// Immutable struct, mutable field
struct Point {
    mut x | int
}
// Point{x = 10}.x = 20;  // OK

// Mutable struct, immutable field
mut struct Point {
    x | int
}
// Point{x = 10}.x = 20;  // OK (struct is mutable)

// Mutable struct, mutable field
mut struct Point {
    mut x | int
}
// Point{x = 10}.x = 20;  // OK
```

### Default Values

Fields can have default values:

```avenir
struct Config {
    host | string = "localhost"
    port | int = 8080
}
```

Default values must be compile-time constants. Fields with default values can be omitted in struct literals.

## Struct Literals

Struct literals create struct values:

```avenir
var p | Point = Point{x = 10, y = 20};
```

Fields are initialized using named initialization: `field = value`.

### Partial Initialization

Fields with default values can be omitted:

```avenir
struct Config {
    host | string = "localhost"
    port | int = 8080
}

var c | Config = Config{port = 9000};  // host uses default "localhost"
```

Fields without default values must be provided.

### Field Assignment

For mutable structs or mutable fields, fields can be assigned after initialization:

```avenir
mut struct Point {
    x | int
    y | int
}

var p | Point = Point{x = 10, y = 20};
p.x = 30;
p.y = 40;
```

Assignment is performed in-place (no copying).

## Field Access

Fields are accessed using dot notation:

```avenir
var p | Point = Point{x = 10, y = 20};
var x | int = p.x;
var y | int = p.y;
```

## Type System

Structs use nominal typing: two structs are equal only if they have the same name, regardless of field structure.

```avenir
struct Point {
    x | int
    y | int
}

struct Coordinate {
    x | int
    y | int
}

// Point and Coordinate are different types, even though they have the same fields
```
