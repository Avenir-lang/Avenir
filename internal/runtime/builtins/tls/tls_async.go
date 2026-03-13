package tls

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	registerAsyncConnect()
	registerAsyncConnectConfig()
	registerAsyncAccept()
	registerAsyncRead()
	registerAsyncWrite()
	registerAsyncClose()
	registerAsyncListenerClose()
	registerAsyncHTTPSRequest()
}

func registerAsyncConnect() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncTLSConnect,
			Name:       "__builtin_async_tls_connect",
			Arity:      3,
			ParamNames: []string{"host", "port", "serverName"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeInt},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("__builtin_async_tls_connect expects 3 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return nil, fmt.Errorf("runtime tls service is nil")
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			snVal := args[2].(value.Value)
			if hostVal.Kind != value.KindString || portVal.Kind != value.KindInt || snVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_tls_connect: invalid argument types")
			}
			host, port, sn := hostVal.Str, int(portVal.Int), snVal.Str
			tlsSvc := env.TLS()
			return builtins.RunAsync(func() (interface{}, error) {
				handle, err := tlsSvc.Connect(host, port, sn)
				if err != nil {
					return nil, err
				}
				return value.Bytes(append([]byte(nil), handle...)), nil
			}), nil
		},
	})
}

func registerAsyncConnectConfig() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncTLSConnectConfig,
			Name:       "__builtin_async_tls_connect_config",
			Arity:      3,
			ParamNames: []string{"host", "port", "config"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeInt},
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("__builtin_async_tls_connect_config expects 3 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return nil, fmt.Errorf("runtime tls service is nil")
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			cfgVal := args[2].(value.Value)
			if hostVal.Kind != value.KindString || portVal.Kind != value.KindInt {
				return nil, fmt.Errorf("__builtin_async_tls_connect_config: invalid argument types")
			}
			cfg, err := extractTLSConfig(cfgVal)
			if err != nil {
				return nil, err
			}
			host, port := hostVal.Str, int(portVal.Int)
			tlsSvc := env.TLS()
			return builtins.RunAsync(func() (interface{}, error) {
				handle, err := tlsSvc.ConnectConfig(host, port, cfg)
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
			ID:         builtins.AsyncTLSAccept,
			Name:       "__builtin_async_tls_accept",
			Arity:      1,
			ParamNames: []string{"listener"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_tls_accept expects 1 argument, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return nil, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			tlsSvc := env.TLS()
			return builtins.RunAsync(func() (interface{}, error) {
				connHandle, err := tlsSvc.Accept(handle)
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
			ID:         builtins.AsyncTLSRead,
			Name:       "__builtin_async_tls_read",
			Arity:      2,
			ParamNames: []string{"conn", "n"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__builtin_async_tls_read expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return nil, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			nVal := args[1].(value.Value)
			if nVal.Kind != value.KindInt {
				return nil, fmt.Errorf("__builtin_async_tls_read expects n as int")
			}
			n := int(nVal.Int)
			tlsSvc := env.TLS()
			return builtins.RunAsync(func() (interface{}, error) {
				data, err := tlsSvc.Read(handle, n)
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
			ID:         builtins.AsyncTLSWrite,
			Name:       "__builtin_async_tls_write",
			Arity:      2,
			ParamNames: []string{"conn", "data"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeBytes},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("__builtin_async_tls_write expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return nil, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			dataVal := args[1].(value.Value)
			if dataVal.Kind != value.KindBytes {
				return nil, fmt.Errorf("__builtin_async_tls_write expects data as bytes")
			}
			data := append([]byte(nil), dataVal.Bytes...)
			tlsSvc := env.TLS()
			return builtins.RunAsync(func() (interface{}, error) {
				n, err := tlsSvc.Write(handle, data)
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
			ID:         builtins.AsyncTLSClose,
			Name:       "__builtin_async_tls_close",
			Arity:      1,
			ParamNames: []string{"conn"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_tls_close expects 1 argument, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return nil, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			tlsSvc := env.TLS()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := tlsSvc.Close(handle); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}

func registerAsyncListenerClose() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncTLSListenerClose,
			Name:       "__builtin_async_tls_listener_close",
			Arity:      1,
			ParamNames: []string{"listener"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_tls_listener_close expects 1 argument, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return nil, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			tlsSvc := env.TLS()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := tlsSvc.CloseListener(handle); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}

func registerAsyncHTTPSRequest() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncHTTPSRequest,
			Name:       "__builtin_async_https_request",
			Arity:      5,
			ParamNames: []string{"method", "url", "headers", "body", "config"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeString}}},
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 5 {
				return nil, fmt.Errorf("__builtin_async_https_request expects 5 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return nil, fmt.Errorf("runtime tls service is nil")
			}
			methodVal := args[0].(value.Value)
			urlVal := args[1].(value.Value)
			headersVal := args[2].(value.Value)
			bodyVal := args[3].(value.Value)
			cfgVal := args[4].(value.Value)

			if methodVal.Kind != value.KindString || urlVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_https_request: method and url must be strings")
			}
			headers := make(map[string]string)
			if headersVal.Kind == value.KindDict && headersVal.Dict != nil {
				for k, v := range headersVal.Dict {
					if v.Kind == value.KindString {
						headers[k] = v.Str
					}
				}
			}
			var body []byte
			if bodyVal.Kind == value.KindBytes {
				body = append([]byte(nil), bodyVal.Bytes...)
			}
			cfg, err := extractTLSConfig(cfgVal)
			if err != nil {
				return nil, err
			}

			method, url := methodVal.Str, urlVal.Str
			tlsSvc := env.TLS()
			return builtins.RunAsync(func() (interface{}, error) {
				resp, err := tlsSvc.HTTPSRequest(method, url, headers, body, cfg)
				if err != nil {
					return nil, err
				}
				respHeaders := make(map[string]value.Value, len(resp.Headers))
				for k, v := range resp.Headers {
					respHeaders[k] = value.Str(v)
				}
				return value.Dict(map[string]value.Value{
					"status":  value.Int(int64(resp.Status)),
					"headers": value.Dict(respHeaders),
					"body":    value.Bytes(resp.Body),
				}), nil
			}), nil
		},
	})
}
