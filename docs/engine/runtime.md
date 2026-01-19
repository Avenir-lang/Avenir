# Runtime and Builtins

This document covers the runtime services and builtins subsystem.

## Overview

The runtime bridges VM execution with host services (I/O, networking,
filesystem, HTTP). Builtins are registered in Go and exposed to Avenir as
functions or methods.

Key files:

- `internal/runtime/runtime.go`
- `internal/runtime/env.go`
- `internal/runtime/builtins/registry.go`

## Environment (`Env`)

`runtime.Env` provides:

- IO (`Println`, `ReadLine`)
- Net (`Connect`, `Listen`, `Accept`, `Read`, `Write`, `Close`)
- FS (`Open`, `Read`, `Write`, `Close`, `Exists`, `Remove`, `Mkdir`)
- HTTP (`Request`, `Listen`, `Accept`, `Respond`)
- `ExecRoot()` for path resolution
- Closure invocation hooks for list functions (`map`, `filter`, `reduce`)

The VM configures the environment when it starts, including struct type names
and the closure caller.

## Builtins Registry

Builtins are registered via `builtins.Register` in `init()` functions. The
runtime imports all builtin packages to trigger registration:

```go
_ "avenir/internal/runtime/builtins/bytes"
_ "avenir/internal/runtime/builtins/collections"
// ...
```

Each builtin exposes:

- `Meta` (name, arity, param types, receiver type)
- `Call` (implementation)

### Functions vs Methods

- Functions use `ReceiverType = TypeVoid`.
- Methods use `ReceiverType` and `MethodName` for lookup by receiver type.

## Call Flow

VM → `runtime.CallBuiltin`:

1. Lookup builtin by ID
2. Convert `[]value.Value` → `[]interface{}`
3. Call builtin with `Env`
4. Convert result back to `value.Value`

Errors from builtins propagate through the VM and are catchable with
`try / catch`.

## Standard Library Integration

The Avenir stdlib (`./std`) is written in Avenir but delegates to internal
builtins for host access:

- `std.fs` → `__builtin_fs_*`
- `std.net` → `__builtin_socket_*`
- `std.json` → `__builtin_json_*`
- `std.http` → `__builtin_http_*`
- `std.time` → `__builtin_time_*`

These builtins are not part of the public language surface but are stable
internal interfaces used by std modules.

## Exec Root and Path Resolution

The runtime environment exposes `ExecRoot()` to resolve relative paths in
`std.fs`. This is set to the directory of the entry `.av` file so that user
paths resolve relative to the executing program, not the compiler’s working
directory.

## Notes

- Builtins operate on `value.Value` and return `value.Value`.
- Builtins should not mutate list values; dict methods may mutate in place.

## References

- `internal/runtime/runtime.go`
- `internal/runtime/env.go`
- `internal/runtime/builtins/registry.go`

