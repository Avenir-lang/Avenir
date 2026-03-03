package collections

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.ListSlice,
			Name:       "slice",
			Arity:      3, // receiver + start + end
			ParamNames: []string{"self", "start", "end"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}}, // receiver: list<any>
				{Kind: builtins.TypeInt}, // start: int
				{Kind: builtins.TypeInt}, // end: int
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeList,
			MethodName:   "slice",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 3 {
				return value.Value{}, fmt.Errorf("list.slice expects 3 arguments (receiver + start + end), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			startVal := args[1].(value.Value)
			endVal := args[2].(value.Value)

			if receiver.Kind != value.KindList {
				return value.Value{}, fmt.Errorf("list.slice called on non-list type %v", receiver.Kind)
			}
			if startVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("list.slice: start argument must be int, got %v", startVal.Kind)
			}
			if endVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("list.slice: end argument must be int, got %v", endVal.Kind)
			}

			start := int(startVal.Int)
			end := int(endVal.Int)
			listLen := len(receiver.List)

			// Validate range bounds
			if start < 0 || start > listLen {
				return value.Value{}, fmt.Errorf("list.slice: start index %d out of bounds [0, %d]", start, listLen)
			}
			if end < 0 || end > listLen {
				return value.Value{}, fmt.Errorf("list.slice: end index %d out of bounds [0, %d]", end, listLen)
			}
			if start > end {
				return value.Value{}, fmt.Errorf("list.slice: start index %d must be <= end index %d", start, end)
			}

			// Create slice
			slice := receiver.List[start:end]
			resultList := make([]value.Value, len(slice))
			copy(resultList, slice)

			return value.List(resultList), nil
		},
	})
}
