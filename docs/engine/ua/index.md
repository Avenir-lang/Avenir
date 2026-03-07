# Документація рушія (Українська)

Цей розділ описує внутрішню будову Avenir: лексер, парсер, AST,
перевірку типів, IR, VM, runtime-сервіси та модульну систему.

## Огляд пайплайна

1. **Лексинг** → `internal/lexer`
2. **Парсинг** → `internal/parser`
3. **AST** → `internal/ast`
4. **Перевірка типів** → `internal/types`
5. **Генерація IR** → `internal/ir`
6. **Виконання у VM** → `internal/vm`
7. **Runtime та builtins** → `internal/runtime`
8. **Модулі та імпорти** → `internal/modules`

## Документи

- [Лексер](lexer.md)
- [Парсер](parser.md)
- [AST](ast.md)
- [Типи та type checker](types.md)
- [IR (проміжне представлення)](ir.md)
- [VM](vm.md)
- [Runtime та builtins](runtime.md)
- [Модулі та імпорти](modules.md)
- [Помилки](errors.md)
- [Пам'ять і продуктивність](memory.md)
- [Тестування](testing.md)
