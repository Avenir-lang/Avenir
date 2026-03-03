package collections

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.ListIndexOf,
			Name:       "indexOf",
			Arity:      2, // receiver + value
			ParamNames: []string{"self", "value"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, // receiver: list<any>
				{Kind: builtins.TypeAny}, // value: any
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeList,
			MethodName:   "indexOf",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("list.indexOf expects 2 arguments (receiver + value), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			searchVal := args[1].(value.Value)

			if receiver.Kind != value.KindList {
				return value.Value{}, fmt.Errorf("list.indexOf called on non-list type %v", receiver.Kind)
			}

			// Search for the value using deep equality
			for i, elem := range receiver.List {
				if equalValues(elem, searchVal) {
					return value.Int(int64(i)), nil
				}
			}

			// Not found - return -1
			return value.Int(-1), nil
		},
	})
}
