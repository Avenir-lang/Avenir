package collections

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.ListMap,
			Name:       "map",
			Arity:      2, // receiver + function
			ParamNames: []string{"self", "fn"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, // receiver: list<any>
				{Kind: builtins.TypeAny}, // fn: function (type checked at compile time)
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeList,
			MethodName:   "map",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("list.map expects 2 arguments (receiver + fn), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			fnVal := args[1].(value.Value)

			if receiver.Kind != value.KindList {
				return value.Value{}, fmt.Errorf("list.map called on non-list type %v", receiver.Kind)
			}
			if fnVal.Kind != value.KindClosure {
				return value.Value{}, fmt.Errorf("list.map: fn argument must be a function, got %v", fnVal.Kind)
			}

			// Apply the function to each element
			resultList := make([]value.Value, len(receiver.List))
			for i, elem := range receiver.List {
				// Call the function with the element as argument
				callArgs := []interface{}{elem}
				result, err := env.CallClosure(fnVal.Closure, callArgs)
				if err != nil {
					return value.Value{}, fmt.Errorf("list.map: error calling function at index %d: %w", i, err)
				}
				resultVal, ok := result.(value.Value)
				if !ok {
					return value.Value{}, fmt.Errorf("list.map: function returned non-Value type at index %d", i)
				}
				resultList[i] = resultVal
			}

			return value.List(resultList), nil
		},
	})
}
