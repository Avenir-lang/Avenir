# Avenir

Avenir is a statically‑typed programming language with an explicit, readable
syntax and a compact VM runtime. It is designed for building reliable software
with clear type boundaries and a pragmatic standard library.

## Why Avenir

- **Explicit types**: no hidden inference; types are always visible.
- **Practical standard library**: filesystem, networking, JSON, HTTP, time.
- **Simple runtime model**: bytecode VM with clear semantics.
- **Error handling**: `try / catch` with a single `error` type.

## Status

**Alpha**. The language is usable for experimentation, but the surface area is
still evolving and some features are not yet implemented (see roadmap).

## Example

```avenir
pckg main;

import std.time;

fun main() | void {
    var now | time.DateTime = time.now();
    print(now.format("YYYY-MM-DD HH:mm:ss"));
}
```

## Build

Requirements:

- Go 1.22+

Build the compiler/VM:

```bash
go build -o avenir ./cmd/avenir
```

## Run

```bash
./avenir run ./examples/hello.av
```

## Tests

```bash
go test ./...
```

## Repository Layout

- `cmd/avenir` — CLI entrypoint
- `internal/lexer` — lexer
- `internal/parser` — parser
- `internal/ast` — AST
- `internal/types` — type checker
- `internal/ir` — bytecode compiler
- `internal/vm` — VM runtime
- `internal/runtime` — builtins and host services
- `std/` — standard library (Avenir code)
- `docs/` — language and engine documentation
- `examples/` — example programs

## Roadmap (Short)

- Native code backend (currently VM only)
- Generics for user‑defined types
- Optional chaining and pattern matching
- Expanded stdlib utilities

## License

MIT — see `LICENSE`.
