package collections

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.ListRemoveAt,
			Name:       "removeAt",
			Arity:      2, // receiver + index
			ParamNames: []string{"self", "index"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, // receiver: list<any>
				{Kind: builtins.TypeInt}, // index: int
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeList,
			MethodName:   "removeAt",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("list.removeAt expects 2 arguments (receiver + index), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			indexVal := args[1].(value.Value)

			if receiver.Kind != value.KindList {
				return value.Value{}, fmt.Errorf("list.removeAt called on non-list type %v", receiver.Kind)
			}
			if indexVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("list.removeAt: index argument must be int, got %v", indexVal.Kind)
			}

			index := int(indexVal.Int)
			listLen := len(receiver.List)

			// Validate index bounds
			if index < 0 || index >= listLen {
				return value.Value{}, fmt.Errorf("list.removeAt: index %d out of bounds [0, %d)", index, listLen)
			}

			// Create new list without the element at index
			newList := make([]value.Value, listLen-1)
			copy(newList[:index], receiver.List[:index])
			copy(newList[index:], receiver.List[index+1:])

			return value.List(newList), nil
		},
	})
}
