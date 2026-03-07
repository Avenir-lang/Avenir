# Швидкий старт

Цей посібник допоможе почати роботу з мовою програмування Avenir.

## Встановлення

Avenir реалізовано на Go. Щоб зібрати з вихідних кодів:

```bash
git clone <repository-url>
cd avenir
go build ./cmd/avenir
```

Зібраний бінарник надає CLI-команду `avenir`.

## Перша програма

Створіть файл `hello.av`:

```avenir
pckg main;

fun main() | void {
    print("Hello, Avenir!");
}
```

Кожна програма Avenir має починатися з `pckg` і містити функцію `main()`.

## Запуск програми

```bash
avenir run hello.av
```

Команда:

1. парсить вихідник
2. перевіряє типи
3. компілює у байткод
4. виконує програму у VM

## Збірка байткоду

```bash
avenir build hello.av -o hello.avc
```

Запуск `.avc` файла:

```bash
avenir run hello.avc
```

## CLI-команди

### `avenir run <file>`

Компілює і запускає `.av`, або запускає готовий `.avc`.

### `avenir build <file> [options]`

Компілює `.av` у байткод.

Опції:

- `-o <file>`: ім'я вихідного файла (за замовчуванням `<input>.avc`)
- `-target <target>`: `bytecode` (за замовчуванням) або `native` (поки не реалізовано)

### `avenir version`

Показує версію Avenir.

### `avenir help`

Показує довідку з використання.

## Структура програми

Програма Avenir складається з:

1. оголошення пакета `pckg <name>;`
2. імпортів (опційно)
3. структур (опційно)
4. інтерфейсів (опційно)
5. функцій і методів
6. функції `main` як точки входу

## Що читати далі

- [Syntax](syntax.md)
- [Types](types.md)
- [Control Flow](control-flow.md)
- [Functions](functions.md)
