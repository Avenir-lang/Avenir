package io

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.Input,
			Name:       "input",
			Arity:      0,
			ParamNames: []string{},
			Params:     []builtins.TypeRef{},
			Result:     builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 0 {
				return value.Value{}, fmt.Errorf("input expects 0 arguments, got %d", len(args))
			}
			if env == nil || env.IO() == nil {
				return value.Value{}, fmt.Errorf("runtime env IO is nil")
			}
			line, err := env.IO().ReadLine()
			if err != nil {
				return value.Value{}, fmt.Errorf("input failed: %w", err)
			}
			return value.Value{
				Kind: value.KindString,
				Str:  line,
			}, nil
		},
	})
}
