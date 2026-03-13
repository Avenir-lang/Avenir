# std.sql

Async SQL client library with connection pooling, parameterized queries, and transaction support.

## Quick Start

```avenir
pckg main;

import std.sql;
import std.sql.postgres;

async fun main() | void {
    postgres.register();

    var db | sql.DB = sql.open("postgres://user:pass@localhost:5432/mydb");

    var rows | sql.Rows = await db.query(
        "SELECT id, name FROM users WHERE age > $1",
        [18]
    );

    while (rows.next()) {
        var id | int = rows.getInt("id");
        var name | string = rows.getString("name");
        print("${id}: ${name}");
    }

    await db.close();
}
```

## DB

The primary entry point. Wraps a connection pool and provides query methods.

### Creating a DB

```avenir
// Lazy pool — connections established on demand
var db | sql.DB = sql.open(
    "postgres://user:pass@localhost:5432/mydb",
    maxConnections = 10
);

// Single async connection
var db | sql.DB = await sql.connect(
    "postgres://user:pass@localhost:5432/mydb"
);
```

### Methods

| Method | Parameters | Returns | Description |
| --- | --- | --- | --- |
| `query` | `sqlText \| string`, `params \| list<any>` | `Rows` | Execute query, return rows |
| `queryRow` | `sqlText \| string`, `params \| list<any>` | `Row?` | Execute query, return first row or none |
| `exec` | `sqlText \| string`, `params \| list<any>` | `ExecResult` | Execute statement |
| `begin` | — | `Transaction` | Begin a transaction |
| `close` | — | `void` | Close all connections |
| `stats` | — | `dict<any>` | Pool statistics |

All query methods are `async` and automatically acquire/release pool connections.

## ExecResult

```avenir
pub struct ExecResult {
    pub rowsAffected | int
    pub lastInsertId | int
}
```

## Rows

Iterates over query results row by row.

```avenir
var rows | sql.Rows = await db.query("SELECT id, name FROM users", []);

while (rows.next()) {
    var id | int = rows.getInt("id");
    var name | string = rows.getString("name");
    print("${id}: ${name}");
}

rows.close();
```

### Methods

| Method | Parameters | Returns | Description |
| --- | --- | --- | --- |
| `next` | — | `bool` | Advance to next row |
| `getInt` | `col \| string` | `int` | Column value as int |
| `getString` | `col \| string` | `string` | Column value as string |
| `getBool` | `col \| string` | `bool` | Column value as bool |
| `getFloat` | `col \| string` | `float` | Column value as float |
| `getBytes` | `col \| string` | `bytes` | Column value as bytes |
| `getAny` | `col \| string` | `any` | Raw column value |
| `isNull` | `col \| string` | `bool` | Check if column is null |
| `row` | — | `Row` | Get current Row struct |
| `count` | — | `int` | Total number of rows |
| `close` | — | `void` | Release resources |

Type mismatch on accessors throws a runtime error.

## Row

Single row with typed column accessors. Same accessor methods as `Rows`:
`getInt()`, `getString()`, `getBool()`, `getFloat()`, `getBytes()`, `getAny()`, `isNull()`, `has()`.

```avenir
var row | sql.Row? = await db.queryRow(
    "SELECT email FROM users WHERE id = $1",
    [42]
);

if (row != none) {
    var r | sql.Row = row;
    var email | string = r.getString("email");
    print(email);
}
```

## Transactions

```avenir
var tx | sql.Transaction = await db.begin();

try {
    await tx.exec(
        "INSERT INTO users(name) VALUES($1)",
        ["Alice"]
    );

    await tx.exec(
        "INSERT INTO log(event) VALUES($1)",
        ["user created"]
    );

    await tx.commit();
} catch (e | error) {
    await tx.rollback();
}
```

### Transaction Methods

| Method | Parameters | Returns | Description |
| --- | --- | --- | --- |
| `query` | `sqlText \| string`, `params \| list<any>` | `Rows` | Query within tx |
| `queryRow` | `sqlText \| string`, `params \| list<any>` | `Row?` | Query single row within tx |
| `exec` | `sqlText \| string`, `params \| list<any>` | `ExecResult` | Execute within tx |
| `commit` | — | `void` | Commit the transaction |
| `rollback` | — | `void` | Rollback the transaction |

Calling `commit()` or `rollback()` twice throws a `transactionError`.

## Connection Pool

The pool is managed internally by `DB`. Configuration:

```avenir
var db | sql.DB = sql.open(
    "postgres://localhost:5432/app",
    maxConnections = 20
);
```

Pool behavior:
- Reuses idle connections
- Creates new connections on demand up to `maxConnections`
- Throws `poolExhaustedError` when limit reached and no idle connections
- `db.stats()` returns `{ "idle": int, "busy": int, "maxConnections": int, "closed": bool }`

## Parameter Binding

Always use parameterized queries. Never concatenate user input into SQL.

```avenir
// Safe — parameters are bound separately
await db.query("SELECT * FROM users WHERE id = $1", [userId]);

// PostgreSQL uses $1, $2, $3 style placeholders
await db.exec(
    "INSERT INTO users(name, email) VALUES($1, $2)",
    ["Alice", "alice@example.com"]
);
```

## Driver System

Drivers are registered at startup. The PostgreSQL driver ships with the library.

```avenir
import std.sql.postgres;

// Register before any database operations
postgres.register();
```

The `connect()` and `open()` functions resolve the driver from the URL scheme:
- `postgres://` or `postgresql://` → PostgreSQL driver
- `sqlite://` → SQLite driver (future)
- `mysql://` → MySQL driver (future)

### Custom Drivers

Implement the `sql.Driver` struct and call `sql.registerDriver()`:

```avenir
var myDriver | sql.Driver = sql.Driver{
    name = "mydb",
    asyncConnectFn = myConnect,
    asyncCloseFn = myClose,
    asyncQueryFn = myQuery,
    asyncExecFn = myExec,
    asyncBeginFn = myBegin,
    asyncCommitFn = myCommit,
    asyncRollbackFn = myRollback,
    parseResultFn = myParseResult
};

sql.registerDriver("mydb", myDriver);
```

## Errors

| Error | Description |
| --- | --- |
| `connectionError(msg)` | Connection failures |
| `queryError(msg)` | Query execution failures |
| `queryErrorWithSql(msg, sql)` | Query error with SQL context |
| `transactionError(msg)` | Transaction state errors |
| `driverError(msg)` | Driver configuration errors |
| `typeError(col, expected, actual)` | Column type mismatch |
| `columnNotFoundError(col)` | Column does not exist |
| `poolExhaustedError()` | Pool has no available connections |
| `closedError(resource)` | Operation on closed resource |

## Integration with coolweb

```avenir
pckg main;

import std.coolweb;
import std.sql;
import std.sql.postgres;

var app | coolweb.App = coolweb.newApp();
var db | sql.DB = sql.open("postgres://localhost:5432/app", maxConnections = 10);

@app.get("/users")
async fun listUsers(ctx | coolweb.Context) | coolweb.Response {
    var rows | sql.Rows = await db.query("SELECT id, name FROM users", []);
    var users | list<any> = [];

    while (rows.next()) {
        var uid | int = rows.getInt("id");
        var uname | string = rows.getString("name");
        users = users.append({
            "id": uid,
            "name": uname
        });
    }
    rows.close();

    return ctx.json(users);
}

@app.get("/users/:id")
async fun getUser(ctx | coolweb.Context) | coolweb.Response {
    var id | string = ctx.params["id"];
    var row | sql.Row? = await db.queryRow(
        "SELECT id, name, email FROM users WHERE id = $1",
        [id]
    );

    if (row == none) {
        return ctx.json({ "error": "not found" }, 404);
    }

    var r | sql.Row = row;
    var rid | int = r.getInt("id");
    var rname | string = r.getString("name");
    var remail | string = r.getString("email");
    return ctx.json({
        "id": rid,
        "name": rname,
        "email": remail
    });
}

async fun main() | void {
    postgres.register();
    await app.run(8080);
}
```

## Module Structure

```
std/sql/
    sql.av              DB, ExecResult, connect(), open()
    connection.av       Connection struct, query execution
    pool.av             Pool struct, acquire/release
    rows.av             Rows struct, iteration + typed accessors
    row.av              Row struct, typed column accessors
    transaction.av      Transaction struct, commit/rollback
    driver.av           Driver struct, registerDriver(), resolveDriverFromUrl()
    errors.av           Error factory functions
    utils.av            parseConnectionUrl(), escapeIdentifier()
    postgres/
        postgres.av     PostgreSQL driver, register()
        protocol.av     PG wire protocol constants and message builders
        types.av        PG OID type mapping
```
