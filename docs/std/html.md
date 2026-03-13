# std.web.html — HTML Templating System

The `std.web.html` module provides two complementary approaches to generating HTML:

1. **HTML Builder DSL** — type-safe, programmatic HTML construction using nested closures
2. **Template Engine** — file-based templates with Jinja-like syntax (inheritance, loops, conditions, filters)

Both approaches auto-escape output by default to prevent XSS.

## Quick Start

### Builder DSL

```avenir
import std.web.html as html;

var page | string = html.build(fun(h | html.Builder) | void {
    h.doctype();
    h.html(fun() | void {
        h.head(fun() | void {
            h.title("My Page");
        });
        h.body(fun() | void {
            h.h1("Hello World");
            h.p("Safe by default");
        });
    });
});
```

### Template Engine

```avenir
import std.web.html as html;

var engine | html.TemplateEngine = html.newEngine("templates/");
var result | string = engine.render("page.html", {
    "title": "My Page",
    "name": "Alice"
});
```

## Builder API

### `html.build(fn | fun(Builder) | void) | string`

Creates a builder, passes it to `fn`, and returns the generated HTML string.

### Builder Methods

Every HTML tag method accepts an optional attributes dict and/or content:

```avenir
h.div("text content");                          // <div>text content</div>
h.div({class: "box"}, "text");                  // <div class="box">text</div>
h.div({class: "box"}, fun() | void { ... });    // <div class="box">...</div>
h.div(fun() | void { ... });                    // <div>...</div>
```

**Container tags**: `html`, `head`, `body`, `div`, `span`, `p`, `a`, `h1`–`h6`, `ul`, `ol`, `li`, `dl`, `dt`, `dd`, `table`, `thead`, `tbody`, `tfoot`, `tr`, `th`, `td`, `caption`, `form`, `textarea`, `button`, `select`, `option`, `optgroup`, `label`, `fieldset`, `legend`, `title`, `style`, `script`, `noscript`, `main`, `header`, `footer`, `nav`, `aside`, `section`, `article`, `strong`, `em`, `small`, `code`, `pre`, `blockquote`, `abbr`, `cite`, `figure`, `figcaption`, `details`, `summary`, `dialog`, `time`, `audio`, `video`, `picture`, `colgroup`

**Void tags**: `br()`, `hr()`, `wbr()`, `img(attrs)`, `input(attrs)`, `meta(attrs)`, `link(attrs)`, `source(attrs)`, `col(attrs)`

**Special**:
- `h.text(s)` — escaped text node
- `h.rawHtml(s)` — unescaped raw HTML (use with caution)
- `h.doctype()` — emits `<!DOCTYPE html>`
- `h.tag(name, ...)` — generic tag by name
- `h.voidTag(name, attrs)` — generic void tag

### Escaping

```avenir
var safe | string = html.escape("<script>alert('xss')</script>");
// &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;
```

## Template Engine API

### `html.newEngine(dir | string, opts | dict<any> = {}) | TemplateEngine`

Creates a template engine that loads `.html` files from `dir`.

Options:
- `devMode` (bool) — when `true`, reloads templates on each request if the file changed

### `engine.render(name | string, data | dict<any> = {}) | string`

Renders a template by name with the given data context.

### `engine.compile(name | string) | Template`

Pre-compiles a template for repeated rendering.

### `template.render(data | dict<any> = {}) | string`

Renders a pre-compiled template.

### `engine.setDevMode(enabled | bool) | void`

Toggles dev mode at runtime.

## Template Syntax

### Expressions

```html
<h1>{{ title }}</h1>
<p>{{ user.name }}</p>
```

All expressions are HTML-escaped by default.

### Filters

```html
{{ name | upper }}
{{ bio | lower }}
{{ value | trim }}
{{ name | default("Anonymous") }}
{{ rawContent | raw }}
```

### Conditions

```html
{% if show %}
    <p>Visible</p>
{% elif altShow %}
    <p>Alt visible</p>
{% else %}
    <p>Hidden</p>
{% endif %}
```

### Loops

```html
{% for item in items %}
    <li>{{ item }}</li>
{% endfor %}

{% for key, value in dict %}
    <dt>{{ key }}</dt>
    <dd>{{ value }}</dd>
{% endfor %}
```

### Template Inheritance

**base.html**:
```html
<!DOCTYPE html>
<html>
<head><title>{% block title %}Default{% endblock %}</title></head>
<body>{% block content %}{% endblock %}</body>
</html>
```

**page.html**:
```html
{% extends "base.html" %}
{% block title %}{{ title }}{% endblock %}
{% block content %}<h1>{{ name }}</h1>{% endblock %}
```

### Includes

```html
{% include "header.html" %}
{% include "sidebar.html" with active="home" %}
```

### Comments

```html
{# This is a comment and won't appear in output #}
```

## CoolWeb Integration

### Setup

```avenir
import std.coolweb;

var app | coolweb.App = coolweb.newApp();
app.setTemplates("templates/");
```

### Rendering in Handlers

```avenir
@app.get("/")
fun index(ctx | coolweb.Context) | coolweb.Response {
    return ctx.render("index.html", {
        "title": "Home",
        "user": ctx.params["name"]
    });
}
```

`ctx.render(name, data, status)` is a convenience method that:
1. Looks up the template engine from the app configuration
2. Renders the named template with the data
3. Returns an HTML response with the given status (default 200)

## Safety

- All `{{ expression }}` output is HTML-escaped by default
- Use `| raw` filter only for trusted content
- Builder `h.text()` always escapes; use `h.rawHtml()` for trusted raw HTML
- `html.escape(s)` is available for manual escaping

## Module Structure

```
std/web/html/
├── html.av      — main entry (build, escape, raw, newEngine)
├── builder.av   — Builder struct + all tag methods
├── engine.av    — TemplateEngine, Template structs + methods
├── safe.av      — SafeString struct
└── errors.av    — TemplateError struct
```
