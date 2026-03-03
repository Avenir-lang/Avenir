# Methods

Methods are functions associated with a type. Avenir supports both instance methods and static methods.

## Instance Methods

Instance methods are functions with a receiver. The receiver is the instance on which the method is called.

### Declaration

Instance methods are declared using a receiver syntax:

```avenir
fun (self | Point).sum() | int {
    return self.x + self.y;
}
```

The receiver syntax is `(name | Type).methodName` where:
- `name` is the receiver variable name (conventionally `self`)
- `Type` is the receiver type (must be a struct type)
- `methodName` is the method name

### Calling Instance Methods

Instance methods are called using dot notation:

```avenir
var p | Point = Point{x = 3, y = 4};
var s | int = p.sum();
```

### Accessing Private Fields

Instance methods can access private fields of the receiver type:

```avenir
struct Point {
    x | int
    y | int
}

fun (self | Point).sum() | int {
    return self.x + self.y;  // OK: method can access private fields
}
```

### Mutating Methods

Instance methods can modify mutable fields:

```avenir
mut struct Point {
    x | int
    y | int
}

fun (self | Point).move(dx | int, dy | int) | void {
    self.x = self.x + dx;
    self.y = self.y + dy;
}
```

## Static Methods

Static methods are functions associated with a type but without a receiver instance.

### Declaration

Static methods are declared using a type name:

```avenir
fun Point.origin() | Point {
    return Point{x = 0, y = 0};
}
```

The receiver syntax is `Type.methodName` where:
- `Type` is the struct type name
- `methodName` is the method name

### Calling Static Methods

Static methods are called using dot notation on the type:

```avenir
var origin | Point = Point.origin();
```

Static methods do not have access to an instance, so they cannot access instance fields.

## Method Visibility

Methods follow the same visibility rules as functions:

- **Public methods**: Marked with `pub`, accessible from other modules
- **Private methods**: Without `pub`, accessible only within the same module

```avenir
pub struct Point {
    x | int
    y | int
}

pub fun (self | Point).distance() | float {
    // Public method
}

fun (self | Point).internal() | void {
    // Private method
}
```

## Method Overloading

Avenir does not support method overloading. Each method name must be unique within a type. Instance and static methods cannot share the same name.

## Methods vs Functions

Methods are essentially functions with a receiver. The receiver is the first parameter for instance methods, and not a parameter for static methods.

The following are equivalent:

```avenir
// Instance method
fun (self | Point).sum() | int {
    return self.x + self.y;
}

// Function (not a method)
fun point_sum(p | Point) | int {
    return p.x + p.y;
}
```

Methods provide a more natural syntax for object-oriented programming.

## Built-in Methods

Built-in methods on `list`, `string`, `bytes`, and `dict` are documented in
`docs/builtins.md`. These methods are implemented by the runtime and are
available without any imports.
