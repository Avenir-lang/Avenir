package strings

import (
	"fmt"
	"strings"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.StringStartsWith,
			Name:       "startsWith",
			Arity:      2, // receiver + prefix
			ParamNames: []string{"self", "prefix"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString}, // receiver: string
				{Kind: builtins.TypeString}, // prefix: string
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBool},
			ReceiverType: builtins.TypeString,
			MethodName:   "startsWith",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("string.startsWith expects 2 arguments (receiver + prefix), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			prefix := args[1].(value.Value)

			if receiver.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.startsWith called on non-string type %v", receiver.Kind)
			}
			if prefix.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.startsWith: prefix argument must be string, got %v", prefix.Kind)
			}

			result := strings.HasPrefix(receiver.Str, prefix.Str)
			return value.Bool(result), nil
		},
	})
}
