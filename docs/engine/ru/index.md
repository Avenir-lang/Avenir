# Документация движка (Русский)

Этот раздел описывает внутреннее устройство Avenir: лексер, парсер, AST,
проверку типов, IR, VM, runtime-сервисы и модульную систему.

## Обзор пайплайна

1. **Лексинг** → `internal/lexer`
2. **Парсинг** → `internal/parser`
3. **AST** → `internal/ast`
4. **Проверка типов** → `internal/types`
5. **Генерация IR** → `internal/ir`
6. **Исполнение в VM** → `internal/vm`
7. **Runtime и builtins** → `internal/runtime`
8. **Модули и импорты** → `internal/modules`

## Документы

- [Лексер](lexer.md)
- [Парсер](parser.md)
- [AST](ast.md)
- [Типы и type checker](types.md)
- [IR (промежуточное представление)](ir.md)
- [VM](vm.md)
- [Runtime и builtins](runtime.md)
- [Модули и импорты](modules.md)
- [Ошибки](errors.md)
- [Память и производительность](memory.md)
- [Тестирование](testing.md)
