# std.fs

`std.fs` provides blocking filesystem access built on runtime file primitives.

## Overview

The module exposes a high-level `File` type with opaque handles. All filesystem
operations throw runtime errors that are catchable via `try`/`catch`.

## API

### Public struct

```avenir
pub struct File {
    handle | any
}
```

### Functions

| Function | Parameters | Returns | Errors |
| --- | --- | --- | --- |
| `open` | `path | string`, `mode | string` | `File` | invalid path, permission, etc. |
| `exists` | `path | string` | `bool` | stat failures |
| `remove` | `path | string` | `void` | missing path, permission |
| `mkdir` | `path | string` | `void` | permission, invalid path |

### Path helpers

```avenir
pub fun join(a | string, b | string) | string
pub fun basename(path | string) | string
```

### Errors

```avenir
pub fun errorWithPath(message | string, path | string) | error
```

### File methods

| Method | Parameters | Returns | Errors |
| --- | --- | --- | --- |
| `read` | `n | int` | `bytes` | invalid handle, I/O errors |
| `readAll` | — | `bytes` | invalid handle, I/O errors |
| `readString` | — | `string` | invalid handle, UTF-8 errors |
| `write` | `data | bytes` | `int` | invalid handle, I/O errors |
| `writeString` | `data | string` | `int` | invalid handle, I/O errors |
| `close` | — | `void` | invalid handle |

## Open modes

`open` accepts the following modes:

- `r`  : read-only
- `w`  : write-only, truncates or creates
- `a`  : append-only, creates if missing
- `r+` : read/write
- `w+` : read/write, truncates or creates
- `a+` : read/write, append-only
- `rw` : read/write, creates if missing

## Examples

### Basic file write / read

```avenir
import std.fs;

fun main() | void {
    try {
        var f | fs.File = fs.open("hello.txt", "w");
        f.writeString("Hello, world!\n");
        f.close();

        var r | fs.File = fs.open("hello.txt", "r");
        var text | string = r.readString();
        print(text);
        r.close();
    } catch (e | error) {
        print("FS error: " + errorMessage(e));
    }
}
```

### Path helpers

```avenir
import std.fs;

fun main() | void {
    var full | string = fs.join("/tmp", "file.txt");
    print(fs.basename(full)); // "file.txt"
}
```

## Path resolution

Relative paths are resolved against the entry file’s directory (the execution
root). Absolute paths are used as-is.

## Error handling

All filesystem operations may throw runtime errors. Use `try`/`catch` to handle
failures such as missing files or permission errors.

## Blocking behavior

`read`, `readAll`, and `readString` are blocking operations. They wait for data
from the underlying file descriptor and return once data is read or EOF is
reached.
