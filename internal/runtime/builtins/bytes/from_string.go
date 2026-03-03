package bytes

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.BytesFromString,
			Name:       "fromString",
			Arity:      1, // string argument
			ParamNames: []string{"s"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString}, // s: string
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeVoid, // Regular function, not a method
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("bytes.fromString expects 1 argument (s), got %d", len(args))
			}
			strVal := args[0].(value.Value)

			if strVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("bytes.fromString: s argument must be string, got %v", strVal.Kind)
			}

			// Convert string to bytes (UTF-8 encode)
			resultBytes := []byte(strVal.Str)
			return value.Bytes(resultBytes), nil
		},
	})
}
