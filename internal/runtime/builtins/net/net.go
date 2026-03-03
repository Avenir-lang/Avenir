package net

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	registerConnect()
	registerListen()
	registerAccept()
	registerRead()
	registerWrite()
	registerClose()
}

func registerConnect() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.SocketConnect,
			Name:       "__builtin_socket_connect",
			Arity:      2,
			ParamNames: []string{"host", "port"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("__builtin_socket_connect expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.Net() == nil {
				return value.Value{}, fmt.Errorf("runtime env net is nil")
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			if hostVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_socket_connect expects host as string")
			}
			if portVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("__builtin_socket_connect expects port as int")
			}
			handle, err := env.Net().Connect(hostVal.Str, int(portVal.Int))
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(append([]byte(nil), handle...)), nil
		},
	})
}

func registerListen() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.SocketListen,
			Name:       "__builtin_socket_listen",
			Arity:      2,
			ParamNames: []string{"host", "port"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("__builtin_socket_listen expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.Net() == nil {
				return value.Value{}, fmt.Errorf("runtime env net is nil")
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			if hostVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_socket_listen expects host as string")
			}
			if portVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("__builtin_socket_listen expects port as int")
			}
			handle, err := env.Net().Listen(hostVal.Str, int(portVal.Int))
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(append([]byte(nil), handle...)), nil
		},
	})
}

func registerAccept() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.SocketAccept,
			Name:       "__builtin_socket_accept",
			Arity:      1,
			ParamNames: []string{"server"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_socket_accept expects 1 argument, got %d", len(args))
			}
			if env == nil || env.Net() == nil {
				return value.Value{}, fmt.Errorf("runtime env net is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			connHandle, err := env.Net().Accept(handle)
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(append([]byte(nil), connHandle...)), nil
		},
	})
}

func registerRead() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.SocketRead,
			Name:       "__builtin_socket_read",
			Arity:      2,
			ParamNames: []string{"sock", "n"},
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
				return value.Value{}, fmt.Errorf("__builtin_socket_read expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.Net() == nil {
				return value.Value{}, fmt.Errorf("runtime env net is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			nVal := args[1].(value.Value)
			if nVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("__builtin_socket_read expects n as int")
			}
			data, err := env.Net().Read(handle, int(nVal.Int))
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
			ID:         builtins.SocketWrite,
			Name:       "__builtin_socket_write",
			Arity:      2,
			ParamNames: []string{"sock", "data"},
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
				return value.Value{}, fmt.Errorf("__builtin_socket_write expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.Net() == nil {
				return value.Value{}, fmt.Errorf("runtime env net is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			dataVal := args[1].(value.Value)
			if dataVal.Kind != value.KindBytes {
				return value.Value{}, fmt.Errorf("__builtin_socket_write expects data as bytes")
			}
			n, err := env.Net().Write(handle, dataVal.Bytes)
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
			ID:         builtins.SocketClose,
			Name:       "__builtin_socket_close",
			Arity:      1,
			ParamNames: []string{"sock"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_socket_close expects 1 argument, got %d", len(args))
			}
			if env == nil || env.Net() == nil {
				return value.Value{}, fmt.Errorf("runtime env net is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			if err := env.Net().Close(handle); err != nil {
				return value.Value{}, err
			}
			return value.Value{}, nil
		},
	})
}

func requireHandle(val value.Value) ([]byte, error) {
	if val.Kind != value.KindBytes {
		return nil, fmt.Errorf("socket handle must be bytes")
	}
	if len(val.Bytes) == 0 {
		return nil, fmt.Errorf("socket handle is empty")
	}
	return val.Bytes, nil
}
