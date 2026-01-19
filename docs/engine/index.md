# Engine Documentation

This section documents the internal engine of Avenir: the lexer, parser, AST,
type checker, IR, VM, runtime services, and module system. It is intended for
contributors working on the compiler, runtime, and standard library.

## Pipeline Overview

1. **Lexing** → `internal/lexer`
2. **Parsing** → `internal/parser`
3. **AST** → `internal/ast`
4. **Type checking** → `internal/types`
5. **IR generation** → `internal/ir`
6. **VM execution** → `internal/vm`
7. **Runtime services & builtins** → `internal/runtime`
8. **Modules and imports** → `internal/modules`

## Documents

- [Lexer](lexer.md)
- [Parser](parser.md)
- [AST](ast.md)
- [Types and Type Checker](types.md)
- [IR (Intermediate Representation)](ir.md)
- [VM](vm.md)
- [Runtime and Builtins](runtime.md)
- [Modules and Imports](modules.md)
- [Errors](errors.md)
- [Memory and Performance](memory.md)
- [Testing](testing.md)

Each document is written to match the current codebase and includes references
to relevant packages and files.
