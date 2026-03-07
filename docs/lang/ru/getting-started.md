# Быстрый старт

Это руководство поможет начать работу с языком программирования Avenir.

## Установка

Avenir реализован на Go. Чтобы собрать из исходников:

```bash
git clone <repository-url>
cd avenir
go build ./cmd/avenir
```

Собранный бинарник предоставляет CLI-команду `avenir`.

## Первая программа

Создайте файл `hello.av`:

```avenir
pckg main;

fun main() | void {
    print("Hello, Avenir!");
}
```

Каждая программа Avenir должна начинаться с `pckg` и содержать функцию `main()`.

## Запуск программы

```bash
avenir run hello.av
```

Команда:

1. парсит исходник
2. проверяет типы
3. компилирует в байткод
4. выполняет программу в VM

## Сборка байткода

```bash
avenir build hello.av -o hello.avc
```

Запуск `.avc` файла:

```bash
avenir run hello.avc
```

## CLI-команды

### `avenir run <file>`

Компилирует и запускает `.av`, либо запускает готовый `.avc`.

### `avenir build <file> [options]`

Компилирует `.av` в байткод.

Опции:

- `-o <file>`: имя выходного файла (по умолчанию `<input>.avc`)
- `-target <target>`: `bytecode` (по умолчанию) или `native` (пока не реализован)

### `avenir version`

Показывает версию Avenir.

### `avenir help`

Показывает справку по использованию.

## Структура программы

Программа Avenir состоит из:

1. объявления пакета `pckg <name>;`
2. импортов (опционально)
3. структур (опционально)
4. интерфейсов (опционально)
5. функций и методов
6. функции `main` как точки входа

## Что читать дальше

- [Syntax](syntax.md)
- [Types](types.md)
- [Control Flow](control-flow.md)
- [Functions](functions.md)
