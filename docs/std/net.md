# std.net

`std.net` provides blocking TCP networking built on runtime socket primitives.

## Overview

The module exposes high-level `Socket` and `Server` types while keeping raw socket
handles opaque. Errors thrown by networking operations are standard runtime
errors and are catchable via `try`/`catch`.

## API

### Public structs

```avenir
pub struct Socket {
    handle | any
}

pub struct Server {
    handle | any
}
```

### Functions

| Function | Parameters | Returns | Errors |
| --- | --- | --- | --- |
| `connect` | `host | string`, `port | int` | `Socket` | connection errors |
| `listen` | `host | string`, `port | int` | `Server` | bind errors |

### Socket methods

| Method | Parameters | Returns | Errors |
| --- | --- | --- | --- |
| `read` | `n | int` | `bytes` | invalid handle, I/O errors |
| `readAll` | — | `bytes` | invalid handle, I/O errors |
| `readString` | — | `string` | invalid handle, UTF-8 errors |
| `write` | `data | bytes` | `int` | invalid handle, I/O errors |
| `writeString` | `data | string` | `int` | invalid handle, I/O errors |
| `close` | — | `void` | invalid handle |

### Server methods

| Method | Parameters | Returns | Errors |
| --- | --- | --- | --- |
| `accept` | — | `Socket` | invalid handle, accept errors |
| `close` | — | `void` | invalid handle |

## Examples

### TCP client

```avenir
import std.net;

fun main() | void {
    try {
        var sock | net.Socket = net.connect("example.com", 80);
        sock.writeString("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n");

        var resp | string = sock.readString();
        print(resp);

        sock.close();
    } catch (e | error) {
        print("Network error: " + errorMessage(e));
    }
}
```

### TCP server

```avenir
import std.net;

fun main() | void {
    var srv | net.Server = net.listen("0.0.0.0", 8080);

    while (true) {
        var client | net.Socket = srv.accept();
        client.writeString("Hello!\n");
        client.close();
    }
}
```

## Error handling

All networking operations may throw runtime errors. Use `try`/`catch` to handle
failures such as connection errors or read/write failures.

## Blocking behavior

`read`, `readAll`, and `accept` are blocking operations. They will wait for data
or incoming connections. Use `try`/`catch` to handle unexpected disconnections.
