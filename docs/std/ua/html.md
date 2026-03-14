# std.web.html — HTML-шаблонізація

Модуль `std.web.html` надає два підходи до генерації HTML: HTML Builder DSL та Template Engine.

## Імпорт

```avenir
import std.web.html;
```

## Швидкий старт

### HTML Builder

```avenir
import std.web.html;

fun main() | void {
    var page | string = std.web.html.build(fun(h | std.web.html.Builder) | void {
        h.doctype();
        h.html({lang: "uk"}, fun(h | std.web.html.Builder) | void {
            h.head({}, fun(h | std.web.html.Builder) | void {
                h.title({}, "Моя сторінка");
            });
            h.body({}, fun(h | std.web.html.Builder) | void {
                h.h1({}, "Привіт!");
                h.p({class: "intro"}, "Ласкаво просимо.");
            });
        });
    });
    print(page);
}
```

### Template Engine

```avenir
import std.web.html;

fun main() | void {
    var engine | std.web.html.TemplateEngine = std.web.html.newEngine("templates/");
    var data | dict<any> = {title: "Головна", name: "Аліса"};
    var result | string = engine.render("index.html", data);
    print(result);
}
```

## Builder API

### Функції

| Функція | Опис |
|---------|------|
| `build(fn \| fun(Builder) \| void) \| string` | Побудувати HTML-рядок |
| `escape(s \| string) \| string` | Екранувати HTML-сутності |
| `raw(s \| string) \| SafeString` | Позначити рядок як безпечний (без екранування) |

### Методи Builder

Builder має методи для всіх стандартних HTML-тегів:

**Контейнерні теги** (приймають атрибути та вміст/замикання):
`html`, `head`, `body`, `div`, `span`, `p`, `h1`–`h6`, `a`, `ul`, `ol`, `li`, `table`, `tr`, `td`, `th`, `thead`, `tbody`, `form`, `button`, `select`, `option`, `textarea`, `label`, `nav`, `header`, `footer`, `main`, `section`, `article`, `aside`, `script`, `style`, `title`, `pre`, `code`, `blockquote`, `strong`, `em`, `small`, `iframe`

**Порожні (void) теги** (приймають лише атрибути):
`br`, `hr`, `img`, `input`, `meta`, `link`

**Спеціальні методи**:
| Метод | Опис |
|-------|------|
| `doctype() \| void` | Генерація `<!DOCTYPE html>` |
| `text(s \| string) \| void` | Текстовий вузол (з екрануванням) |
| `rawHTML(s \| string) \| void` | Сирий HTML (без екранування) |

### Екранування

Builder автоматично екранує текстовий вміст та значення атрибутів. Використовуйте `raw()` або `rawHTML()` для вставки неекранованого HTML.

## Template Engine API

### Функції

| Функція | Опис |
|---------|------|
| `newEngine(dir \| string) \| TemplateEngine` | Створити движок шаблонів |

### Методи TemplateEngine

| Метод | Опис |
|-------|------|
| `render(name \| string, data \| dict<any>) \| string` | Рендеринг шаблону |
| `compile(name \| string) \| Template` | Попередня компіляція шаблону |
| `setDevMode(enabled \| bool) \| void` | Увімкнути dev-режим (автоперезавантаження) |

### Методи Template

| Метод | Опис |
|-------|------|
| `render(data \| dict<any>) \| string` | Рендеринг скомпільованого шаблону |

## Синтаксис шаблонів

### Вирази

```html
{{ variable }}
{{ user.name }}
{{ title | upper }}
```

### Фільтри

| Фільтр | Опис |
|--------|------|
| `upper` | Верхній регістр |
| `lower` | Нижній регістр |
| `trim` | Видалити пробіли |
| `default(val)` | Значення за замовчуванням |
| `raw` | Без екранування |

### Умови

```html
{% if user %}
    <p>Привіт, {{ user.name }}!</p>
{% elif guest %}
    <p>Привіт, гостю!</p>
{% else %}
    <p>Привіт!</p>
{% endif %}
```

### Цикли

```html
{% for item in items %}
    <li>{{ item }}</li>
{% endfor %}

{% for key, value in dict %}
    <dt>{{ key }}</dt>
    <dd>{{ value }}</dd>
{% endfor %}
```

### Наслідування шаблонів

**base.html:**
```html
<html>
<head><title>{% block title %}Default{% endblock %}</title></head>
<body>{% block content %}{% endblock %}</body>
</html>
```

**page.html:**
```html
{% extends "base.html" %}
{% block title %}Моя сторінка{% endblock %}
{% block content %}<p>Вміст сторінки</p>{% endblock %}
```

### Включення

```html
{% include "header.html" %}
```

### Коментарі

```html
{# Це коментар — не буде в результаті #}
```

## Інтеграція з CoolWeb

```avenir
import std.coolweb;
import std.web.html;

var app | std.coolweb.App = std.coolweb.newApp();
app.setTemplates("templates/");

@app.get("/")
async fun home(ctx | std.coolweb.Context) | void {
    var data | dict<any> = {title: "Головна", user: "Аліса"};
    ctx.render("index.html", data, 200);
}
```

## Безпека

- Вміст екранується за замовчуванням (Builder та Template Engine)
- Використовуйте `raw` фільтр або `rawHTML()` лише для довіреного вмісту
- Значення атрибутів екрануються автоматично

## Структура модуля

```
std/web/html/
├── html.av      # Головний API: build(), escape(), raw(), newEngine()
├── builder.av   # Builder з методами тегів
├── engine.av    # TemplateEngine та Template
├── safe.av      # SafeString
└── errors.av    # TemplateError
```
