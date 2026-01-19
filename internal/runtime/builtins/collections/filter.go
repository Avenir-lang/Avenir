package collections

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.ListFilter,
			Name:       "filter",
			Arity:      2, // receiver + predicate
			ParamNames: []string{"self", "predicate"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, // receiver: list<any>
				{Kind: builtins.TypeAny}, // predicate: function (type checked at compile time)
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeList,
			MethodName:   "filter",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("list.filter expects 2 arguments (receiver + predicate), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			predVal := args[1].(value.Value)

			if receiver.Kind != value.KindList {
				return value.Value{}, fmt.Errorf("list.filter called on non-list type %v", receiver.Kind)
			}
			if predVal.Kind != value.KindClosure {
				return value.Value{}, fmt.Errorf("list.filter: predicate argument must be a function, got %v", predVal.Kind)
			}

			// Filter elements where predicate returns true
			resultList := make([]value.Value, 0, len(receiver.List))
			for i, elem := range receiver.List {
				// Call the predicate with the element as argument
				callArgs := []interface{}{elem}
				result, err := env.CallClosure(predVal.Closure, callArgs)
				if err != nil {
					return value.Value{}, fmt.Errorf("list.filter: error calling predicate at index %d: %w", i, err)
				}
				resultVal, ok := result.(value.Value)
				if !ok {
					return value.Value{}, fmt.Errorf("list.filter: predicate returned non-Value type at index %d", i)
				}
				if resultVal.Kind != value.KindBool {
					return value.Value{}, fmt.Errorf("list.filter: predicate must return bool, got %v at index %d", resultVal.Kind, i)
				}
				if resultVal.Bool {
					resultList = append(resultList, elem)
				}
			}

			return value.List(resultList), nil
		},
	})
}
