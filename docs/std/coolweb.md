# std.coolweb

Async web framework with decorator-based routing, built on `std.http.server`.

## Quick Start

```avenir
pckg main;

import std.coolweb;

var app | coolweb.App = coolweb.newApp();

@app.get("/")
fun index(ctx | coolweb.Context) | coolweb.Response {
    return ctx.text("Hello Avenir!");
}

async fun main() | void {
    await app.run(8080);
}
```

## App

Create an app with `coolweb.newApp()`. The app is declared as a top-level variable so decorators can access it at init time.

### Decorator Route Registration

```avenir
var app | coolweb.App = coolweb.newApp();

@app.get("/users")
fun listUsers(ctx | coolweb.Context) | coolweb.Response { ... }

@app.post("/users")
fun createUser(ctx | coolweb.Context) | coolweb.Response { ... }

@app.put("/users/:id")
fun updateUser(ctx | coolweb.Context) | coolweb.Response { ... }

@app.delete("/users/:id")
fun deleteUser(ctx | coolweb.Context) | coolweb.Response { ... }
```

Available methods: `.get`, `.post`, `.put`, `.patch`, `.delete`, `.options`.

Each returns a decorator that registers the route and returns the handler unchanged.

### Middleware

```avenir
app.use(coolweb.loggerMiddleware);
app.use(myMiddleware);
```

Middleware signature: `fun(coolweb.Context, fun() | coolweb.Response) | coolweb.Response`

Call `next()` to continue the chain, or return a response directly to short-circuit.

### Error Handling

```avenir
app.onError(fun(ctx | coolweb.Context, e | error) | coolweb.Response {
    return ctx.json({ "error": "something went wrong" }, 500);
});
```

### Starting the Server

```avenir
async fun main() | void {
    await app.run(8080);
}
```

## Context

The `Context` struct is passed to every handler and middleware.

### Request Data

- `ctx.request` — `Request` struct with `method`, `path`, `fullPath`, `headers`, `body`
- `ctx.params` — `dict<string>` of path parameters (e.g. `:id` → `ctx.params["id"]`)
- `ctx.query` — `dict<string>` of query parameters
- `ctx.cookies` — `dict<string>` of parsed cookies
- `ctx.state` — `dict<any>` for middleware-shared state

### Response Builders

```avenir
ctx.text("Hello", 200)          // text/plain
ctx.json({ "key": "value" })    // application/json (default 200)
ctx.html("<h1>Hi</h1>")         // text/html
ctx.redirect("/other", 302)     // redirect
ctx.file("public/index.html")   // file bytes + content type by extension
```

### Body Parsers

```avenir
var data | any = ctx.jsonBody();       // parse JSON body
var text | string = ctx.textBody();    // body as string
var raw | bytes = ctx.bodyBytes();     // raw bytes
```

## Router

Routers allow modular route grouping with path prefixes.

```avenir
var users | coolweb.Router = coolweb.newRouter("/users");

@users.get("/")
fun listUsers(ctx | coolweb.Context) | coolweb.Response { ... }

@users.get("/:id")
fun getUser(ctx | coolweb.Context) | coolweb.Response { ... }
```

Mount routers into the app or nest them:

```avenir
var api | coolweb.Router = coolweb.newRouter("/api");
api.mount(users);
app.mount(api);
```

Routers support their own middleware via `router.use(mw)`.

## Path Parameters

Use `:name` in route patterns:

```avenir
@app.get("/users/:id/posts/:postId")
fun getPost(ctx | coolweb.Context) | coolweb.Response {
    var userId | string = ctx.params["id"];
    var postId | string = ctx.params["postId"];
    return ctx.json({ "userId": userId, "postId": postId });
}
```

## Built-in Middleware

### Logger

```avenir
app.use(coolweb.loggerMiddleware);
```

Default logger emits structured `key=value` lines for request start/end and errors.

Typical fields include:
- `event=req_start|req_end|req_error`
- `id=<request-id>`
- `ip=<client-ip>`
- `method=<HTTP method>`
- `route=<route pattern, e.g. /users/:id>`
- `path=<concrete path>`
- `status=<response status>`
- `duration_ms=<latency>`

Request ID behavior:
- Uses `X-Request-Id` from incoming headers when present
- Generates one when missing
- Adds `X-Request-Id` to response headers if absent

Configurable logger middleware:

```avenir
app.use(coolweb.newLogger(
    includeQuery = true,
    includeHeaders = true,
    includeCookies = false,
    includeBodyBytes = false,
    includeClientIp = true,
    skipPaths = ["/health"]
));
```

`loggerMiddleware` remains as a backward-compatible default wrapper.

### CORS

```avenir
var corsConfig | coolweb.CorsConfig = coolweb.CorsConfig{
    allowOrigins = ["*"],
    allowMethods = ["GET", "POST", "PUT", "DELETE"],
    allowHeaders = ["Content-Type", "Authorization"]
};
app.use(coolweb.corsMiddleware(corsConfig));
```

## Response Constructors

Module-level constructors (without context):

```avenir
coolweb.textResponse("body", 200)
coolweb.jsonResponse(data, 200)
coolweb.htmlResponse("<h1>Hi</h1>", 200)
coolweb.redirectResponse("/other", 302)
coolweb.fileResponse("public/logo.png", 200)
```

## Module Structure

```
std/coolweb/
    coolweb.av      App, newApp, decorator methods, run(), dispatch
    router.av       Router, newRouter, route registration, resolve
    context.av      Context, response builders, body parsers
    response.av     Response, textResponse, jsonResponse, htmlResponse, redirectResponse, fileResponse
    request.av      Request
    route.av        Route, compileRoute, matchRoute
    middleware.av    executeChain
    utils.av        parseQueryString, parseCookieHeader
    cors.av         CorsConfig, corsMiddleware
    logger.av       loggerMiddleware
    errors.av       notFoundError, methodNotAllowedError, internalError
```
