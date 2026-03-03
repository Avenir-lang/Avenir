# Lexer

This document describes Avenir’s lexical analysis as implemented in
`internal/lexer/lexer.go` and token definitions in `internal/token/token.go`.

## Overview

The lexer converts source text into a stream of tokens with line/column
positions. It is rune-based (UTF-8) and tracks the current character, line,
and column for precise error reporting.

## Token Categories

The lexer recognizes:

- **Identifiers** and **keywords**
- **Numbers**: integers and floats
- **Strings**: single-quoted and double-quoted
- **Bytes literals**: `b"..."` (double quotes only)
- **Operators**: `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `||`, `!`
- **Symbols**: `(` `)` `{` `}` `[` `]` `.` `,` `;` `:` `|` `?`
- **Interpolation markers**: `${`, `}`, plus string-part tokens for interpolated strings
- **EOF** and **Illegal** tokens

See `internal/token/token.go` for the canonical token list.

## Whitespace and Comments

- All Unicode whitespace is skipped.
- Line comments start with `//` and run to the end of the line.
- Block comments are delimited by `/* ... */` and can span multiple lines.

## Numbers

Numbers are lexed as:

- **Int**: digits only, e.g. `42`
- **Float**: digits with `.` or exponent `e/E`, e.g. `3.14`, `1e3`

The lexer classifies numeric tokens as `Int` or `Float` based on the presence
of `.` or `e/E`.

## Strings

Strings are delimited by either `"` or `'`. Both forms produce the same token
types and semantics.

### Escapes

Escape sequences are handled in the lexer (AST receives decoded strings):

| Escape | Meaning |
| --- | --- |
| `\\` | backslash |
| `\"` | double quote |
| `\'` | single quote |
| `\n` | newline |
| `\t` | tab |
| `\r` | carriage return |
| `\0` | null |
| `\uXXXX` | Unicode codepoint |
| `\xNN` | byte value |

Invalid escapes and unterminated strings are recorded as lexer errors and
produce `Illegal` tokens.

### Interpolation

Interpolated strings use `${ ... }`. The lexer emits a sequence of tokens:

- `StringPart` for plain text segments
- `InterpStart` for `${`
- tokens for the interpolated expression
- `InterpEnd` for `}`
- `StringEnd` when the interpolated string ends

Non-interpolated strings produce a single `String` token.

Example:

```avenir
"hello ${name}"
```

Token stream (simplified):

```
StringPart("hello ")
InterpStart
Ident(name)
InterpEnd
StringEnd
```

## Bytes Literals

Bytes literals use `b"..."` and support the same escape sequences as strings.
Interpolation is not supported in bytes literals.

## Error Reporting

The lexer collects errors as formatted strings with line/column positions.
Consumers can inspect `lexer.Errors()` to retrieve them.

## References

- `internal/lexer/lexer.go`
- `internal/token/token.go`
