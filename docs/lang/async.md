# Async & Concurrency

Avenir provides built-in support for asynchronous programming using `async`/`await` and a cooperative task scheduler.

## Overview

- `async fun` declares a function that runs as a concurrent task
- Calling an async function returns a `Future<T>` immediately
- `await` suspends the current task until a future resolves
- Multiple async calls can run concurrently (non-blocking I/O, timers, etc.)

## Async Functions

```avenir
async fun fetchData() | string {
    var content | string = await asyncReadFile("data.txt");
    return content;
}
```

The declared return type is the *inner* type. The caller receives `Future<string>`, but `return` statements inside the function use the inner type directly.

## Async Methods

Methods can also be async. The return type wrapping works the same way:

```avenir
pub struct HttpClient {
    base_url | string
}

pub async fun (client | HttpClient).get(path | string) | string {
    var url | string = client.base_url + path;
    return await asyncHttpGet(url);
}

// Usage
var client | HttpClient = HttpClient{base_url = "https://api.example.com"};
var data | Future<string> = client.get("/users");
var result | string = await data;
```

Async methods follow the same rules as async functions:
- Instance methods have the receiver as the first parameter (not wrapped in Future)
- The return type is wrapped in `Future<T>`
- `await` can be used inside async method bodies

## Future\<T\>

`Future<T>` is a built-in generic type representing a value that will be available later. You can store futures in variables:

```avenir
var f | Future<int> = compute(42);
```

A future is resolved when the spawned task completes. Use `await` to extract the result:

```avenir
var result | int = await f;
```

Awaiting an already-resolved future returns immediately.

## Concurrency

Calling multiple async functions before awaiting any of them runs the tasks concurrently:

```avenir
async fun download(url | string) | string {
    return await asyncHttpGet(url);
}

async fun main() | void {
    var a | Future<string> = download("https://example.com/1");
    var b | Future<string> = download("https://example.com/2");

    var ra | string = await a;
    var rb | string = await b;

    print(ra);
    print(rb);
}
```

Both downloads run in parallel. The total wall time is approximately `max(time_a, time_b)`, not `time_a + time_b`.

## Execution Model

Avenir uses a single-threaded cooperative scheduler with a non-blocking event loop:

1. Each async function call creates a **child task** with its own stack
2. The task is placed on a **ready queue**
3. The event loop runs one task at a time until it either completes or suspends (at an `await` on a pending future)
4. Suspended tasks are parked until their awaited future resolves
5. Background I/O operations run in Go goroutines; when they complete, the associated future is resolved and the waiting task is re-scheduled

This model avoids shared-state data races while enabling concurrent I/O.

## Async Standard Library

The standard library provides async variants of I/O operations:

### Time

```avenir
import std.time;

async fun example() | void {
    await std.time.asyncSleep(1000000000);  // 1 second in nanoseconds
}
```

### File System

```avenir
import std.fs;

async fun example() | void {
    var f | std.fs.File = await std.fs.asyncOpen("file.txt");
    var data | string = await f.asyncReadString();
    await f.asyncClose();
}
```

Async FS functions: `asyncOpen`, `asyncExists`, `asyncRemove`, `asyncMkdir`.
Async File methods: `asyncRead`, `asyncReadAll`, `asyncReadString`, `asyncWrite`, `asyncWriteString`, `asyncClose`.

### Network

```avenir
import std.net;

async fun example() | void {
    var sock | std.net.Socket = await std.net.asyncConnect("127.0.0.1", 8080);
    await sock.asyncWrite("hello");
    var resp | string = await sock.asyncRead(1024);
    await sock.asyncClose();
}
```

### HTTP

```avenir
import std.http.client;

async fun example() | string {
    var resp | string = await std.http.client.asyncGet("https://example.com");
    return resp;
}
```

Async HTTP functions: `asyncRequest`, `asyncGet`, `asyncPost`, `asyncPut`, `asyncDelete`.

## Error Handling

Async functions support the same `try`/`catch` error handling as synchronous code. If an async operation fails, the future is rejected and `await` propagates the error:

```avenir
async fun safeFetch() | string {
    try {
        return await asyncHttpGet("https://example.com");
    } catch (e | error) {
        return "fallback";
    }
}
```

## Timeouts

Use `withTimeout` from `std.time` to race a future against a deadline:

```avenir
import std.time;

async fun main() | void {
    var f | Future<string> = fetchData();
    try {
        var result | string = await std.time.withTimeout(f, std.time.fromSeconds(5));
        print(result);
    } catch (e | error) {
        print("request timed out");
    }
}
```

If the future resolves before the deadline, `withTimeout` returns the result. Otherwise it throws a timeout error. The first-write-wins semantics ensure the original future's late resolution is safely ignored.

You can also use the builtin directly with nanosecond durations:

```avenir
var result | int = await __builtin_async_with_timeout(future, 5000000000);
```

## Rules

- `await` can only be used inside `async fun` bodies
- `Future<T>` is the only type that can be awaited
- Async functions cannot be called with `spawn` — concurrency is automatic when calling an `async fun`
- The `main` function can be `async`
