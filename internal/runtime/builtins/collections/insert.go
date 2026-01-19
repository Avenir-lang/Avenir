package collections

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.ListInsert,
			Name:       "insert",
			Arity:      3, // receiver + index + value
			ParamNames: []string{"self", "index", "value"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, // receiver: list<any>
				{Kind: builtins.TypeInt}, // index: int
				{Kind: builtins.TypeAny}, // value: any
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeList,
			MethodName:   "insert",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 3 {
				return value.Value{}, fmt.Errorf("list.insert expects 3 arguments (receiver + index + value), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			indexVal := args[1].(value.Value)
			element := args[2].(value.Value)

			if receiver.Kind != value.KindList {
				return value.Value{}, fmt.Errorf("list.insert called on non-list type %v", receiver.Kind)
			}
			if indexVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("list.insert: index argument must be int, got %v", indexVal.Kind)
			}

			index := int(indexVal.Int)
			listLen := len(receiver.List)

			// Validate index bounds (allow insertion at end: index == listLen)
			if index < 0 || index > listLen {
				return value.Value{}, fmt.Errorf("list.insert: index %d out of bounds [0, %d]", index, listLen)
			}

			// Create new list with element inserted
			newList := make([]value.Value, listLen+1)
			copy(newList[:index], receiver.List[:index])
			newList[index] = element
			copy(newList[index+1:], receiver.List[index:])

			return value.List(newList), nil
		},
	})
}
