# std.websocket

WebSocket support for Avenir. Provides async, non-blocking WebSocket connections compatible with the VM scheduler.

## Import

```avenir
import std.websocket as ws;
```

## Constants

| Name | Value | Description |
|------|-------|-------------|
| `MSG_TEXT` | 1 | Text message |
| `MSG_BINARY` | 2 | Binary message |
| `MSG_CLOSE` | 8 | Close frame |
| `MSG_PING` | 9 | Ping frame |
| `MSG_PONG` | 10 | Pong frame |

## Structs

### WebSocket

```avenir
pub mut struct WebSocket {
    pub id | string
    pub path | string
    pub headers | dict<string>
    pub query | string
    pub remoteAddr | string
    pub mut isOpen | bool
    pub _handle | any
}
```

### Message

```avenir
pub struct Message {
    pub msgType | int
    pub text | string
    pub data | bytes
    pub closeCode | int
}
```

**Methods:**

- `isText() | bool`
- `isBinary() | bool`
- `isClose() | bool`
- `isPing() | bool`
- `isPong() | bool`

### Hub

Optional broadcast helper for managing multiple connections.

```avenir
pub mut struct Hub {
    pub mut connections | list<any>
}
```

## Functions

### upgradeRequest

Upgrades an HTTP connection to WebSocket.

```avenir
pub async fun upgradeRequest(reqHandle | any, protocols | list<string> = [], headers | dict<string> = {}) | WebSocket
```

### newHub

```avenir
pub fun newHub() | Hub
```

## WebSocket Methods

### sendText

```avenir
pub async fun (ws | WebSocket).sendText(text | string) | void
```

### sendBytes

```avenir
pub async fun (ws | WebSocket).sendBytes(data | bytes) | void
```

### sendPing

```avenir
pub async fun (ws | WebSocket).sendPing(data | bytes = fromString("")) | void
```

### receive

Reads the next message. Returns a `Message` struct. When the peer sends a close frame, `msg.isClose()` returns `true` and `ws.isOpen` is set to `false`.

```avenir
pub async fun (ws | WebSocket).receive() | Message
```

### close

Sends a close frame and shuts down the connection.

```avenir
pub async fun (ws | WebSocket).close(code | int = 1000, reason | string = "") | void
```

### setReadLimit

Sets the maximum message size in bytes (default 1MB).

```avenir
pub fun (ws | WebSocket).setReadLimit(limit | int) | void
```

## Hub Methods

### add / remove

```avenir
pub fun (h | Hub).add(ws | WebSocket) | void
pub fun (h | Hub).remove(ws | WebSocket) | void
```

### broadcast / broadcastBytes

Sends a message to all open connections. Silently marks failed connections as closed.

```avenir
pub async fun (h | Hub).broadcast(text | string) | void
pub async fun (h | Hub).broadcastBytes(data | bytes) | void
```

### count / cleanup

```avenir
pub fun (h | Hub).count() | int
pub fun (h | Hub).cleanup() | void
```

`cleanup()` removes closed connections from the hub.

## Error Types

```avenir
pub struct WebSocketError {
    pub message | string
    pub code | int
}

pub struct WebSocketClosed {
    pub message | string
    pub code | int
    pub reason | string
}

pub struct ProtocolError {
    pub message | string
}
```

## CoolWeb Integration

### Registering WebSocket routes

```avenir
import std.coolweb as cw;
import std.websocket as ws;

var app | cw.App = cw.newApp();

@app.websocket("/echo")
async fun echo(ctx | cw.Context, conn | ws.WebSocket) | void {
    while (conn.isOpen) {
        var msg | ws.Message = await conn.receive();
        if (msg.isText()) {
            await conn.sendText("echo: " + msg.text);
        }
    }
}
```

### Router-level WebSocket routes

```avenir
var chatRouter | cw.Router = cw.newRouter("/chat");

@chatRouter.websocket("/room/:id")
async fun chatRoom(ctx | cw.Context, conn | ws.WebSocket) | void {
    var roomId | string = ctx.params["id"];
    // ...
}

app.mount(chatRouter);
```

### Middleware

WebSocket routes inherit middleware from the router. Middleware executes during the HTTP upgrade phase, before the WebSocket handler is called.

## Examples

- `examples/coolweb/websocket_echo.av` — Echo server
- `examples/coolweb/websocket_chat.av` — Chat room with Hub broadcasting

## Protocol Details

- Implements RFC 6455
- Automatic ping/pong handling (pong is auto-sent for received pings)
- Fragmentation support (continuation frames assembled automatically)
- Client-to-server frame unmasking
- Close handshake with configurable status code and reason
- Default max message size: 1MB (configurable via `setReadLimit`)
- Write deadline: 30s per frame (prevents zombie connections)
