# Control Flow

Avenir provides several control flow constructs: conditionals, loops, and exception handling.

## Conditionals

### If Statements

If statements execute code conditionally:

```avenir
if (condition) {
    // code
}
```

The condition must be of type `bool`.

### If-Else Statements

Else clauses provide alternative execution paths:

```avenir
if (condition) {
    // code
} else {
    // alternative code
}
```

### Else-If Chaining

Multiple conditions can be chained using else-if:

```avenir
if (condition1) {
    // code
} else if (condition2) {
    // code
} else {
    // default code
}
```

### Multiple Conditions

If statements support multiple conditions using semicolons:

```avenir
if (x > 0; x < 10) {
    // equivalent to: if (x > 0 && x < 10)
}
```

## Loops

### While Loops

While loops execute code while a condition is true:

```avenir
while (condition) {
    // code
}
```

The condition must be of type `bool`.

### For Loops

For loops provide C-style iteration:

```avenir
for (init; condition; post) {
    // code
}
```

All parts (init, condition, post) are optional:

```avenir
for (; condition; ) { }    // while loop equivalent
for (;;) { }               // infinite loop
```

The init can be a variable declaration or an assignment:

```avenir
for (var i | int = 0; i < 10; i = i + 1) {
    // code
}
```

### For-Each Loops

For-each loops iterate over lists:

```avenir
for (item in list) {
    // code
}
```

The loop variable `item` is scoped to the loop body. The list expression must be of type `list<T>`.

Example:

```avenir
var numbers | list<int> = [1, 2, 3];
for (n in numbers) {
    print(n);
}
```

### Break Statements

Break statements exit the innermost loop:

```avenir
while (true) {
    if (condition) {
        break;
    }
}
```

Break can only be used inside loops.

## Exception Handling

### Try-Catch Statements

Try-catch statements handle exceptions:

```avenir
try {
    // code that may throw
} catch (e | error) {
    // handle exception
}
```

The catch variable must be of type `error`. The catch block is executed if an exception is thrown in the try block, including runtime and builtin errors.

### Throw Statements

Throw statements raise exceptions:

```avenir
throw error("something went wrong");
```

The expression must be of type `error`.

### Exception Propagation

If an exception is not caught, it propagates up the call stack. If it reaches the top level, the program terminates with an error.

## Block Statements

Blocks create new scopes:

```avenir
{
    var x | int = 10;
    // x is available here
}
// x is not available here
```

Blocks are used in conditionals, loops, and functions.

## Expression Statements

Expressions can be used as statements:

```avenir
print("Hello");
x + y;  // Expression statement (result is discarded)
```
