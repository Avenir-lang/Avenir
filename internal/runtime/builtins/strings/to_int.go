package strings

import (
	"fmt"
	"strconv"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.ToInt,
			Name:       "toInt",
			Arity:      1,
			ParamNames: []string{"value"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("toInt expects 1 argument, got %d", len(args))
			}
			arg := args[0].(value.Value)
			if arg.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("toInt expects string, got %v", arg.Kind)
			}
			parsed, err := strconv.Atoi(arg.Str)
			if err != nil {
				return value.Value{}, fmt.Errorf("toInt: invalid integer %q", arg.Str)
			}
			return value.Int(int64(parsed)), nil
		},
	})
}
