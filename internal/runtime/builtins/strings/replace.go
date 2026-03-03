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
			ID:         builtins.StringReplace,
			Name:       "replace",
			Arity:      3, // receiver + old + new
			ParamNames: []string{"self", "old", "new"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString}, // receiver: string
				{Kind: builtins.TypeString}, // old: string
				{Kind: builtins.TypeString}, // new: string
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeString,
			MethodName:   "replace",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 3 {
				return value.Value{}, fmt.Errorf("string.replace expects 3 arguments (receiver + old + new), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			oldStr := args[1].(value.Value)
			newStr := args[2].(value.Value)

			if receiver.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.replace called on non-string type %v", receiver.Kind)
			}
			if oldStr.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.replace: old argument must be string, got %v", oldStr.Kind)
			}
			if newStr.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.replace: new argument must be string, got %v", newStr.Kind)
			}

			result := strings.ReplaceAll(receiver.Str, oldStr.Str, newStr.Str)
			return value.Str(result), nil
		},
	})
}
