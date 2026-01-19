# Strings

Avenir supports string literals with both double quotes (`"..."`) and single
quotes (`'...'`). Both forms are identical and produce a `string` value.

Single-quoted strings are **not** characters. There is no `char` type.

Examples:

```avenir
var a | string = "hello";
var b | string = 'world';
var c | string = a + " " + b;
```

## Escapes

Both quote styles support the same escape sequences:

- `\\` backslash
- `\"` double quote
- `\'` single quote
- `\n` newline
- `\t` tab
- `\r` carriage return
- `\0` null byte
- `\uXXXX` Unicode codepoint (hex)
- `\xNN` byte (hex)

Examples:

```avenir
var s1 | string = "line\nbreak";
var s2 | string = 'quote: \'';
var s3 | string = "backslash: \\";
```

Strings cannot contain a raw newline. Use `\n` to embed line breaks.

## Interpolation

String interpolation uses `${ ... }` and works the same in both quote styles:

```avenir
var name | string = 'Avenir';
print("hello ${name}");
print('hello ${name}');
```

Interpolation converts embedded values to strings using the same runtime
stringification as `print`.

## Concatenation

The `+` operator supports string concatenation only when **both operands are
strings**. There are no implicit conversions for `+`. Use interpolation when
you need to include non-strings.
