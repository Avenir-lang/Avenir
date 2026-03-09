package fs

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	registerAsyncOpen()
	registerAsyncRead()
	registerAsyncWrite()
	registerAsyncClose()
	registerAsyncExists()
	registerAsyncRemove()
	registerAsyncMkdir()
}

func registerAsyncOpen() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncFSOpen,
			Name:       "__builtin_async_fs_open",
			Arity:      2,
			ParamNames: []string{"path", "mode"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__builtin_async_fs_open expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return nil, fmt.Errorf("runtime env fs is nil")
			}
			pathVal := args[0].(value.Value)
			modeVal := args[1].(value.Value)
			if pathVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_fs_open expects path as string")
			}
			if modeVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_fs_open expects mode as string")
			}
			p, m := pathVal.Str, modeVal.Str
			fsService := env.FS()
			return builtins.RunAsync(func() (interface{}, error) {
				handle, err := fsService.Open(p, m)
				if err != nil {
					return nil, err
				}
				return value.Bytes(append([]byte(nil), handle...)), nil
			}), nil
		},
	})
}

func registerAsyncRead() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncFSRead,
			Name:       "__builtin_async_fs_read",
			Arity:      2,
			ParamNames: []string{"file", "n"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__builtin_async_fs_read expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return nil, fmt.Errorf("runtime env fs is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			nVal := args[1].(value.Value)
			if nVal.Kind != value.KindInt {
				return nil, fmt.Errorf("__builtin_async_fs_read expects n as int")
			}
			n := int(nVal.Int)
			fsService := env.FS()
			return builtins.RunAsync(func() (interface{}, error) {
				data, err := fsService.Read(handle, n)
				if err != nil {
					return nil, err
				}
				return value.Bytes(data), nil
			}), nil
		},
	})
}

func registerAsyncWrite() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncFSWrite,
			Name:       "__builtin_async_fs_write",
			Arity:      2,
			ParamNames: []string{"file", "data"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeBytes},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__builtin_async_fs_write expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return nil, fmt.Errorf("runtime env fs is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			dataVal := args[1].(value.Value)
			if dataVal.Kind != value.KindBytes {
				return nil, fmt.Errorf("__builtin_async_fs_write expects data as bytes")
			}
			data := append([]byte(nil), dataVal.Bytes...)
			fsService := env.FS()
			return builtins.RunAsync(func() (interface{}, error) {
				n, err := fsService.Write(handle, data)
				if err != nil {
					return nil, err
				}
				return value.Int(int64(n)), nil
			}), nil
		},
	})
}

func registerAsyncClose() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncFSClose,
			Name:       "__builtin_async_fs_close",
			Arity:      1,
			ParamNames: []string{"file"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_fs_close expects 1 argument, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return nil, fmt.Errorf("runtime env fs is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			fsService := env.FS()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := fsService.Close(handle); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}

func registerAsyncExists() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncFSExists,
			Name:       "__builtin_async_fs_exists",
			Arity:      1,
			ParamNames: []string{"path"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBool},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_fs_exists expects 1 argument, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return nil, fmt.Errorf("runtime env fs is nil")
			}
			pathVal := args[0].(value.Value)
			if pathVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_fs_exists expects path as string")
			}
			p := pathVal.Str
			fsService := env.FS()
			return builtins.RunAsync(func() (interface{}, error) {
				ok, err := fsService.Exists(p)
				if err != nil {
					return nil, err
				}
				return value.Bool(ok), nil
			}), nil
		},
	})
}

func registerAsyncRemove() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncFSRemove,
			Name:       "__builtin_async_fs_remove",
			Arity:      1,
			ParamNames: []string{"path"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_fs_remove expects 1 argument, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return nil, fmt.Errorf("runtime env fs is nil")
			}
			pathVal := args[0].(value.Value)
			if pathVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_fs_remove expects path as string")
			}
			p := pathVal.Str
			fsService := env.FS()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := fsService.Remove(p); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}

func registerAsyncMkdir() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncFSMkdir,
			Name:       "__builtin_async_fs_mkdir",
			Arity:      1,
			ParamNames: []string{"path"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_fs_mkdir expects 1 argument, got %d", len(args))
			}
			if env == nil || env.FS() == nil {
				return nil, fmt.Errorf("runtime env fs is nil")
			}
			pathVal := args[0].(value.Value)
			if pathVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_fs_mkdir expects path as string")
			}
			p := pathVal.Str
			fsService := env.FS()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := fsService.Mkdir(p); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}
