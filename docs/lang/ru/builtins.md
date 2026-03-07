# Встроенные функции и методы

Avenir предоставляет встроенные функции и методы для типовых операций.
При ошибке встроенные функции выбрасывают исключения времени выполнения; используйте `try / catch`
для обработки ошибок ввода-вывода и преобразований.

## Встроенные функции

| Функция | Параметры | Возвращает | Ошибки |
| --- | --- | --- | --- |
| `print` | `value | any` | `any` | ошибки I/O |
| `input` | — | `string` | ошибки I/O |
| `len` | `value | list<any> \| bytes` | `int` | неверный тип |
| `typeOf` | `value | any` | `string` | некорректное значение времени выполнения |
| `toInt` | `value | string` | `int` | неверное целое число |
| `error` | `message | string` | `error` | — |
| `errorMessage` | `e | error` | `string` | — |
| `fromString` | `s | string` | `bytes` | — |

### `print(value | any) | any`

Печатает значение в стандартный вывод и возвращает это же значение.

```avenir
print("Hello, World!");
print(42);
```

### `input() | string`

Читает строку из стандартного ввода. Символы перевода строки по краям удаляются.
На EOF возвращает пустую строку.

```avenir
var name | string = input();
```

### `len(value | any) | int`

Возвращает длину списка или значения типа `bytes`. Если аргумент имеет другой
тип, выбрасывается ошибка времени выполнения.

```avenir
var length | int = len([1, 2, 3]);
var size | int = len(b"bytes");
```

### `typeOf(value | any) | string`

Возвращает каноническое имя типа Avenir для значения времени выполнения.
Для значений интерфейсного типа `typeOf()` возвращает конкретный тип времени выполнения.
Списки отображаются как `list<T...>`, а опциональные типы — как `T?`.

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

Преобразует десятичную строку, например `"123"` или `"-42"`, в целое число.
Если строка не является корректным целым числом, выбрасывается `error`.

```avenir
var ok | int = toInt("123");
try {
    var bad | int = toInt("hello");
} catch (e | error) {
    print(errorMessage(e));
}
```

### `error(message | string) | error`

Создаёт значение ошибки.

```avenir
var e | error = error("что-то пошло не так");
```

### `errorMessage(e | error) | string`

Извлекает текст сообщения из значения ошибки.

```avenir
var msg | string = errorMessage(e);
```

### `fromString(s | string) | bytes`

Преобразует строку в `bytes`.

```avenir
var data | bytes = fromString("hello");
```

## Методы списков

Списки поддерживают следующие методы:

> Примечание: методы списков обычно возвращают новый список, а не изменяют
> исходный, если не указано иное.

| Метод | Параметры | Возвращает | Примечания |
| --- | --- | --- | --- |
| `append` | `element | any` | `list<any>` | возвращает новый список |
| `length` | — | `int` | число элементов |
| `get` | `index | int` | `any` | ошибка при выходе за границы |
| `contains` | `element | any` | `bool` | глубокое сравнение |
| `indexOf` | `element | any` | `int` | `-1`, если не найдено |
| `slice` | `start | int`, `end | int` | `list<any>` | `end` не включается |
| `reverse` | — | `list<any>` | возвращает новый список |
| `copy` | — | `list<any>` | поверхностная копия |
| `pop` | — | `any` | возвращает последний элемент |
| `insert` | `index | int`, `element | any` | `list<any>` | возвращает новый список |
| `removeAt` | `index | int` | `list<any>` | возвращает новый список |
| `clear` | — | `list<any>` | пустой список |
| `isEmpty` | — | `bool` | — |
| `map` | `fn | fun(any) | any` | `list<any>` | вызывает функцию для каждого элемента |
| `filter` | `fn | fun(any) | bool` | `list<any>` | вызывает предикат для каждого элемента |
| `reduce` | `initial | any`, `reducer | fun(any, any) | any` | `any` | аккумулятор |

### `append(element | any) | list<any>`

Добавляет элемент в список и возвращает новый список.

```avenir
var list | list<int> = [1, 2];
list = list.append(3);  // [1, 2, 3]
```

### `length() | int`

Возвращает длину списка.

```avenir
var len | int = list.length();
```

### `get(index | int) | any`

Возвращает элемент по индексу. При выходе за границы выбрасывается ошибка.

```avenir
var item | int = list.get(0);
```

### `contains(element | any) | bool`

Проверяет, содержит ли список элемент.

```avenir
var found | bool = list.contains(42);
```

### `indexOf(element | any) | int`

Возвращает индекс первого вхождения элемента или `-1`, если элемент не найден.

```avenir
var idx | int = list.indexOf(42);
```

### `slice(start | int, end | int) | list<any>`

Возвращает срез списка от `start` включительно до `end` не включительно.

```avenir
var sub | list<int> = list.slice(1, 3);
```

### `reverse() | list<any>`

Возвращает перевёрнутую копию списка.

```avenir
var reversed | list<int> = list.reverse();
```

### `copy() | list<any>`

Возвращает копию списка.

```avenir
var copied | list<int> = list.copy();
```

### `pop() | any`

Возвращает последний элемент списка. Для пустого списка выбрасывает ошибку.

```avenir
var last | int = list.pop();
```

### `insert(index | int, element | any) | list<any>`

Вставляет элемент по указанному индексу и возвращает новый список. При выходе
за границы выбрасывает ошибку.

```avenir
list = list.insert(1, 42);
```

### `removeAt(index | int) | list<any>`

Удаляет элемент по указанному индексу и возвращает новый список. При выходе за
границы выбрасывает ошибку.

```avenir
list = list.removeAt(1);
```

### `clear() | list<any>`

Возвращает пустой список того же типа.

```avenir
list = list.clear();
```

### `isEmpty() | bool`

Проверяет, пуст ли список.

```avenir
var empty | bool = list.isEmpty();
```

### `map(fn | fun (any) | any) | list<any>`

Применяет функцию к каждому элементу и возвращает новый список.
Если `fn` не является функцией или сама функция выбрасывает ошибку, возникает ошибка времени выполнения.

```avenir
var doubled | list<int> = list.map(fun (x | int) | int {
    return x * 2;
});
```

### `filter(fn | fun (any) | bool) | list<any>`

Фильтрует элементы с помощью предиката и возвращает новый список.
Если `fn` не является функцией или сама функция выбрасывает ошибку, возникает ошибка времени выполнения.

```avenir
var evens | list<int> = list.filter(fun (x | int) | bool {
    return x % 2 == 0;
});
```

### `reduce(initial | any, reducer | fun (any, any) | any) | any`

Сворачивает список в одно значение с помощью функции-редуктора.
Если `reducer` не является функцией или сама функция выбрасывает ошибку, возникает ошибка времени выполнения.

```avenir
var sum | int = list.reduce(0, fun (acc | int, x | int) | int {
    return acc + x;
});
```

## Методы строк

Строки поддерживают следующие методы:

| Метод | Параметры | Возвращает | Примечания |
| --- | --- | --- | --- |
| `length` | — | `int` | длина в байтах UTF-8 |
| `toUpperCase` | — | `string` | — |
| `toLowerCase` | — | `string` | — |
| `trim` | — | `string` | убирает пробелы по краям |
| `trimLeft` | — | `string` | убирает пробелы слева |
| `trimRight` | — | `string` | убирает пробелы справа |
| `contains` | `substr | string` | `bool` | — |
| `startsWith` | `prefix | string` | `bool` | — |
| `endsWith` | `suffix | string` | `bool` | — |
| `replace` | `old | string`, `new | string` | `string` | заменяет все вхождения |
| `split` | `sep | string` | `list<string>` | — |
| `indexOf` | `substr | string` | `int` | `-1`, если не найдено |
| `lastIndexOf` | `substr | string` | `int` | `-1`, если не найдено |

### `length() | int`

Возвращает длину строки в байтах UTF-8.

```avenir
var len | int = str.length();
```

### `toUpperCase() | string`

Возвращает строку в верхнем регистре.

```avenir
var upper | string = str.toUpperCase();
```

### `toLowerCase() | string`

Возвращает строку в нижнем регистре.

```avenir
var lower | string = str.toLowerCase();
```

### `trim() | string`

Возвращает копию строки без пробелов по краям.

```avenir
var trimmed | string = str.trim();
```

### `trimLeft() | string`

Возвращает копию строки без пробелов слева.

```avenir
var trimmed | string = str.trimLeft();
```

### `trimRight() | string`

Возвращает копию строки без пробелов справа.

```avenir
var trimmed | string = str.trimRight();
```

### `contains(substr | string) | bool`

Проверяет, содержит ли строка подстроку.

```avenir
var found | bool = str.contains("hello");
```

### `startsWith(prefix | string) | bool`

Проверяет, начинается ли строка с указанного префикса.

```avenir
var matches | bool = str.startsWith("http");
```

### `endsWith(suffix | string) | bool`

Проверяет, заканчивается ли строка указанным суффиксом.

```avenir
var matches | bool = str.endsWith(".av");
```

### `replace(old | string, new | string) | string`

Заменяет все вхождения `old` на `new`.

```avenir
var replaced | string = str.replace("old", "new");
```

### `split(sep | string) | list<string>`

Разбивает строку по разделителю и возвращает список строк.

```avenir
var parts | list<string> = str.split(",");
```

### `indexOf(substr | string) | int`

Возвращает индекс первого вхождения подстроки или `-1`, если она не найдена.

```avenir
var idx | int = str.indexOf("hello");
```

### `lastIndexOf(substr | string) | int`

Возвращает индекс последнего вхождения подстроки или `-1`, если она не найдена.

```avenir
var idx | int = str.lastIndexOf("hello");
```

## Методы `bytes`

Значения `bytes` поддерживают следующие методы:

| Метод | Параметры | Возвращает | Примечания |
| --- | --- | --- | --- |
| `length` | — | `int` | количество байтов |
| `append` | `b | bytes` | `bytes` | возвращает новое значение |
| `concat` | `b | bytes` | `bytes` | возвращает новое значение |
| `slice` | `start | int`, `end | int` | `bytes` | `end` не включается |
| `toString` | — | `string` | UTF-8 decode, ошибка на невалидных данных |

### `length() | int`

Возвращает длину значения `bytes`.

```avenir
var len | int = data.length();
```

### `append(b | bytes) | bytes`

Добавляет байты и возвращает новое значение `bytes`.

```avenir
data = data.append(b"more");
```

### `concat(b | bytes) | bytes`

Конкатенирует два значения `bytes` и возвращает новый результат.

```avenir
var combined | bytes = data.concat(b"more");
```

### `slice(start | int, end | int) | bytes`

Возвращает срез байтов от `start` включительно до `end` не включительно.

```avenir
var sub | bytes = data.slice(0, 10);
```

### `toString() | string`

Преобразует `bytes` в строку UTF-8. Если данные не являются корректным UTF-8, выбрасывается ошибка.

```avenir
var str | string = data.toString();
```

## Методы словарей

Словари поддерживают следующие методы:

| Метод | Параметры | Возвращает | Примечания |
| --- | --- | --- | --- |
| `length` | — | `int` | число записей |
| `keys` | — | `list<string>` | порядок не гарантирован |
| `values` | — | `list<any>` | порядок не гарантирован |
| `has` | `key | string` | `bool` | проверка наличия |
| `get` | `key | string` | `any?` | `none`, если ключ отсутствует |
| `set` | `key | string`, `value | any` | `void` | изменяет словарь на месте |
| `remove` | `key | string` | `bool` | возвращает, существовал ли ключ |

### `length() | int`

Возвращает количество записей в словаре.

### `keys() | list<string>`

Возвращает ключи словаря. Порядок не гарантирован.

### `values() | list<any>`

Возвращает значения словаря. Порядок не гарантирован.

### `has(key | string) | bool`

Проверяет, присутствует ли ключ в словаре.

### `get(key | string) | any?`

Возвращает значение как optional. Если ключ отсутствует, возвращается `none`.

### `set(key | string, value | any) | void`

Добавляет или обновляет ключ. Изменяет словарь на месте.

### `remove(key | string) | bool`

Удаляет ключ и возвращает, существовал ли он.
