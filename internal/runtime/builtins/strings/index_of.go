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
			ID:         builtins.StringIndexOf,
			Name:       "indexOf",
			Arity:      2, // receiver + substr
			ParamNames: []string{"self", "substr"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString}, // receiver: string
				{Kind: builtins.TypeString}, // substr: string
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeString,
			MethodName:   "indexOf",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("string.indexOf expects 2 arguments (receiver + substr), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			substr := args[1].(value.Value)

			if receiver.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.indexOf called on non-string type %v", receiver.Kind)
			}
			if substr.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.indexOf: substr argument must be string, got %v", substr.Kind)
			}

			index := strings.Index(receiver.Str, substr.Str)
			return value.Int(int64(index)), nil
		},
	})
}
