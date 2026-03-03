package bytes

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.BytesConcat,
			Name:       "concat",
			Arity:      2, // receiver + other
			ParamNames: []string{"self", "other"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeBytes}, // receiver: bytes
				{Kind: builtins.TypeBytes}, // other: bytes
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeBytes,
			MethodName:   "concat",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("bytes.concat expects 2 arguments (receiver + other), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			other := args[1].(value.Value)

			if receiver.Kind != value.KindBytes {
				return value.Value{}, fmt.Errorf("bytes.concat called on non-bytes type %v", receiver.Kind)
			}
			if other.Kind != value.KindBytes {
				return value.Value{}, fmt.Errorf("bytes.concat: other argument must be bytes, got %v", other.Kind)
			}

			// Create new bytes by concatenating receiver and other
			newBytes := make([]byte, len(receiver.Bytes)+len(other.Bytes))
			copy(newBytes, receiver.Bytes)
			copy(newBytes[len(receiver.Bytes):], other.Bytes)

			return value.Bytes(newBytes), nil
		},
	})
}
