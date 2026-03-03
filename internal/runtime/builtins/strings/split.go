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
			ID:         builtins.StringSplit,
			Name:       "split",
			Arity:      2, // receiver + sep
			ParamNames: []string{"self", "sep"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString}, // receiver: string
				{Kind: builtins.TypeString}, // sep: string
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeString}}},
			ReceiverType: builtins.TypeString,
			MethodName:   "split",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("string.split expects 2 arguments (receiver + sep), got %d", len(args))
			}
			receiver := args[0].(value.Value)
			sep := args[1].(value.Value)

			if receiver.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.split called on non-string type %v", receiver.Kind)
			}
			if sep.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("string.split: sep argument must be string, got %v", sep.Kind)
			}

			parts := strings.Split(receiver.Str, sep.Str)
			resultList := make([]value.Value, len(parts))
			for i, part := range parts {
				resultList[i] = value.Str(part)
			}

			return value.List(resultList), nil
		},
	})
}
