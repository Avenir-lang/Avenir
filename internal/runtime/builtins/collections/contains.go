package collections

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

// equalValues performs deep equality comparison of two values.
// This matches the semantics used by the VM's OpEq operation.
func equalValues(a, b value.Value) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case value.KindInt:
		return a.Int == b.Int
	case value.KindFloat:
		return a.Float == b.Float
	case value.KindString:
		return a.Str == b.Str
	case value.KindBool:
		return a.Bool == b.Bool
	case value.KindError:
		return a.Str == b.Str
	case value.KindBytes:
		if len(a.Bytes) != len(b.Bytes) {
			return false
		}
		for i := range a.Bytes {
			if a.Bytes[i] != b.Bytes[i] {
				return false
			}
		}
		return true
	case value.KindList:
		if len(a.List) != len(b.List) {
			return false
		}
		for i := range a.List {
			if !equalValues(a.List[i], b.List[i]) {
				return false
			}
		}
		return true
	case value.KindOptional:
		if a.Optional == nil || b.Optional == nil {
			return a.Optional == b.Optional
		}
		if a.Optional.IsSome != b.Optional.IsSome {
			return false
		}
		if a.Optional.IsSome {
			return equalValues(a.Optional.Value, b.Optional.Value)
		}
		return true
	case value.KindStruct:
		if a.Struct == nil || b.Struct == nil {
			return a.Struct == b.Struct
		}
		if a.Struct.TypeIndex != b.Struct.TypeIndex {
			return false
		}
		if len(a.Struct.Fields) != len(b.Struct.Fields) {
			return false
		}
		for i := range a.Struct.Fields {
			if !equalValues(a.Struct.Fields[i], b.Struct.Fields[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.ListContains,
			Name:       "contains",
			Arity:      2, // receiver + value
			ParamNames: []string{"self", "value"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, // receiver: list<any>
				{Kind: builtins.TypeAny}, // value: any
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBool},
			ReceiverType: builtins.TypeList,
			MethodName:   "contains",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("list.contains expects 2 arguments (receiver + value), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			searchVal := args[1].(value.Value)

			if receiver.Kind != value.KindList {
				return value.Value{}, fmt.Errorf("list.contains called on non-list type %v", receiver.Kind)
			}

			// Search for the value using deep equality
			for _, elem := range receiver.List {
				if equalValues(elem, searchVal) {
					return value.Bool(true), nil
				}
			}

			return value.Bool(false), nil
		},
	})
}
