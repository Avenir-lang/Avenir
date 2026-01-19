package bytes

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.BytesAppend,
			Name:       "append",
			Arity:      2, // receiver + byte value
			ParamNames: []string{"self", "b"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeBytes}, // receiver: bytes
				{Kind: builtins.TypeInt},   // b: int (byte value 0-255)
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeBytes,
			MethodName:   "append",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("bytes.append expects 2 arguments (receiver + b), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			byteVal := args[1].(value.Value)

			if receiver.Kind != value.KindBytes {
				return value.Value{}, fmt.Errorf("bytes.append called on non-bytes type %v", receiver.Kind)
			}
			if byteVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("bytes.append: b argument must be int, got %v", byteVal.Kind)
			}

			// Validate byte range [0, 255]
			if byteVal.Int < 0 || byteVal.Int > 255 {
				return value.Value{}, fmt.Errorf("bytes.append: byte value must be in range [0, 255], got %d", byteVal.Int)
			}

			// Create new bytes with appended byte
			newBytes := make([]byte, len(receiver.Bytes)+1)
			copy(newBytes, receiver.Bytes)
			newBytes[len(receiver.Bytes)] = byte(byteVal.Int)

			return value.Bytes(newBytes), nil
		},
	})
}
