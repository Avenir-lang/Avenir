package fs

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	registerOpen()
	registerRead()
	registerWrite()
	registerClose()
	registerExists()
	registerRemove()
	registerMkdir()
	registerExecRoot()
}

func registerOpen() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.FSOpen,
			Name:       "__builtin_fs_open",
			Arity:      2,
			ParamNames: []string{"path", "mode"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("__builtin_fs_open expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return value.Value{}, fmt.Errorf("runtime env fs is nil")
			}
			pathVal := args[0].(value.Value)
			modeVal := args[1].(value.Value)
			if pathVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_fs_open expects path as string")
			}
			if modeVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_fs_open expects mode as string")
			}
			handle, err := env.FS().Open(pathVal.Str, modeVal.Str)
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(append([]byte(nil), handle...)), nil
		},
	})
}

func registerRead() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.FSRead,
			Name:       "__builtin_fs_read",
			Arity:      2,
			ParamNames: []string{"file", "n"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("__builtin_fs_read expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return value.Value{}, fmt.Errorf("runtime env fs is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			nVal := args[1].(value.Value)
			if nVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("__builtin_fs_read expects n as int")
			}
			data, err := env.FS().Read(handle, int(nVal.Int))
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(data), nil
		},
	})
}

func registerWrite() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.FSWrite,
			Name:       "__builtin_fs_write",
			Arity:      2,
			ParamNames: []string{"file", "data"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeBytes},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("__builtin_fs_write expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return value.Value{}, fmt.Errorf("runtime env fs is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			dataVal := args[1].(value.Value)
			if dataVal.Kind != value.KindBytes {
				return value.Value{}, fmt.Errorf("__builtin_fs_write expects data as bytes")
			}
			n, err := env.FS().Write(handle, dataVal.Bytes)
			if err != nil {
				return value.Value{}, err
			}
			return value.Int(int64(n)), nil
		},
	})
}

func registerClose() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.FSClose,
			Name:       "__builtin_fs_close",
			Arity:      1,
			ParamNames: []string{"file"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_fs_close expects 1 argument, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return value.Value{}, fmt.Errorf("runtime env fs is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			if err := env.FS().Close(handle); err != nil {
				return value.Value{}, err
			}
			return value.Value{}, nil
		},
	})
}

func registerExists() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.FSExists,
			Name:       "__builtin_fs_exists",
			Arity:      1,
			ParamNames: []string{"path"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBool},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_fs_exists expects 1 argument, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return value.Value{}, fmt.Errorf("runtime env fs is nil")
			}
			pathVal := args[0].(value.Value)
			if pathVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_fs_exists expects path as string")
			}
			ok, err := env.FS().Exists(pathVal.Str)
			if err != nil {
				return value.Value{}, err
			}
			return value.Bool(ok), nil
		},
	})
}

func registerRemove() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.FSRemove,
			Name:       "__builtin_fs_remove",
			Arity:      1,
			ParamNames: []string{"path"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_fs_remove expects 1 argument, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return value.Value{}, fmt.Errorf("runtime env fs is nil")
			}
			pathVal := args[0].(value.Value)
			if pathVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_fs_remove expects path as string")
			}
			if err := env.FS().Remove(pathVal.Str); err != nil {
				return value.Value{}, err
			}
			return value.Value{}, nil
		},
	})
}

func registerMkdir() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.FSMkdir,
			Name:       "__builtin_fs_mkdir",
			Arity:      1,
			ParamNames: []string{"path"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_fs_mkdir expects 1 argument, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return value.Value{}, fmt.Errorf("runtime env fs is nil")
			}
			pathVal := args[0].(value.Value)
			if pathVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_fs_mkdir expects path as string")
			}
			if err := env.FS().Mkdir(pathVal.Str); err != nil {
				return value.Value{}, err
			}
			return value.Value{}, nil
		},
	})
}

func requireHandle(val value.Value) ([]byte, error) {
	if val.Kind != value.KindBytes {
		return nil, fmt.Errorf("file handle must be bytes")
	}
	if len(val.Bytes) == 0 {
		return nil, fmt.Errorf("file handle is empty")
	}
	return val.Bytes, nil
}
