# Getting Started

This guide will help you get started with the Avenir programming language.

## Installation

Avenir is implemented in Go. To build from source:

```bash
git clone <repository-url>
cd avenir
go build ./cmd/avenir
```

The compiled binary provides the `avenir` command-line tool.

## Your First Program

Create a file `hello.av`:

```avenir
pckg main;

fun main() | void {
    print("Hello, Avenir!");
}
```

Every Avenir program must start with a package declaration (`pckg main;`) and must have a `main()` function.

## Running Programs

To compile and run a program:

```bash
avenir run hello.av
```

This will:
1. Parse the source file
2. Type-check the program
3. Compile to bytecode
4. Execute on the virtual machine

The `main()` function serves as the entry point.

## Building Bytecode

To compile to bytecode for later execution:

```bash
avenir build hello.av -o hello.avc
```

This creates a `.avc` bytecode file that can be executed directly:

```bash
avenir run hello.avc
```

## Command-Line Interface

The `avenir` tool provides the following commands:

### `avenir run <file>`

Compile and execute a `.av` source file or execute a `.avc` bytecode file.

```bash
avenir run program.av
avenir run program.avc
```

### `avenir build <file> [options]`

Compile a `.av` source file to bytecode.

Options:
- `-o <file>`: Output file name (default: `<input>.avc`)
- `-target <target>`: Build target: `bytecode` (default) or `native` (not yet implemented)

```bash
avenir build program.av -o program.avc
```

### `avenir version`

Display the Avenir version.

### `avenir help`

Display usage information.

## Program Structure

An Avenir program consists of:

1. **Package declaration**: `pckg <name>;`
2. **Imports** (optional): `import <module>;`
3. **Structs** (optional): Struct type definitions
4. **Interfaces** (optional): Interface type definitions
5. **Functions**: Function and method declarations
6. **Main function**: Required entry point

Example:

```avenir
pckg main;

import std.json as json;

struct Point {
    x | int
    y | int
}

fun (p | Point).sum() | int {
    return p.x + p.y;
}

fun main() | void {
    var p | Point = Point{x = 3, y = 4};
    print(p.sum());
    print(json.stringify({ "point": "ok" }));
}
```

**Note**: When importing modules, the file name must match a struct name in that file (if the file contains structs). See [Modules](modules.md) for details on the import system.

## Next Steps

- Read about [Syntax](syntax.md) to understand the language grammar
- Learn about [Types](types.md) to understand the type system
- Explore [Control Flow](control-flow.md) for conditionals and loops
- See [Functions](functions.md) for function definitions and calls
