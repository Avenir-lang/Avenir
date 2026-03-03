# Runtime

Avenir programs are executed on a virtual machine (VM). The VM is stack-based and executes bytecode.

## Execution Model

Programs are compiled to bytecode, which is then executed by the VM. The VM:

1. **Parses** the source code
2. **Type-checks** the program
3. **Compiles** to bytecode (IR)
4. **Executes** on the VM

## Virtual Machine

The VM is stack-based, using a call stack and operand stack.

### Stack

The VM uses a stack for:

- **Operands**: Values used in operations
- **Function calls**: Arguments and return values
- **Local variables**: Function-local variables

### Execution

The VM executes instructions sequentially, using an instruction pointer (IP). Control flow instructions (jumps, branches) modify the IP.

### Function Calls

Functions are called by:

1. Pushing arguments onto the stack
2. Creating a new frame
3. Executing the function body
4. Popping the return value
5. Restoring the previous frame

### Closures

Closures capture variables from their enclosing scope. Captured variables are stored in the closure's upvalue array.

## Built-in Functions

Built-in functions are implemented in Go and called directly by the VM. They provide:

- **I/O operations**: `print`, `input`
- **Type operations**: `len`, `typeOf`, `error`, `errorMessage`
- **Type conversions**: `fromString`, `toInt`

## Built-in Methods

Built-in methods are implemented for:

- **Lists**: Collection operations
- **Strings**: String manipulation
- **Bytes**: Byte operations
- **Dicts**: Key/value operations

Methods are called using the same mechanism as user-defined methods.

## Internal stdlib builtins

The standard library is implemented in Avenir and delegates to internal
runtime builtins. These are not part of the public API, but they exist for
completeness:

- `__builtin_fs_*` (filesystem primitives)
- `__builtin_socket_*` (TCP primitives)
- `__builtin_json_*` (JSON parse/stringify)
- `__builtin_http_*` (HTTP client/server primitives)

## Error Handling

The VM uses a unified error model. Any error after successful compilation becomes a language-level `error` value and can be caught with `try`/`catch`:

1. **Try blocks**: Register exception handlers
2. **Throw statements**: Raise explicit exceptions
3. **Runtime checks**: Division by zero, invalid access, etc.
4. **Builtins**: Failures from built-in functions

Unhandled exceptions propagate up the call stack. If an exception reaches the top level, the program terminates with a clean error message.

## Memory Management

The VM manages memory for:

- **Stack values**: Automatically allocated and freed
- **Closures**: Captured variables
- **Structs**: Field values
- **Lists**: Elements

Memory is managed automatically; there is no manual memory management.

## Performance

The VM is designed for:

- **Fast execution**: Optimized instruction dispatch
- **Small bytecode**: Compact representation
- **Efficient function calls**: Fast frame management

For production use, consider compiling to native code (not yet implemented).
