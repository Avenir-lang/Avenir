# Memory and Performance

This document covers runtime memory behavior and performance considerations.

## Value Semantics

Most values are immutable at the language level:

- `string` and `bytes` are immutable (operations return new values)
- `list` methods return new lists (no in‑place mutation)
- `dict.set` mutates the dictionary in place

Struct values are copied on assignment (value semantics), but they contain
references to underlying list/dict/bytes storage (shallow copy).

## Lists

Lists are stored as Go slices. List‑producing operations allocate new slices
and copy elements. Assigning a list value copies the slice header, so multiple
variables can reference the same underlying slice (though list APIs do not
mutate).

## Dicts

Dicts are backed by Go maps. `dict.set` and `dict.remove` mutate the map in
place, so sharing a dict value across variables shares underlying state.

## Strings

Strings are Go strings. Interpolation lowers to `OpStringify` and
`OpConcatString`. Concatenation uses a `strings.Builder` in the VM to reduce
allocations.

## Bytes

Bytes are `[]byte`. Methods like `append`, `concat`, and `slice` allocate new
byte slices and copy data.

## Structs

Struct fields are stored in a slice within `value.Value`. Field assignment
updates the struct value in place (for mutable fields).

## Error Values

Errors are regular values (`value.KindError`) and are propagated by the VM when
runtime failures occur.

## Optimization Notes

Current IR/VM is straightforward and prioritizes clarity over aggressive
optimization. Typical optimizations (constant folding, common subexpression
elimination) are not implemented.

## References

- `internal/value`
- `internal/vm/vm.go`
