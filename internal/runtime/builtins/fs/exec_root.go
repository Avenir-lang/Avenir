package fs

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func registerExecRoot() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.FSExecRoot,
			Name:       "__builtin_fs_exec_root",
			Arity:      0,
			ParamNames: []string{},
			Params:     []builtins.TypeRef{},
			Result:     builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 0 {
				return value.Value{}, fmt.Errorf("__builtin_fs_exec_root expects 0 arguments, got %d", len(args))
			}
			if env == nil {
				return value.Value{}, fmt.Errorf("runtime env is nil")
			}
			return value.Str(env.ExecRoot()), nil
		},
	})
}
