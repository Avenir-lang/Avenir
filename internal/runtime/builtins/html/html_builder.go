package html

import (
	"fmt"
	"sort"
	"strings"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

// builderHandle is the Go-side state for an HTML builder.
type builderHandle struct {
	buf strings.Builder
}

// safeStringMarker is a sentinel dict key that marks a SafeString value.
const safeStringMarker = "__avenir_safe_html__"

func init() {
	registerNewBuilder()
	registerTag()
	registerVoidTag()
	registerText()
	registerRawHTML()
	registerDoctype()
	registerBuilderResult()
	registerEscape()
	registerRaw()
}

func registerNewBuilder() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.HTMLNewBuilder,
			Name:         "__builtin_html_new_builder",
			Arity:        0,
			ParamNames:   []string{},
			Params:       []builtins.TypeRef{},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			h := &builderHandle{}
			return value.Value{Kind: value.KindBytes, Bytes: encodeBuilderPtr(h)}, nil
		},
	})
}

func registerTag() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTMLTag,
			Name:       "__builtin_html_tag",
			Arity:      4,
			ParamNames: []string{"handle", "tag", "first", "second"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 4 {
				return value.Value{}, fmt.Errorf("html.tag: expected 4 args, got %d", len(args))
			}
			handleVal := args[0].(value.Value)
			tagVal := args[1].(value.Value)
			first := args[2].(value.Value)
			second := args[3].(value.Value)

			h := decodeBuilderPtr(handleVal.Bytes)
			if h == nil {
				return value.Value{}, fmt.Errorf("html.tag: invalid builder handle")
			}
			tag := tagVal.Str

			attrs, content := classifyArgs(first, second)

			h.buf.WriteByte('<')
			h.buf.WriteString(tag)
			if attrs != nil {
				writeAttrs(&h.buf, attrs)
			}
			h.buf.WriteByte('>')

			if err := writeContent(env, h, content); err != nil {
				return value.Value{}, err
			}

			h.buf.WriteString("</")
			h.buf.WriteString(tag)
			h.buf.WriteByte('>')

			return value.Value{Kind: value.KindOptional, Optional: &value.OptionalValue{IsSome: false}}, nil
		},
	})
}

func registerVoidTag() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTMLVoidTag,
			Name:       "__builtin_html_void_tag",
			Arity:      3,
			ParamNames: []string{"handle", "tag", "attrs"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 3 {
				return value.Value{}, fmt.Errorf("html.void_tag: expected 3 args, got %d", len(args))
			}
			handleVal := args[0].(value.Value)
			tagVal := args[1].(value.Value)
			attrsVal := args[2].(value.Value)

			h := decodeBuilderPtr(handleVal.Bytes)
			if h == nil {
				return value.Value{}, fmt.Errorf("html.void_tag: invalid builder handle")
			}

			h.buf.WriteByte('<')
			h.buf.WriteString(tagVal.Str)
			if attrsVal.Kind == value.KindDict && len(attrsVal.Dict) > 0 {
				writeAttrs(&h.buf, attrsVal.Dict)
			}
			h.buf.WriteByte('>')

			return value.Value{Kind: value.KindOptional, Optional: &value.OptionalValue{IsSome: false}}, nil
		},
	})
}

func registerText() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTMLText,
			Name:       "__builtin_html_text",
			Arity:      2,
			ParamNames: []string{"handle", "content"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			handleVal := args[0].(value.Value)
			contentVal := args[1].(value.Value)

			h := decodeBuilderPtr(handleVal.Bytes)
			if h == nil {
				return value.Value{}, fmt.Errorf("html.text: invalid builder handle")
			}

			h.buf.WriteString(escapeHTML(contentVal.Str))
			return value.Value{Kind: value.KindOptional, Optional: &value.OptionalValue{IsSome: false}}, nil
		},
	})
}

func registerRawHTML() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTMLRawHTML,
			Name:       "__builtin_html_raw_html",
			Arity:      2,
			ParamNames: []string{"handle", "content"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			handleVal := args[0].(value.Value)
			contentVal := args[1].(value.Value)

			h := decodeBuilderPtr(handleVal.Bytes)
			if h == nil {
				return value.Value{}, fmt.Errorf("html.raw_html: invalid builder handle")
			}

			h.buf.WriteString(contentVal.Str)
			return value.Value{Kind: value.KindOptional, Optional: &value.OptionalValue{IsSome: false}}, nil
		},
	})
}

func registerDoctype() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTMLDoctype,
			Name:       "__builtin_html_doctype",
			Arity:      1,
			ParamNames: []string{"handle"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			handleVal := args[0].(value.Value)
			h := decodeBuilderPtr(handleVal.Bytes)
			if h == nil {
				return value.Value{}, fmt.Errorf("html.doctype: invalid builder handle")
			}
			h.buf.WriteString("<!DOCTYPE html>")
			return value.Value{Kind: value.KindOptional, Optional: &value.OptionalValue{IsSome: false}}, nil
		},
	})
}

func registerBuilderResult() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTMLBuilderResult,
			Name:       "__builtin_html_builder_result",
			Arity:      1,
			ParamNames: []string{"handle"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			handleVal := args[0].(value.Value)
			h := decodeBuilderPtr(handleVal.Bytes)
			if h == nil {
				return value.Value{}, fmt.Errorf("html.builder_result: invalid builder handle")
			}
			return value.Str(h.buf.String()), nil
		},
	})
}

func registerEscape() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTMLEscape,
			Name:       "__builtin_html_escape",
			Arity:      1,
			ParamNames: []string{"s"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			s := args[0].(value.Value)
			return value.Str(escapeHTML(s.Str)), nil
		},
	})
}

func registerRaw() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTMLRaw,
			Name:       "__builtin_html_raw",
			Arity:      1,
			ParamNames: []string{"s"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeDict},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			s := args[0].(value.Value)
			dict := map[string]value.Value{
				safeStringMarker: value.Bool(true),
				"value":          value.Str(s.Str),
			}
			return value.Dict(dict), nil
		},
	})
}

// --- helpers ---

// We store the builder pointer in a global map keyed by a unique ID encoded in bytes.
// This avoids unsafe pointer tricks.

var (
	builderMu      = make(map[uint64]*builderHandle)
	builderCounter uint64
)

func encodeBuilderPtr(h *builderHandle) []byte {
	builderCounter++
	id := builderCounter
	builderMu[id] = h
	b := make([]byte, 8)
	b[0] = byte(id)
	b[1] = byte(id >> 8)
	b[2] = byte(id >> 16)
	b[3] = byte(id >> 24)
	b[4] = byte(id >> 32)
	b[5] = byte(id >> 40)
	b[6] = byte(id >> 48)
	b[7] = byte(id >> 56)
	return b
}

func decodeBuilderPtr(b []byte) *builderHandle {
	if len(b) != 8 {
		return nil
	}
	id := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
	return builderMu[id]
}

func classifyArgs(first, second value.Value) (map[string]value.Value, value.Value) {
	isNone := func(v value.Value) bool {
		return v.Kind == value.KindOptional && (v.Optional == nil || !v.Optional.IsSome)
	}

	if isNone(first) && isNone(second) {
		return nil, value.Value{}
	}

	if first.Kind == value.KindDict && !isSafeString(first) {
		if isNone(second) {
			return first.Dict, value.Value{}
		}
		return first.Dict, second
	}

	return nil, first
}

func isSafeString(v value.Value) bool {
	if v.Kind != value.KindDict {
		return false
	}
	marker, ok := v.Dict[safeStringMarker]
	return ok && marker.Kind == value.KindBool && marker.Bool
}

func writeAttrs(b *strings.Builder, attrs map[string]value.Value) {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := attrs[k]
		b.WriteByte(' ')
		b.WriteString(k)
		b.WriteString(`="`)
		if v.Kind == value.KindString {
			b.WriteString(escapeAttrValue(v.Str))
		} else {
			b.WriteString(escapeAttrValue(v.String()))
		}
		b.WriteByte('"')
	}
}

func writeContent(env builtins.Env, h *builderHandle, content value.Value) error {
	switch content.Kind {
	case value.KindString:
		h.buf.WriteString(escapeHTML(content.Str))
	case value.KindClosure:
		_, err := env.CallClosure(content.Closure, []interface{}{})
		if err != nil {
			return fmt.Errorf("html.tag: error calling content closure: %w", err)
		}
	case value.KindDict:
		if isSafeString(content) {
			if raw, ok := content.Dict["value"]; ok && raw.Kind == value.KindString {
				h.buf.WriteString(raw.Str)
			}
		}
	case value.KindOptional:
		// none — no content
	case value.KindInvalid:
		// no content (zero value)
	default:
		h.buf.WriteString(escapeHTML(content.String()))
	}
	return nil
}
