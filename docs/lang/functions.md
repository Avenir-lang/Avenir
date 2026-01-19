# Functions

Functions in Avenir are first-class values. Functions can be named or anonymous (function literals).

## Function Declaration

Functions are declared using the `fun` keyword:

```avenir
fun name(param1 | type1, param2 | type2) | returnType {
    // body
}
```

The return type is required. Use `void` if the function doesn't return a value.

Example:

```avenir
fun add(x | int, y | int) | int {
    return x + y;
}

fun print_hello() | void {
    print("Hello");
}
```

## Public Functions

Functions can be marked as public using the `pub` keyword:

```avenir
pub fun exported_function() | void {
    // This function is accessible from other modules
}
```

Private functions (without `pub`) are only accessible within the same module.

## Parameters

Function parameters must have explicit type annotations:

```avenir
fun greet(name | string, age | int) | void {
    print(name);
    print(age);
}
```

### Default Parameters

Parameters can have default values:

```avenir
fun greet(name | string, greeting | string = "Hello") | void {
    print(greeting);
    print(name);
}
```

Default values must be compile-time constants.

When calling a function with default parameters, you can omit arguments:

```avenir
greet("Alice");              // greeting defaults to "Hello"
greet("Bob", "Hi");          // greeting is "Hi"
```

### Named Arguments

Function calls support named arguments:

```avenir
fun create_point(x | int, y | int) | Point {
    return Point{x = x, y = y};
}

create_point(x = 10, y = 20);
create_point(y = 20, x = 10);  // Order doesn't matter
```

Named arguments can be mixed with positional arguments, but positional arguments must come before named arguments:

```avenir
create_point(10, y = 20);      // OK
create_point(x = 10, 20);      // Error: positional after named
```

## Return Values

Functions return values using the `return` statement:

```avenir
fun add(x | int, y | int) | int {
    return x + y;
}
```

Functions with `void` return type can use `return` without a value:

```avenir
fun early_exit() | void {
    if (condition) {
        return;
    }
    // ...
}
```

## Function Literals

Function literals (anonymous functions) are supported:

```avenir
var add | fun (int, int) | int = fun (x | int, y | int) | int {
    return x + y;
};
```

Function literals can be passed as arguments and stored in variables.

## Closures

Function literals can capture variables from their enclosing scope (closures):

```avenir
fun make_counter() | fun () | int {
    var count | int = 0;
    return fun () | int {
        count = count + 1;
        return count;
    };
}
```

Closures capture variables by reference, allowing them to modify captured variables.

## Function Types

Function types are specified using the `fun` keyword:

```avenir
var fn | fun (int, int) | int = add;
```

Function types use structural typing: two function types are equal if they have the same parameter types and return type.

## Function Calls

Functions are called using parentheses:

```avenir
var result | int = add(10, 20);
```

Method calls use dot notation:

```avenir
var length | int = str.length();
```

## Main Function

Every program must have a `main` function that serves as the entry point:

```avenir
fun main() | void {
    // Program entry point
}
```

The `main` function is called automatically when the program runs.
