package errors

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.ErrorMessage,
			Name:       "errorMessage",
			Arity:      1,
			ParamNames: []string{"e"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeError},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("errorMessage expects 1 argument, got %d", len(args))
			}
			arg := args[0].(value.Value)
			if arg.Kind != value.KindError {
				return value.Value{}, fmt.Errorf("errorMessage expects error, got %v", arg.Kind)
			}
			msg := arg.Str
			if arg.Error != nil && arg.Error.Message != "" {
				msg = arg.Error.Message
			}
			return value.Str(msg), nil
		},
	})
}
