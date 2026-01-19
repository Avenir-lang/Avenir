package errors

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.Error,
			Name:       "error",
			Arity:      1,
			ParamNames: []string{"message"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeError},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("error expects 1 argument, got %d", len(args))
			}
			arg := args[0].(value.Value)
			if arg.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("error expects string, got %v", arg.Kind)
			}
			return value.ErrorValue(arg.Str), nil
		},
	})
}
