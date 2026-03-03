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
			ID:         builtins.StringEndsWith,
			Name:       "endsWith",
			Arity:      2, // receiver + suffix
			ParamNames: []string{"self", "suffix"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString}, // receiver: string
				{Kind: builtins.TypeString}, // suffix: string
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBool},
			ReceiverType: builtins.TypeString,
			MethodName:   "endsWith",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("string.endsWith expects 2 arguments (receiver + suffix), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			suffix := args[1].(value.Value)

			if receiver.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.endsWith called on non-string type %v", receiver.Kind)
			}
			if suffix.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.endsWith: suffix argument must be string, got %v", suffix.Kind)
			}

			result := strings.HasSuffix(receiver.Str, suffix.Str)
			return value.Bool(result), nil
		},
	})
}
