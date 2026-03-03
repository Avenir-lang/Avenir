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
			ID:         builtins.StringContains,
			Name:       "contains",
			Arity:      2, // receiver + substr
			ParamNames: []string{"self", "substr"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString}, // receiver: string
				{Kind: builtins.TypeString}, // substr: string
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBool},
			ReceiverType: builtins.TypeString,
			MethodName:   "contains",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("string.contains expects 2 arguments (receiver + substr), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			substr := args[1].(value.Value)

			if receiver.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.contains called on non-string type %v", receiver.Kind)
			}
			if substr.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.contains: substr argument must be string, got %v", substr.Kind)
			}

			result := strings.Contains(receiver.Str, substr.Str)
			return value.Bool(result), nil
		},
	})
}
