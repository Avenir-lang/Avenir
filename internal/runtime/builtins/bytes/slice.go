package bytes

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.BytesSlice,
			Name:       "slice",
			Arity:      3, // receiver + start + end
			ParamNames: []string{"self", "start", "end"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeBytes}, // receiver: bytes
				{Kind: builtins.TypeInt},   // start: int
				{Kind: builtins.TypeInt},   // end: int
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeBytes,
			MethodName:   "slice",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 3 {
				return value.Value{}, fmt.Errorf("bytes.slice expects 3 arguments (receiver + start + end), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			startVal := args[1].(value.Value)
			endVal := args[2].(value.Value)

			if receiver.Kind != value.KindBytes {
				return value.Value{}, fmt.Errorf("bytes.slice called on non-bytes type %v", receiver.Kind)
			}
			if startVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("bytes.slice: start argument must be int, got %v", startVal.Kind)
			}
			if endVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("bytes.slice: end argument must be int, got %v", endVal.Kind)
			}

			start := int(startVal.Int)
			end := int(endVal.Int)
			bytesLen := len(receiver.Bytes)

			// Validate range bounds: 0 ≤ start ≤ end ≤ length
			if start < 0 || start > bytesLen {
				return value.Value{}, fmt.Errorf("bytes.slice: start index %d out of bounds [0, %d]", start, bytesLen)
			}
			if end < 0 || end > bytesLen {
				return value.Value{}, fmt.Errorf("bytes.slice: end index %d out of bounds [0, %d]", end, bytesLen)
			}
			if start > end {
				return value.Value{}, fmt.Errorf("bytes.slice: start index %d must be <= end index %d", start, end)
			}

			// Create slice
			slice := receiver.Bytes[start:end]
			resultBytes := make([]byte, len(slice))
			copy(resultBytes, slice)

			return value.Bytes(resultBytes), nil
		},
	})
}
