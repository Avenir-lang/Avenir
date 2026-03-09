package net

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	registerAsyncConnect()
	registerAsyncAccept()
	registerAsyncRead()
	registerAsyncWrite()
	registerAsyncClose()
}

func registerAsyncConnect() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSocketConnect,
			Name:       "__builtin_async_socket_connect",
			Arity:      2,
			ParamNames: []string{"host", "port"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__builtin_async_socket_connect expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.Net() == nil {
				return nil, fmt.Errorf("runtime env net is nil")
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			if hostVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_socket_connect expects host as string")
			}
			if portVal.Kind != value.KindInt {
				return nil, fmt.Errorf("__builtin_async_socket_connect expects port as int")
			}
			host, port := hostVal.Str, int(portVal.Int)
			netService := env.Net()
			return builtins.RunAsync(func() (interface{}, error) {
				handle, err := netService.Connect(host, port)
				if err != nil {
					return nil, err
				}
				return value.Bytes(append([]byte(nil), handle...)), nil
			}), nil
		},
	})
}

func registerAsyncAccept() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSocketAccept,
			Name:       "__builtin_async_socket_accept",
			Arity:      1,
			ParamNames: []string{"server"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_socket_accept expects 1 argument, got %d", len(args))
			}
			if env == nil || env.Net() == nil {
				return nil, fmt.Errorf("runtime env net is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			netService := env.Net()
			return builtins.RunAsync(func() (interface{}, error) {
				connHandle, err := netService.Accept(handle)
				if err != nil {
					return nil, err
				}
				return value.Bytes(append([]byte(nil), connHandle...)), nil
			}), nil
		},
	})
}

func registerAsyncRead() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncSocketRead,
			Name:       "__builtin_async_socket_read",
			Arity:      2,
			ParamNames: []string{"sock", "n"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__builtin_async_socket_read expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.Net() == nil {
				return nil, fmt.Errorf("runtime env net is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			nVal := args[1].(value.Value)
			if nVal.Kind != value.KindInt {
				return nil, fmt.Errorf("__builtin_async_socket_read expects n as int")
			}
			n := int(nVal.Int)
			netService := env.Net()
			return builtins.RunAsync(func() (interface{}, error) {
				data, err := netService.Read(handle, n)
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
			ID:         builtins.AsyncSocketWrite,
			Name:       "__builtin_async_socket_write",
			Arity:      2,
			ParamNames: []string{"sock", "data"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeBytes},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__builtin_async_socket_write expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.Net() == nil {
				return nil, fmt.Errorf("runtime env net is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			dataVal := args[1].(value.Value)
			if dataVal.Kind != value.KindBytes {
				return nil, fmt.Errorf("__builtin_async_socket_write expects data as bytes")
			}
			data := append([]byte(nil), dataVal.Bytes...)
			netService := env.Net()
			return builtins.RunAsync(func() (interface{}, error) {
				n, err := netService.Write(handle, data)
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
			ID:         builtins.AsyncSocketClose,
			Name:       "__builtin_async_socket_close",
			Arity:      1,
			ParamNames: []string{"sock"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_socket_close expects 1 argument, got %d", len(args))
			}
			if env == nil || env.Net() == nil {
				return nil, fmt.Errorf("runtime env net is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			netService := env.Net()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := netService.Close(handle); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}
