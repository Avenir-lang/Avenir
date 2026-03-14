# std.sql — Асинхронний SQL-клієнт

Модуль `std.sql` надає асинхронну бібліотеку для роботи з SQL-базами даних.

## Імпорт

```avenir
import std.sql;
import std.sql.postgres;  // для PostgreSQL
```

## Швидкий старт

### PostgreSQL

```avenir
import std.sql;
import std.sql.postgres;

async fun main() | void {
    var db | std.sql.DB = await std.sql.open("postgres", "host=localhost dbname=mydb");

    var rows | std.sql.Rows = await db.query("SELECT id, name FROM users WHERE age > $1", [18]);
    while (rows.next()) {
        var row | std.sql.Row = rows.current();
        var id | int = row.getInt("id");
        var name | string = row.getString("name");
        print("${id}: ${name}");
    }

    await db.close();
}
```

### SQLite

```avenir
import std.sql;

async fun main() | void {
    var db | std.sql.DB = await std.sql.open("sqlite", "mydb.db");
    await db.exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)");
    await db.exec("INSERT INTO users (name) VALUES ($1)", ["Аліса"]);
    await db.close();
}
```

## DB

### Створення

```avenir
var db | std.sql.DB = await std.sql.open(driver | string, dsn | string);
```

### Методи

| Метод | Опис |
|-------|------|
| `query(sql \| string, params \| list<any>) \| Future<Rows>` | Виконати запит з результатами |
| `exec(sql \| string, params \| list<any>) \| Future<ExecResult>` | Виконати без результатів |
| `begin() \| Future<Transaction>` | Розпочати транзакцію |
| `close() \| Future<void>` | Закрити з'єднання |

## ExecResult

```avenir
struct ExecResult {
    rowsAffected | int
    lastInsertId | int
}
```

## Rows

### Ітерація

```avenir
var rows | std.sql.Rows = await db.query("SELECT * FROM users");
while (rows.next()) {
    var row | std.sql.Row = rows.current();
    // обробка рядка
}
```

### Методи

| Метод | Опис |
|-------|------|
| `next() \| bool` | Перейти до наступного рядка |
| `current() \| Row` | Отримати поточний рядок |
| `columns() \| list<string>` | Отримати імена колонок |
| `count() \| int` | Кількість рядків |

## Row

| Метод | Опис |
|-------|------|
| `getString(col \| string) \| string` | Отримати рядкове значення |
| `getInt(col \| string) \| int` | Отримати цілочисельне значення |
| `getFloat(col \| string) \| float` | Отримати дробове значення |
| `getBool(col \| string) \| bool` | Отримати булеве значення |
| `isNull(col \| string) \| bool` | Перевірити на NULL |

## Транзакції

```avenir
var tx | std.sql.Transaction = await db.begin();
try {
    await tx.exec("INSERT INTO users (name) VALUES ($1)", ["Боб"]);
    await tx.exec("UPDATE counters SET count = count + 1");
    await tx.commit();
} catch (e | error) {
    await tx.rollback();
    print("Помилка транзакції: " + e.message());
}
```

### Методи Transaction

| Метод | Опис |
|-------|------|
| `query(sql \| string, params \| list<any>) \| Future<Rows>` | Запит у транзакції |
| `exec(sql \| string, params \| list<any>) \| Future<ExecResult>` | Виконання у транзакції |
| `commit() \| Future<void>` | Зафіксувати транзакцію |
| `rollback() \| Future<void>` | Відкотити транзакцію |

## Пул з'єднань

```avenir
var db | std.sql.DB = await std.sql.open("postgres", dsn);
db.setMaxOpenConns(10);
db.setMaxIdleConns(5);
```

## Прив'язка параметрів

Параметри прив'язуються через `$1`, `$2`, ... :

```avenir
await db.query("SELECT * FROM users WHERE age > $1 AND name = $2", [18, "Аліса"]);
```

## Система драйверів

### Реєстрація

Драйвери реєструються автоматично при імпорті:

```avenir
import std.sql.postgres;  // Авто-реєстрація PostgreSQL-драйвера
```

### Кастомні драйвери

```avenir
var driver | std.sql.Driver = std.sql.newDriver(
    name = "mydb",
    connectFn = ...,
    queryFn = ...,
    execFn = ...,
    closeFn = ...
);
std.sql.registerDriver(driver);
```

## Помилки

SQL-операції можуть викидати помилки:

```avenir
try {
    var rows | std.sql.Rows = await db.query("INVALID SQL");
} catch (e | error) {
    print("SQL помилка: " + e.message());
}
```

## Інтеграція з CoolWeb

```avenir
import std.coolweb;
import std.sql;
import std.sql.postgres;

var app | std.coolweb.App = std.coolweb.newApp();
var db | std.sql.DB = await std.sql.open("postgres", "host=localhost dbname=mydb");

@app.get("/users")
async fun listUsers(ctx | std.coolweb.Context) | void {
    var rows | std.sql.Rows = await db.query("SELECT name FROM users");
    var names | list<string> = [];
    while (rows.next()) {
        var row | std.sql.Row = rows.current();
        names.push(row.getString("name"));
    }
    ctx.json(std.json.stringify(names));
}
```

## Структура модуля

```
std/sql/
├── sql.av           # Головний API: open(), registerDriver()
├── connection.av    # DB з'єднання
├── pool.av          # Пул з'єднань
├── rows.av          # Rows ітератор
├── row.av           # Row доступ до значень
├── transaction.av   # Транзакції
├── driver.av        # Інтерфейс драйвера
├── errors.av        # Типи помилок
├── utils.av         # Утиліти
└── postgres/        # PostgreSQL драйвер
    ├── postgres.av
    ├── protocol.av
    └── types.av
```
