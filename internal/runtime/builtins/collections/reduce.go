package collections

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.ListReduce,
			Name:       "reduce",
			Arity:      3, // receiver + initial + reducer
			ParamNames: []string{"self", "initial", "reducer"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, // receiver: list<any>
				{Kind: builtins.TypeAny}, // initial: any (accumulator initial value)
				{Kind: builtins.TypeAny}, // reducer: function (type checked at compile time)
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeList,
			MethodName:   "reduce",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 3 {
				return value.Value{}, fmt.Errorf("list.reduce expects 3 arguments (receiver + initial + reducer), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			initial := args[1].(value.Value)
			reducerVal := args[2].(value.Value)

			if receiver.Kind != value.KindList {
				return value.Value{}, fmt.Errorf("list.reduce called on non-list type %v", receiver.Kind)
			}
			if reducerVal.Kind != value.KindClosure {
				return value.Value{}, fmt.Errorf("list.reduce: reducer argument must be a function, got %v", reducerVal.Kind)
			}

			// Start with the initial accumulator value
			acc := initial

			// Apply reducer to each element
			for i, elem := range receiver.List {
				// Call the reducer with (accumulator, element)
				callArgs := []interface{}{acc, elem}
				result, err := env.CallClosure(reducerVal.Closure, callArgs)
				if err != nil {
					return value.Value{}, fmt.Errorf("list.reduce: error calling reducer at index %d: %w", i, err)
				}
				resultVal, ok := result.(value.Value)
				if !ok {
					return value.Value{}, fmt.Errorf("list.reduce: reducer returned non-Value type at index %d", i)
				}
				acc = resultVal
			}

			return acc, nil
		},
	})
}
