package collections

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.Len,
			Name:       "len",
			Arity:      1,
			ParamNames: []string{"value"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny}, // Accept list<any> or bytes
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("len expects 1 argument, got %d", len(args))
			}
			arg := args[0].(value.Value)
			if arg.Kind == value.KindList {
				return value.Int(int64(len(arg.List))), nil
			}
			if arg.Kind == value.KindBytes {
				return value.Int(int64(len(arg.Bytes))), nil
			}
			return value.Value{}, fmt.Errorf("len expects list<T> or bytes, got %v", arg.Kind)
		},
	})
}
