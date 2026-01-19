package collections

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.ListAppend,
			Name:       "append",
			Arity:      2, // receiver + element
			ParamNames: []string{"self", "element"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, // receiver: list<any>
				{Kind: builtins.TypeAny}, // element to append
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeList,
			MethodName:   "append",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("list.append expects 2 arguments (receiver + element), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			element := args[1].(value.Value)

			if receiver.Kind != value.KindList {
				return value.Value{}, fmt.Errorf("list.append called on non-list type %v", receiver.Kind)
			}

			// Append element to the list
			newList := make([]value.Value, len(receiver.List)+1)
			copy(newList, receiver.List)
			newList[len(receiver.List)] = element

			return value.List(newList), nil
		},
	})
}
