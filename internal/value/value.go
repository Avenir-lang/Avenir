package value

import (
	"fmt"
	"strings"

	"avenir/internal/ir"
)

// Kind is the type of a value at runtime.
type Kind int

const (
	KindInvalid Kind = iota
	KindInt
	KindFloat
	KindString
	KindBool
	KindList
	KindClosure
	KindError
	KindBytes
	KindOptional
	KindStruct
	KindDict
)

// Upvalue represents a captured variable.
// If open, points to a slot in the VM stack.
// If closed, holds a copied Value.
type Upvalue struct {
	IsClosed bool
	Index    int   // stack index when open
	Closed   Value // captured value when closed
}

// Closure represents a function closure with captured variables.
type Closure struct {
	Fn       *ir.Function
	Upvalues []*Upvalue
}

// OptionalValue represents an optional value (some or none)
type OptionalValue struct {
	IsSome bool
	Value  Value // only valid if IsSome is true
}

// StructValue represents a struct instance with its fields.
type StructValue struct {
	TypeIndex int     // index into struct type registry
	Fields    []Value // field values in declaration order
}

// ErrorInfo represents a runtime error with extensible metadata.
type ErrorInfo struct {
	Message string
	Meta    map[string]string
}

// Value is a universal value for the VM/runtime.
type Value struct {
	Kind     Kind
	Int      int64
	Float    float64
	Str      string
	Bool     bool
	List     []Value
	Dict     map[string]Value
	Closure  *Closure       // for KindClosure
	Bytes    []byte         // for KindBytes
	Optional *OptionalValue // for KindOptional
	Struct   *StructValue   // for KindStruct
	Error    *ErrorInfo
}

func (v Value) String() string {
	switch v.Kind {
	case KindInt:
		return fmt.Sprintf("%d", v.Int)
	case KindFloat:
		return fmt.Sprintf("%g", v.Float)
	case KindString:
		return v.Str
	case KindBool:
		if v.Bool {
			return "true"
		}
		return "false"
	case KindList:
		var b strings.Builder
		b.WriteByte('[')
		for i, el := range v.List {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(el.String())
		}
		b.WriteByte(']')
		return b.String()
	case KindClosure:
		if v.Closure != nil {
			return fmt.Sprintf("<closure %s>", v.Closure.Fn.Name)
		}
		return "<closure nil>"
	case KindError:
		msg := v.Str
		if v.Error != nil && v.Error.Message != "" {
			msg = v.Error.Message
		}
		return "error(" + msg + ")"
	case KindBytes:
		return fmt.Sprintf("bytes(%d)", len(v.Bytes))
	case KindOptional:
		if v.Optional != nil && v.Optional.IsSome {
			return fmt.Sprintf("some(%s)", v.Optional.Value.String())
		}
		return "none"
	case KindStruct:
		if v.Struct == nil {
			return "<nil struct>"
		}
		var b strings.Builder
		b.WriteString("{")
		for i, f := range v.Struct.Fields {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(f.String())
		}
		b.WriteString("}")
		return b.String()
	case KindDict:
		var b strings.Builder
		b.WriteString("{")
		first := true
		for k, val := range v.Dict {
			if !first {
				b.WriteString(", ")
			}
			first = false
			b.WriteString(k)
			b.WriteString(": ")
			b.WriteString(val.String())
		}
		b.WriteString("}")
		return b.String()
	default:
		return "<invalid>"
	}
}

// Helpers

func Int(v int64) Value {
	return Value{Kind: KindInt, Int: v}
}

func Float(v float64) Value {
	return Value{Kind: KindFloat, Float: v}
}

func Str(s string) Value {
	return Value{Kind: KindString, Str: s}
}

func Bool(v bool) Value {
	return Value{Kind: KindBool, Bool: v}
}

func Bytes(b []byte) Value {
	return Value{Kind: KindBytes, Bytes: b}
}

func List(vals []Value) Value {
	return Value{Kind: KindList, List: vals}
}

func NewClosure(fn *ir.Function, ups []*Upvalue) Value {
	return Value{Kind: KindClosure, Closure: &Closure{Fn: fn, Upvalues: ups}}
}

func ErrorValue(msg string) Value {
	return Value{
		Kind:  KindError,
		Str:   msg,
		Error: &ErrorInfo{Message: msg},
	}
}

func Some(v Value) Value {
	return Value{
		Kind: KindOptional,
		Optional: &OptionalValue{
			IsSome: true,
			Value:  v,
		},
	}
}

func None() Value {
	return Value{
		Kind: KindOptional,
		Optional: &OptionalValue{
			IsSome: false,
		},
	}
}

// Struct creates a struct value with the given type index and fields.
func Struct(typeIndex int, fields []Value) Value {
	return Value{
		Kind: KindStruct,
		Struct: &StructValue{
			TypeIndex: typeIndex,
			Fields:    fields,
		},
	}
}

// Dict creates a dict value with the given key/value map.
func Dict(entries map[string]Value) Value {
	return Value{
		Kind: KindDict,
		Dict: entries,
	}
}
