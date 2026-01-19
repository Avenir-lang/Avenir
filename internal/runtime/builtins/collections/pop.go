package collections

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.ListPop,
			Name:       "pop",
			Arity:      1, // receiver only
			ParamNames: []string{"self"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, // receiver: list<any>
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeList,
			MethodName:   "pop",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("list.pop expects 1 argument (receiver), got %d", len(args))
			}
			receiver := args[0].(value.Value)

			if receiver.Kind != value.KindList {
				return value.Value{}, fmt.Errorf("list.pop called on non-list type %v", receiver.Kind)
			}

			if len(receiver.List) == 0 {
				return value.Value{}, fmt.Errorf("list.pop: cannot pop from empty list")
			}

			// Return the last element and create a new list without it
			popped := receiver.List[len(receiver.List)-1]
			return popped, nil
		},
	})
}
