package bytes

import (
	"fmt"
	"unicode/utf8"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.BytesToString,
			Name:       "toString",
			Arity:      1, // receiver only
			ParamNames: []string{"self"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeBytes}, // receiver: bytes
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeBytes,
			MethodName:   "toString",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("bytes.toString expects 1 argument (receiver), got %d", len(args))
			}
			receiver := args[0].(value.Value)

			if receiver.Kind != value.KindBytes {
				return value.Value{}, fmt.Errorf("bytes.toString called on non-bytes type %v", receiver.Kind)
			}

			// Validate UTF-8 encoding
			if !utf8.Valid(receiver.Bytes) {
				return value.Value{}, fmt.Errorf("bytes.toString: bytes are not valid UTF-8")
			}

			// Convert bytes to string (UTF-8 decode)
			result := string(receiver.Bytes)
			return value.Str(result), nil
		},
	})
}
