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
			ID:         builtins.StringTrim,
			Name:       "trim",
			Arity:      1, // receiver only
			ParamNames: []string{"self"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString}, // receiver: string
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeString,
			MethodName:   "trim",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("string.trim expects 1 argument (receiver), got %d", len(args))
			}
			receiver := args[0].(value.Value)

			if receiver.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.trim called on non-string type %v", receiver.Kind)
			}

			result := strings.TrimSpace(receiver.Str)
			return value.Str(result), nil
		},
	})
}
