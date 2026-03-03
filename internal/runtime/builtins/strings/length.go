package strings

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.StringLength,
			Name:       "length",
			Arity:      1, // receiver only
			ParamNames: []string{"self"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString}, // receiver: string
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeString,
			MethodName:   "length",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("string.length expects 1 argument (receiver), got %d", len(args))
			}
			receiver := args[0].(value.Value)

			if receiver.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.length called on non-string type %v", receiver.Kind)
			}

			return value.Int(int64(len(receiver.Str))), nil
		},
	})
}
