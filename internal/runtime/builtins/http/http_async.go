package http

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	registerAsyncRequest()
	registerAsyncAccept()
	registerAsyncRespond()
}

func registerAsyncRequest() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncHTTPRequest,
			Name:       "__builtin_async_http_request",
			Arity:      4,
			ParamNames: []string{"method", "url", "headers", "body"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeString}}},
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if env == nil {
				return nil, fmt.Errorf("runtime env is nil")
			}
			if env.HTTP() == nil {
				return nil, fmt.Errorf("http service is nil")
			}
			if len(args) != 4 {
				return nil, fmt.Errorf("__builtin_async_http_request expects 4 arguments, got %d", len(args))
			}
			methodVal := args[0].(value.Value)
			urlVal := args[1].(value.Value)
			headersVal := args[2].(value.Value)
			bodyVal := args[3].(value.Value)

			if methodVal.Kind != value.KindString || urlVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_http_request expects method and url as strings")
			}
			headers, err := requireStringHeaders(headersVal, "__builtin_async_http_request")
			if err != nil {
				return nil, err
			}
			body, err := optionalBytes(bodyVal, "__builtin_async_http_request")
			if err != nil {
				return nil, err
			}

			method, url := methodVal.Str, urlVal.Str
			httpService := env.HTTP()
			return builtins.RunAsync(func() (interface{}, error) {
				resp, err := httpService.Request(method, url, headers, body)
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

func registerAsyncAccept() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncHTTPAccept,
			Name:       "__builtin_async_http_accept",
			Arity:      1,
			ParamNames: []string{"server"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if env == nil {
				return nil, fmt.Errorf("runtime env is nil")
			}
			if env.HTTP() == nil {
				return nil, fmt.Errorf("http service is nil")
			}
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_http_accept expects 1 argument, got %d", len(args))
			}
			handle, err := extractHandle(args[0].(value.Value), "__builtin_async_http_accept")
			if err != nil {
				return nil, err
			}
			httpService := env.HTTP()
			return builtins.RunAsync(func() (interface{}, error) {
				req, err := httpService.Accept(handle)
				if err != nil {
					return nil, err
				}
				headers := make(map[string]value.Value, len(req.Headers))
				for k, v := range req.Headers {
					headers[k] = value.Str(v)
				}
				return value.Dict(map[string]value.Value{
					requestHandleKey: value.Bytes(req.Handle),
					"method":         value.Str(req.Method),
					"path":           value.Str(req.Path),
					"headers":        value.Dict(headers),
					"body":           value.Bytes(req.Body),
				}), nil
			}), nil
		},
	})
}

func registerAsyncRespond() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncHTTPRespond,
			Name:       "__builtin_async_http_respond",
			Arity:      4,
			ParamNames: []string{"req", "status", "headers", "body"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeInt},
				{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeString}}},
				{Kind: builtins.TypeBytes},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if env == nil {
				return nil, fmt.Errorf("runtime env is nil")
			}
			if env.HTTP() == nil {
				return nil, fmt.Errorf("http service is nil")
			}
			if len(args) != 4 {
				return nil, fmt.Errorf("__builtin_async_http_respond expects 4 arguments, got %d", len(args))
			}
			reqHandle, err := extractHandle(args[0].(value.Value), "__builtin_async_http_respond")
			if err != nil {
				return nil, err
			}
			statusVal := args[1].(value.Value)
			headersVal := args[2].(value.Value)
			bodyVal := args[3].(value.Value)
			if statusVal.Kind != value.KindInt {
				return nil, fmt.Errorf("__builtin_async_http_respond expects status as int")
			}
			headers, err := requireStringHeaders(headersVal, "__builtin_async_http_respond")
			if err != nil {
				return nil, err
			}
			if bodyVal.Kind != value.KindBytes {
				return nil, fmt.Errorf("__builtin_async_http_respond expects body as bytes")
			}

			status := int(statusVal.Int)
			bodyData := append([]byte(nil), bodyVal.Bytes...)
			httpService := env.HTTP()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := httpService.Respond(reqHandle, status, headers, bodyData); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}
