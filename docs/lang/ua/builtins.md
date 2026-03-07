# Вбудовані функції та методи

Avenir надає вбудовані функції та методи для типових операцій.
У разі помилки вбудовані функції викидають винятки часу виконання; використовуйте `try / catch`
для обробки помилок вводу-виводу та перетворень.

## Вбудовані функції

| Функція | Параметри | Повертає | Помилки |
| --- | --- | --- | --- |
| `print` | `value | any` | `any` | помилки I/O |
| `input` | — | `string` | помилки I/O |
| `len` | `value | list<any> \| bytes` | `int` | неправильний тип |
| `typeOf` | `value | any` | `string` | некоректне значення часу виконання |
| `toInt` | `value | string` | `int` | некоректне ціле число |
| `error` | `message | string` | `error` | — |
| `errorMessage` | `e | error` | `string` | — |
| `fromString` | `s | string` | `bytes` | — |

### `print(value | any) | any`

Друкує значення у стандартний вивід і повертає це ж значення.

```avenir
print("Hello, World!");
print(42);
```

### `input() | string`

Читає рядок зі стандартного вводу. Символи нового рядка по краях видаляються.
На EOF повертає порожній рядок.

```avenir
var name | string = input();
```

### `len(value | any) | int`

Повертає довжину списку або значення типу `bytes`. Якщо аргумент має інший
тип, викидається помилка часу виконання.

```avenir
var length | int = len([1, 2, 3]);
var size | int = len(b"bytes");
```

### `typeOf(value | any) | string`

Повертає канонічне ім'я типу Avenir для значення часу виконання.
Для значень інтерфейсного типу `typeOf()` повертає конкретний тип часу виконання.
Списки відображаються як `list<T...>`, а опціональні типи — як `T?`.

```avenir
typeOf(10);            // "int"
typeOf(1.5);           // "float"
typeOf("hello");       // "string"
typeOf(true);          // "bool"
typeOf([]);            // "list<any>"
typeOf([1, 2]);        // "list<int>"
typeOf(fromString("abc")); // "bytes"
typeOf(error("x"));    // "error"
typeOf(Point{x = 1, y = 2}); // "Point"
typeOf(some(1));       // "int?"
typeOf(none);          // "any?"
```

### `toInt(value | string) | int`

Перетворює десятковий рядок, наприклад `"123"` або `"-42"`, на ціле число.
Якщо рядок не є коректним цілим числом, викидається `error`.

```avenir
var ok | int = toInt("123");
try {
    var bad | int = toInt("hello");
} catch (e | error) {
    print(errorMessage(e));
}
```

### `error(message | string) | error`

Створює значення помилки.

```avenir
var e | error = error("щось пішло не так");
```

### `errorMessage(e | error) | string`

Витягує текст повідомлення зі значення помилки.

```avenir
var msg | string = errorMessage(e);
```

### `fromString(s | string) | bytes`

Перетворює рядок на `bytes`.

```avenir
var data | bytes = fromString("hello");
```

## Методи списків

Списки підтримують такі методи:

> Примітка: методи списків зазвичай повертають новий список, а не змінюють
> початковий, якщо не вказано інше.

| Метод | Параметри | Повертає | Примітки |
| --- | --- | --- | --- |
| `append` | `element | any` | `list<any>` | повертає новий список |
| `length` | — | `int` | кількість елементів |
| `get` | `index | int` | `any` | помилка при виході за межі |
| `contains` | `element | any` | `bool` | глибоке порівняння |
| `indexOf` | `element | any` | `int` | `-1`, якщо не знайдено |
| `slice` | `start | int`, `end | int` | `list<any>` | `end` не включається |
| `reverse` | — | `list<any>` | повертає новий список |
| `copy` | — | `list<any>` | поверхнева копія |
| `pop` | — | `any` | повертає останній елемент |
| `insert` | `index | int`, `element | any` | `list<any>` | повертає новий список |
| `removeAt` | `index | int` | `list<any>` | повертає новий список |
| `clear` | — | `list<any>` | порожній список |
| `isEmpty` | — | `bool` | — |
| `map` | `fn | fun(any) | any` | `list<any>` | викликає функцію для кожного елемента |
| `filter` | `fn | fun(any) | bool` | `list<any>` | викликає предикат для кожного елемента |
| `reduce` | `initial | any`, `reducer | fun(any, any) | any` | `any` | акумулятор |

### `append(element | any) | list<any>`

Додає елемент до списку та повертає новий список.

```avenir
var list | list<int> = [1, 2];
list = list.append(3);  // [1, 2, 3]
```

### `length() | int`

Повертає довжину списку.

```avenir
var len | int = list.length();
```

### `get(index | int) | any`

Повертає елемент за індексом. При виході за межі викидає помилку.

```avenir
var item | int = list.get(0);
```

### `contains(element | any) | bool`

Перевіряє, чи містить список елемент.

```avenir
var found | bool = list.contains(42);
```

### `indexOf(element | any) | int`

Повертає індекс першого входження елемента або `-1`, якщо елемент не знайдено.

```avenir
var idx | int = list.indexOf(42);
```

### `slice(start | int, end | int) | list<any>`

Повертає зріз списку від `start` включно до `end` не включно.

```avenir
var sub | list<int> = list.slice(1, 3);
```

### `reverse() | list<any>`

Повертає перевернуту копію списку.

```avenir
var reversed | list<int> = list.reverse();
```

### `copy() | list<any>`

Повертає копію списку.

```avenir
var copied | list<int> = list.copy();
```

### `pop() | any`

Повертає останній елемент списку. Для порожнього списку викидає помилку.

```avenir
var last | int = list.pop();
```

### `insert(index | int, element | any) | list<any>`

Вставляє елемент за вказаним індексом і повертає новий список. При виході за
межі викидає помилку.

```avenir
list = list.insert(1, 42);
```

### `removeAt(index | int) | list<any>`

Видаляє елемент за вказаним індексом і повертає новий список. При виході за
межі викидає помилку.

```avenir
list = list.removeAt(1);
```

### `clear() | list<any>`

Повертає порожній список того ж типу.

```avenir
list = list.clear();
```

### `isEmpty() | bool`

Перевіряє, чи порожній список.

```avenir
var empty | bool = list.isEmpty();
```

### `map(fn | fun (any) | any) | list<any>`

Застосовує функцію до кожного елемента і повертає новий список.
Якщо `fn` не є функцією або сама функція викидає помилку, виникає помилка часу виконання.

```avenir
var doubled | list<int> = list.map(fun (x | int) | int {
    return x * 2;
});
```

### `filter(fn | fun (any) | bool) | list<any>`

Фільтрує елементи за допомогою предиката і повертає новий список.
Якщо `fn` не є функцією або сама функція викидає помилку, виникає помилка часу виконання.

```avenir
var evens | list<int> = list.filter(fun (x | int) | bool {
    return x % 2 == 0;
});
```

### `reduce(initial | any, reducer | fun (any, any) | any) | any`

Згортає список в одне значення за допомогою функції-редуктора.
Якщо `reducer` не є функцією або сама функція викидає помилку, виникає помилка часу виконання.

```avenir
var sum | int = list.reduce(0, fun (acc | int, x | int) | int {
    return acc + x;
});
```

## Методи рядків

Рядки підтримують такі методи:

| Метод | Параметри | Повертає | Примітки |
| --- | --- | --- | --- |
| `length` | — | `int` | довжина в байтах UTF-8 |
| `toUpperCase` | — | `string` | — |
| `toLowerCase` | — | `string` | — |
| `trim` | — | `string` | прибирає пробіли по краях |
| `trimLeft` | — | `string` | прибирає пробіли зліва |
| `trimRight` | — | `string` | прибирає пробіли справа |
| `contains` | `substr | string` | `bool` | — |
| `startsWith` | `prefix | string` | `bool` | — |
| `endsWith` | `suffix | string` | `bool` | — |
| `replace` | `old | string`, `new | string` | `string` | замінює всі входження |
| `split` | `sep | string` | `list<string>` | — |
| `indexOf` | `substr | string` | `int` | `-1`, якщо не знайдено |
| `lastIndexOf` | `substr | string` | `int` | `-1`, якщо не знайдено |

### `length() | int`

Повертає довжину рядка в байтах UTF-8.

```avenir
var len | int = str.length();
```

### `toUpperCase() | string`

Повертає рядок у верхньому регістрі.

```avenir
var upper | string = str.toUpperCase();
```

### `toLowerCase() | string`

Повертає рядок у нижньому регістрі.

```avenir
var lower | string = str.toLowerCase();
```

### `trim() | string`

Повертає копію рядка без пробілів по краях.

```avenir
var trimmed | string = str.trim();
```

### `trimLeft() | string`

Повертає копію рядка без пробілів зліва.

```avenir
var trimmed | string = str.trimLeft();
```

### `trimRight() | string`

Повертає копію рядка без пробілів справа.

```avenir
var trimmed | string = str.trimRight();
```

### `contains(substr | string) | bool`

Перевіряє, чи містить рядок підрядок.

```avenir
var found | bool = str.contains("hello");
```

### `startsWith(prefix | string) | bool`

Перевіряє, чи починається рядок із заданого префікса.

```avenir
var matches | bool = str.startsWith("http");
```

### `endsWith(suffix | string) | bool`

Перевіряє, чи закінчується рядок заданим суфіксом.

```avenir
var matches | bool = str.endsWith(".av");
```

### `replace(old | string, new | string) | string`

Замінює всі входження `old` на `new`.

```avenir
var replaced | string = str.replace("old", "new");
```

### `split(sep | string) | list<string>`

Розбиває рядок за роздільником і повертає список рядків.

```avenir
var parts | list<string> = str.split(",");
```

### `indexOf(substr | string) | int`

Повертає індекс першого входження підрядка або `-1`, якщо його не знайдено.

```avenir
var idx | int = str.indexOf("hello");
```

### `lastIndexOf(substr | string) | int`

Повертає індекс останнього входження підрядка або `-1`, якщо його не знайдено.

```avenir
var idx | int = str.lastIndexOf("hello");
```

## Методи `bytes`

Значення `bytes` підтримують такі методи:

| Метод | Параметри | Повертає | Примітки |
| --- | --- | --- | --- |
| `length` | — | `int` | кількість байтів |
| `append` | `b | bytes` | `bytes` | повертає нове значення |
| `concat` | `b | bytes` | `bytes` | повертає нове значення |
| `slice` | `start | int`, `end | int` | `bytes` | `end` не включається |
| `toString` | — | `string` | UTF-8 decode, помилка на невалідних даних |

### `length() | int`

Повертає довжину значення `bytes`.

```avenir
var len | int = data.length();
```

### `append(b | bytes) | bytes`

Додає байти і повертає нове значення `bytes`.

```avenir
data = data.append(b"more");
```

### `concat(b | bytes) | bytes`

Конкатенує два значення `bytes` і повертає новий результат.

```avenir
var combined | bytes = data.concat(b"more");
```

### `slice(start | int, end | int) | bytes`

Повертає зріз байтів від `start` включно до `end` не включно.

```avenir
var sub | bytes = data.slice(0, 10);
```

### `toString() | string`

Перетворює `bytes` на рядок UTF-8. Якщо дані не є коректним UTF-8, викидається помилка.

```avenir
var str | string = data.toString();
```

## Методи словників

Словники підтримують такі методи:

| Метод | Параметри | Повертає | Примітки |
| --- | --- | --- | --- |
| `length` | — | `int` | кількість записів |
| `keys` | — | `list<string>` | порядок не гарантовано |
| `values` | — | `list<any>` | порядок не гарантовано |
| `has` | `key | string` | `bool` | перевірка наявності |
| `get` | `key | string` | `any?` | `none`, якщо ключ відсутній |
| `set` | `key | string`, `value | any` | `void` | змінює словник на місці |
| `remove` | `key | string` | `bool` | повертає, чи існував ключ |

### `length() | int`

Повертає кількість записів у словнику.

### `keys() | list<string>`

Повертає ключі словника. Порядок не гарантовано.

### `values() | list<any>`

Повертає значення словника. Порядок не гарантовано.

### `has(key | string) | bool`

Перевіряє, чи присутній ключ у словнику.

### `get(key | string) | any?`

Повертає значення як optional. Якщо ключ відсутній, повертається `none`.

### `set(key | string, value | any) | void`

Додає або оновлює ключ. Змінює словник на місці.

### `remove(key | string) | bool`

Видаляє ключ і повертає, чи існував він.
