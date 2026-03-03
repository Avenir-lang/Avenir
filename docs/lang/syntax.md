# Syntax

This document describes the syntax of the Avenir programming language.

## Program Structure

Every Avenir program must start with a package declaration:

```avenir
pckg <name>;
```

The package name can be a simple identifier or a dotted path (e.g., `std.net`, `app.utils`).

After the package declaration, a program can contain:

1. **Imports** (optional, zero or more)
2. **Struct declarations** (optional, zero or more)
3. **Function declarations** (at least one, including `main`)

## Keywords

The following keywords are reserved:

- `pckg` - Package declaration
- `import` - Import declaration
- `fun` - Function declaration
- `struct` - Struct declaration
- `pub` - Visibility modifier
- `mut` - Mutability modifier
- `var` - Variable declaration
- `if`, `else` - Conditional statements
- `while` - While loop
- `for` - For loop
- `return` - Return statement
- `throw` - Throw exception
- `try`, `catch` - Exception handling
- `break` - Break from loop
- `true`, `false` - Boolean literals
- `none`, `some` - Optional literals

## Type Keywords

- `int` - Integer type
- `float` - Floating-point type
- `string` - String type
- `bool` - Boolean type
- `void` - Void type (no return value)
- `any` - Any type
- `error` - Error type
- `bytes` - Bytes type
- `list` - List type

## Operators

### Arithmetic

- `+` - Addition
- `-` - Subtraction
- `*` - Multiplication
- `/` - Division
- `%` - Modulo (integer only)

### Comparison

- `==` - Equality
- `!=` - Inequality
- `<` - Less than
- `<=` - Less than or equal
- `>` - Greater than
- `>=` - Greater than or equal

### Logical

- `&&` - Logical AND
- `||` - Logical OR
- `!` - Logical NOT

### Assignment

- `=` - Assignment

## Symbols

- `|` - Type separator (used in `name | type`)
- `;` - Statement terminator
- `,` - List separator
- `.` - Member access
- `(` `)` - Parentheses (grouping, function calls)
- `{` `}` - Braces (blocks, struct literals)
- `[` `]` - Brackets (list literals, indexing)
- `<` `>` - Angle brackets (type parameters, unions)
- `?` - Optional type suffix
- `:` - Dictionary key separator (in dict literals: `key: value`)

## Comments

Avenir supports both single-line and multi-line comments:

```avenir
// Single-line comment

/*
 * Multi-line comment
 * Can span multiple lines
 */
```

## Identifiers

Identifiers start with a letter or underscore, followed by letters, digits, or underscores.

## Literals

### Integer Literals

```avenir
42
-10
0
```

### Float Literals

```avenir
3.14
-0.5
1.0
```

### String Literals

```avenir
"Hello, World!"
"Line1\nLine2"
'Example'
```

See `docs/strings.md` for escape sequences, interpolation, and single-quote rules.

### Boolean Literals

```avenir
true
false
```

### Bytes Literals

```avenir
b"bytes data"
```

### List Literals

```avenir
[1, 2, 3]
["a", "b", "c"]
[]
```

### Dict Literals

```avenir
{ name: "Alex", "age": 30 }
```

### Struct Literals

```avenir
Point{x = 10, y = 20}
Config{host = "localhost", port = 8080}
```

## Expression Precedence

From highest to lowest:

1. Primary expressions (literals, identifiers, parentheses)
2. Unary operators (`!`, `-`)
3. Multiplicative (`*`, `/`, `%`)
4. Additive (`+`, `-`)
5. Relational (`<`, `<=`, `>`, `>=`)
6. Equality (`==`, `!=`)
7. Logical AND (`&&`)
8. Logical OR (`||`)

## Statement Terminators

Statements are terminated with semicolons (`;`):

```avenir
var x | int = 10;
x = 20;
print(x);
```

Blocks use braces and do not require semicolons:

```avenir
{
    var x | int = 10;
    print(x);
}
```
