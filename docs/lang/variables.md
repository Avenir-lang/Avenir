# Variables

Variables in Avenir must be explicitly declared with a type.

## Variable Declaration

Variables are declared using the `var` keyword:

```avenir
var name | type = value;
```

The type annotation is required. The initial value is also required.

Example:

```avenir
var x | int = 10;
var name | string = "Avenir";
var is_ready | bool = true;
```

## Variable Assignment

Variables can be reassigned using the assignment operator:

```avenir
var x | int = 10;
x = 20;
```

The assigned value must be type-compatible with the variable's type.

## Scope

Variables are scoped to the block in which they are declared:

```avenir
{
    var x | int = 10;
    // x is available here
}
// x is not available here
```

Function parameters are scoped to the function body.

## Variable Visibility

Variables are always private to their scope. There is no way to export variables from a module.

## Type Requirements

All variables must have an explicit type annotation. Type inference is not supported.

Example:

```avenir
var x | int = 10;        // OK
var y = 10;              // Error: type annotation required
```
