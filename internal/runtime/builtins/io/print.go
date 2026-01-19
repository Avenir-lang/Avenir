package io

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.Print,
			Name:       "print",
			Arity:      1,
			ParamNames: []string{"value"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("print expects 1 argument, got %d", len(args))
			}
			if env == nil || env.IO() == nil {
				return value.Value{}, fmt.Errorf("runtime env IO is nil")
			}
			val := args[0].(value.Value)
			env.IO().Println(val.String())
			return val, nil
		},
	})
}
