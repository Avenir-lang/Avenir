# Асинхронність і конкурентність

Avenir має вбудовану підтримку асинхронного програмування через `async`/`await` та кооперативний планувальник задач.

## Огляд

- `async fun` оголошує функцію, що виконується як конкурентна задача
- Виклик async-функції одразу повертає `Future<T>`
- `await` призупиняє поточну задачу до розв'язання future
- Кілька async-викликів можуть працювати конкурентно (неблокуючий I/O, таймери тощо)

## Асинхронні функції

```avenir
async fun fetchData() | string {
    var content | string = await asyncReadFile("data.txt");
    return content;
}
```

Оголошений тип повернення — це *внутрішній* тип. Викликач отримує `Future<string>`, але `return` усередині функції використовує внутрішній тип напряму.

## Асинхронні методи

Методи також можуть бути асинхронними. Обгортка типу повернення працює так само:

```avenir
pub struct HttpClient {
    base_url | string
}

pub async fun (client | HttpClient).get(path | string) | string {
    var url | string = client.base_url + path;
    return await asyncHttpGet(url);
}

// Використання
var client | HttpClient = HttpClient{base_url = "https://api.example.com"};
var data | Future<string> = client.get("/users");
var result | string = await data;
```

Асинхронні методи підпорядковуються тим самим правилам, що й async-функції:
- Методи екземпляра мають receiver як перший параметр (не обгорнутий у Future)
- Тип повернення обгортається у `Future<T>`
- `await` можна використовувати всередині тіла async-метода

## Future\<T\>

`Future<T>` — вбудований узагальнений тип, що представляє значення, яке стане доступним пізніше. Future можна зберігати у змінних:

```avenir
var f | Future<int> = compute(42);
```

Future розв'язується, коли породжена задача завершується. Використовуйте `await` для отримання результату:

```avenir
var result | int = await f;
```

Очікування вже розв'язаного future повертає результат негайно.

## Конкурентність

Виклик кількох async-функцій перед очікуванням будь-якої з них запускає задачі конкурентно:

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

Обидва завантаження відбуваються паралельно. Загальний час виконання приблизно `max(time_a, time_b)`, а не `time_a + time_b`.

## Модель виконання

Avenir використовує однопотоковий кооперативний планувальник з неблокуючим циклом подій:

1. Кожен виклик async-функції створює **дочірню задачу** з власним стеком
2. Задача потрапляє в **чергу готових**
3. Цикл подій виконує одну задачу за раз, доки вона не завершиться або не призупиниться (при `await` на незавершеному future)
4. Призупинені задачі «паркуються» до розв'язання їхнього future
5. Фонові I/O операції виконуються в Go-горутинах; після завершення відповідний future розв'язується і задача-очікувач переплановується

Ця модель запобігає гонкам за спільний стан, водночас забезпечуючи конкурентний I/O.

## Асинхронна стандартна бібліотека

Стандартна бібліотека надає async-варіанти операцій I/O:

### Час

```avenir
import std.time;

async fun example() | void {
    await std.time.asyncSleep(1000000000);  // 1 секунда в наносекундах
}
```

### Файлова система

```avenir
import std.fs;

async fun example() | void {
    var f | std.fs.File = await std.fs.asyncOpen("file.txt");
    var data | string = await f.asyncReadString();
    await f.asyncClose();
}
```

Async-функції FS: `asyncOpen`, `asyncExists`, `asyncRemove`, `asyncMkdir`.
Async-методи File: `asyncRead`, `asyncReadAll`, `asyncReadString`, `asyncWrite`, `asyncWriteString`, `asyncClose`.

### Мережа

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

Async-функції HTTP: `asyncRequest`, `asyncGet`, `asyncPost`, `asyncPut`, `asyncDelete`.

## Обробка помилок

Async-функції підтримують ту саму обробку помилок `try`/`catch`, що й синхронний код. Якщо async-операція зазнає невдачі, future відхиляється і `await` пробрасує помилку:

```avenir
async fun safeFetch() | string {
    try {
        return await asyncHttpGet("https://example.com");
    } catch (e | error) {
        return "fallback";
    }
}
```

## Таймаути

Використовуйте `withTimeout` зі `std.time`, щоб обмежити час очікування future:

```avenir
import std.time;

async fun main() | void {
    var f | Future<string> = fetchData();
    try {
        var result | string = await std.time.withTimeout(f, std.time.fromSeconds(5));
        print(result);
    } catch (e | error) {
        print("запит перевищив ліміт часу");
    }
}
```

Якщо future розв'язується до дедлайну, `withTimeout` повертає результат. Інакше викидається помилка таймауту. Семантика «перший-записує-виграє» гарантує, що пізнє розв'язання оригінального future безпечно ігнорується.

Також можна використовувати вбудовану функцію безпосередньо з наносекундами:

```avenir
var result | int = await __builtin_async_with_timeout(future, 5000000000);
```

## Правила

- `await` можна використовувати лише всередині `async fun`
- `Future<T>` — єдиний тип, що підтримує `await`
- Async-функції не можна викликати через `spawn` — конкурентність автоматична при виклику `async fun`
- Функція `main` може бути `async`
