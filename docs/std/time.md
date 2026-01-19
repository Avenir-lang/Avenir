# std.time

`std.time` provides date/time and duration utilities with a thin runtime layer.
Values are immutable and all errors are catchable via `try / catch`.

## Overview

- **DateTime** stores a Unix timestamp in **nanoseconds** (`int`)
- **Duration** stores nanoseconds (`int`)
- All operations use **UTC** for parsing/formatting and component access

## Public Structs

```avenir
pub struct Duration {
    nanos | int
}

pub struct DateTime {
    timestamp | int
}
```

## Functions

| Function | Parameters | Returns | Errors |
| --- | --- | --- | --- |
| `now` | — | `DateTime` | — |
| `sleep` | `d | Duration` | `void` | negative duration |
| `parseDateTime` | `text | string`, `format | string` | `DateTime` | parse errors |
| `parseDuration` | `text | string` | `Duration` | parse errors |
| `formatDateTime` | `dt | DateTime`, `format | string` | `string` | format errors |
| `formatISO8601` | `dt | DateTime` | `string` | — |
| `parseISO8601` | `text | string` | `DateTime` | parse errors |

### Duration Constructors

| Function | Parameters | Returns |
| --- | --- | --- |
| `fromNanos` | `n | int` | `Duration` |
| `fromMillis` | `ms | int` | `Duration` |
| `fromSeconds` | `sec | int` | `Duration` |
| `fromMinutes` | `min | int` | `Duration` |
| `fromHours` | `hours | int` | `Duration` |

## DateTime Methods

| Method | Parameters | Returns | Notes |
| --- | --- | --- | --- |
| `year` | — | `int` | UTC |
| `month` | — | `int` | 1‑12 |
| `day` | — | `int` | 1‑31 |
| `hour` | — | `int` | 0‑23 |
| `minute` | — | `int` | 0‑59 |
| `second` | — | `int` | 0‑59 |
| `add` | `d | Duration` | `DateTime` | Adds nanoseconds |
| `sub` | `d | Duration` | `DateTime` | Subtracts nanoseconds |
| `format` | `fmt | string` | `string` | UTC formatting |

## Duration Methods

| Method | Parameters | Returns |
| --- | --- | --- |
| `hours` | — | `float` |
| `minutes` | — | `float` |
| `seconds` | — | `float` |
| `milliseconds` | — | `float` |
| `add` | `other | Duration` | `Duration` |
| `sub` | `other | Duration` | `Duration` |

## Format Tokens

`formatDateTime` and `parseDateTime` accept a simple token format. Tokens are
mapped to Go layouts internally.

| Token | Meaning |
| --- | --- |
| `YYYY` | 4‑digit year |
| `YY` | 2‑digit year |
| `MM` | month (01‑12) |
| `DD` | day (01‑31) |
| `HH` | hour (00‑23) |
| `mm` | minute (00‑59) |
| `ss` | second (00‑59) |
| `SSS` | milliseconds |
| `Z` | UTC offset (`Z` or `+/-HH:MM`) |

Example format: `YYYY-MM-DD HH:mm:ss`

## Duration Parsing

`parseDuration` uses Go‑style duration strings:

- `300ms`, `1.5s`, `2m`, `1h30m`

## Examples

### Current time and formatting

```avenir
import std.time;

fun main() | void {
    var now | time.DateTime = time.now();
    print(now.format("YYYY-MM-DD HH:mm:ss"));
}
```

### Parsing and arithmetic

```avenir
import std.time;

fun main() | void {
    var dt | time.DateTime = time.parseDateTime("2024-02-03 04:05:06", "YYYY-MM-DD HH:mm:ss");
    var d | time.Duration = time.parseDuration("2h30m");
    var later | time.DateTime = dt.add(d);
    print(later.format("YYYY-MM-DD HH:mm:ss"));
}
```

### Sleep

```avenir
import std.time;

fun main() | void {
    time.sleep(time.parseDuration("50ms"));
    print("awake");
}
```

## Error Handling

All parse/format errors and invalid inputs throw runtime errors and can be
caught with `try / catch`.
