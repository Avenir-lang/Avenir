# Modules and Imports

This document describes module loading and import resolution in
`internal/modules`.

## Overview

Modules are identified by their `pckg` declaration and imported by dotted path.
The loader parses the entry file, resolves all imports recursively, merges
multi‑file modules, and validates file/struct rules.

Key files:

- `internal/modules/loader.go`
- `internal/ast/ast.go`

## Load Flow

`LoadWorld(entryFile)`:

1. Resolve the entry file path.
2. Determine `projectRoot` as the entry file directory.
3. Load entry module and recursively load imports.
4. Detect import cycles.

Loaded modules are stored in a `World` keyed by fully‑qualified module name.

## File Resolution

For a module `app.utils`:

1. Try folder‑based: `app/utils/utils.av`
2. Try flat file: `app/utils.av`

For `std.*` modules, resolution is identical but rooted under `std/`. If not
found in the entry project root, the loader falls back to the repository root.

## Multi‑File Modules

If the module is a folder‑based import, all `.av` files in the module directory
are loaded and merged into a single AST (`mergePrograms`).

## File‑to‑Struct Mapping

If a file contains structs, at least one struct must match the file name
(`Foo.av` must contain `struct Foo`). Files without structs are allowed.

This rule is validated by `validateFileStructMapping`.

## Import Aliases

Imports use the last path segment as the local name by default, but aliases can
override:

```avenir
import std.http.client as http;
```

The checker inserts both the alias and the full module path into scope to
support qualified type names (`net.Socket`).

## Notes

- Import cycles are detected and reported as errors.
- All files in a module must agree on the package name.

## References

- `internal/modules/loader.go`
