# std.json — JSON

Модуль `std.json` надає парсинг та серіалізацію JSON.

## Імпорт

```avenir
import std.json;
```

## Відображення типів

| JSON | Avenir |
|------|--------|
| `object` | `dict<any>` |
| `array` | `list<any>` |
| `string` | `string` |
| `number` (ціле) | `int` |
| `number` (дробове) | `float` |
| `boolean` | `bool` |
| `null` | `none` |

## Функції API

| Функція | Опис |
|---------|------|
| `parse(s \| string) \| any` | Парсинг JSON-рядка в значення Avenir |
| `stringify(v \| any) \| string` | Серіалізація значення Avenir в JSON-рядок |

## Типізовані хелпери

| Функція | Опис |
|---------|------|
| `asDict(v \| any) \| dict<any>` | Привести до словника |
| `asList(v \| any) \| list<any>` | Привести до списку |
| `asString(v \| any) \| string` | Привести до рядка |
| `asInt(v \| any) \| int` | Привести до цілого |
| `asBool(v \| any) \| bool` | Привести до булевого |

## Опціональні пошуки

| Функція | Опис |
|---------|------|
| `getOpt(d \| dict<any>, key \| string) \| any?` | Безпечне отримання за ключем |

## Приклади

### Парсинг

```avenir
import std.json;

fun main() | void {
    var data | any = std.json.parse("{\"name\": \"Alice\", \"age\": 30}");
    var obj | dict<any> = std.json.asDict(data);
    var name | string = std.json.asString(obj.get("name"));
    print(name);
}
```

### Серіалізація

```avenir
import std.json;

fun main() | void {
    var obj | dict<any> = {name: "Bob", age: 25};
    var s | string = std.json.stringify(obj);
    print(s);
}
```

### Вкладені об'єкти

```avenir
import std.json;

fun main() | void {
    var raw | string = "{\"user\": {\"name\": \"Charlie\"}, \"scores\": [10, 20]}";
    var data | dict<any> = std.json.asDict(std.json.parse(raw));

    var user | dict<any> = std.json.asDict(data.get("user"));
    print(std.json.asString(user.get("name")));

    var scores | list<any> = std.json.asList(data.get("scores"));
    print(std.json.asInt(scores.get(0)));
}
```

## Обробка помилок

`parse()` викидає помилку при невалідному JSON:

```avenir
try {
    var data | any = std.json.parse("invalid json");
} catch (e | error) {
    print("Помилка парсингу JSON: " + e.message());
}
```

## Хелпери помилок

| Функція | Опис |
|---------|------|
| `isParseError(e \| error) \| bool` | Перевірити, чи є помилка помилкою парсингу |

## Обмеження

- JSON-числа без дробової частини парсяться як `int`, з дробовою — як `float`
- `none` серіалізується як JSON `null`
- Серіалізація структур напряму не підтримується — конвертуйте в `dict` спочатку
