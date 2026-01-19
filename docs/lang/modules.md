# Modules

Avenir uses a module system for organizing code. Each file is a module, identified by its package declaration.

## Package Declaration

Every file must start with a package declaration:

```avenir
pckg <name>;
```

The package name can be a simple identifier or a dotted path:

```avenir
pckg main;
pckg std.net;
pckg app.utils;
```

The package name identifies the module and is used in imports.

## Imports

Modules can import other modules using the `import` statement:

```avenir
import <module_path>;
```

The import path is a dotted identifier matching the package name of the imported module:

```avenir
import std.net;
import std.json as json;
```

### Import Aliases

Imports can have aliases:

```avenir
import std.json as json;
```

The alias allows accessing the module using a different name. If no alias is specified, the last segment of the path is used as the local name.

### Using Imported Modules

Imported modules are accessed using dot notation:

```avenir
import std.net as net;

var sock | net.Socket = net.connect("example.com", 80);
```

## File-to-Struct Mapping

**A file can only be imported if it contains a struct with the same name as the file.**

### Rule

If a file contains structs, **at least one struct must have the same name as the file** (without the `.av` extension).

Examples:

**Valid:**
```
File: Point.av
Contains: struct Point { x | int; y | int; }
→ Can be imported
```

**Invalid:**
```
File: Point.av
Contains: struct Rectangle { x | int; y | int; }
→ Compile-time error: "file 'Point.av' does not contain struct 'Point'"
```

**Files without structs:**
Files that contain only functions (no structs) can still be imported. The file-to-struct mapping rule only applies to files that contain structs.

**Multiple structs:**
If a file contains multiple structs, at least one must match the file name:

```avenir
// File: Point.av
pckg geometry;

pub struct Point {
    x | int
    y | int
}

pub struct Rectangle {
    w | int
    h | int
}
```

This is valid because `Point.av` contains `struct Point`.

## Folder-Based Imports

If a folder `A` contains a file `A.av`, the import should reference only the folder:

**Example:**
```
Folder: module/geometry
File: geometry/geometry.av
Contains: struct geometry { x | int; y | int; }

Import: import module.geometry;
```

The compiler automatically resolves `import module.geometry` to `module/geometry/geometry.av`.

### Resolution Order

When resolving an import path like `module.A`, the compiler tries:

1. **Folder-based**: `module/A/A.av` (if folder `A` exists and contains `A.av`)
2. **Flat file**: `module/A.av` (if file exists directly)

This applies to both standard library modules (`std.*`) and application modules.

### Nested Module Imports

For nested module paths, the same resolution logic applies:

```
Folder: app/utils/
File: app/utils/utils.av
Package: pckg app.utils;
Struct: struct utils { ... }

Import: import app.utils;
```

The compiler resolves `import app.utils` to `app/utils/utils.av`.

## Module Resolution

The compiler resolves imports by:

1. For `std.*` modules:
   - First checks `std/<path>/<last-segment>.av` (folder-based)
   - Then checks `std/<path>.av` (flat file)
   - Falls back to repository root `std/` directory if needed

2. For other modules:
   - First checks `<project-root>/<path>/<last-segment>.av` (folder-based)
   - Then checks `<project-root>/<path>.av` (flat file)

The project root is determined as the directory containing the entry file.

## Multi-File Modules

Modules can be split across multiple files when using folder-based imports.
All `.av` files in the module directory are merged into a single module.

Example layout:

```
std/net/
  net.av
  socket.av
  server.av
```

If a file contains structs, it must contain a struct that matches the filename.
For helper files with public structs that do not match the filename, add a
private placeholder:

```avenir
// std/net/socket.av
pckg std.net;
struct socket {}  // placeholder to satisfy file-to-struct mapping
pub struct Socket { handle | any }
```

## Compile-Time Validation

The compiler performs strict validation:

- **Missing file/folder**: Error if import path cannot be resolved
- **Struct name mismatch**: Error if file contains structs but none match the file name
- **Folder exists but file missing**: Clear error message indicating expected file location

### Error Messages

The compiler provides clear, actionable error messages:

```
Error: folder "module/geometry" exists but does not contain required file "geometry.av"
Expected: module/geometry/geometry.av
```

```
Error: file "Point.av" does not contain struct "Point" (found structs: Rectangle, Circle)
A file can only be imported if it contains a struct with the same name as the file
```

```
Error: cannot find module "nonexistent" (looked for folder .../nonexistent/nonexistent.av and file .../nonexistent.av)
```

## Module Access

After importing a module, you can access its public functions and structs:

```avenir
import std.net as net;

fun main() | void {
    var sock | net.Socket = net.connect("example.com", 80);
    sock.close();
}
```

## Module Visibility

Only `pub` functions, structs, and interfaces are accessible from other modules. Private functions and structs are only accessible within the same module.

## Import Alias Notes

If no alias is provided, the last segment of the module path becomes the local
name. For convenience with `std.http.client` and `std.http.server`, import with
`as http`:

```avenir
import std.http.client as http;
import std.http.server as http;
```

## Examples

### Example 1: Flat File Import

```
File: utils.av
Package: pckg utils;
Struct: struct utils { ... }

Import: import utils;
```

### Example 2: Folder-Based Import

```
Folder: math/
File: math/math.av
Package: pckg math;
Struct: struct math { ... }

Import: import math;
```

### Example 3: Nested Module Import

```
Folder: app/utils/
File: app/utils/utils.av
Package: pckg app.utils;
Struct: struct utils { ... }

Import: import app.utils;
```

### Example 4: Standard Library Import

```
Folder: std/net/
File: std/net/net.av
Package: pckg std.net;
Functions: pub fun connect(...), listen(...)

Import: import std.net;
```

## Module Structure

A module can contain:

1. **Package declaration** (required)
2. **Imports** (optional, zero or more)
3. **Struct declarations** (optional, zero or more)
4. **Interface declarations** (optional, zero or more)
5. **Function declarations** (required, at least one)

## Import Cycles

Import cycles are detected and reported as errors. A module cannot import itself (directly or indirectly).
