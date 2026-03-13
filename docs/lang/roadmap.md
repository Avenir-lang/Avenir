# Roadmap

This roadmap outlines potential future improvements. Items may change as the
language evolves.

## Language Core

- Native backend for `avenir build -target=native` (currently unimplemented)
- ~~Generic type argument inference~~ (implemented: explicit type args only)
- ~~Generics for built-in collections ergonomics~~ (implemented: generic dict<K,V>)
- Advanced optional ergonomics (coalescing/operators beyond `?.`)
- Pattern matching / match expressions (beyond `switch`)
- Extended `defer` semantics and diagnostics
- ~~Typed errors with struct types~~ (implemented: ! syntax, multiple catch clauses)

## Runtime and VM

- Improved diagnostics (stack traces, error metadata)
- Performance profiling hooks

## Standard Library

- ~~Async I/O primitives~~ (implemented: async FS, Net, HTTP, timers)
- ~~Task cancellation and timeouts~~ (implemented: withTimeout function)
- Expanded filesystem APIs (metadata, directory iteration)
- ~~HTTP enhancements (TLS, middleware, streaming bodies)~~ (implemented: TLS support)
- ~~WebSocket support~~ (implemented: std.net.socket)
- ~~TLS/HTTPS support~~ (implemented: std.crypto.tls)
- ~~SQL client library~~ (implemented: std.sql with postgres driver)
- ~~HTML templating system~~ (implemented: std.web.html)
- ~~Decorators and variadic generics~~ (implemented: @decorator syntax)

## Tooling

- Formatter and linter
- Package manager / registry