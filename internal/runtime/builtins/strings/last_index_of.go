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
			ID:         builtins.StringLastIndexOf,
			Name:       "lastIndexOf",
			Arity:      2, // receiver + substr
			ParamNames: []string{"self", "substr"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString}, // receiver: string
				{Kind: builtins.TypeString}, // substr: string
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeString,
			MethodName:   "lastIndexOf",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("string.lastIndexOf expects 2 arguments (receiver + substr), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			substr := args[1].(value.Value)

			if receiver.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.lastIndexOf called on non-string type %v", receiver.Kind)
			}
			if substr.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.lastIndexOf: substr argument must be string, got %v", substr.Kind)
			}

			index := strings.LastIndex(receiver.Str, substr.Str)
			return value.Int(int64(index)), nil
		},
	})
}
