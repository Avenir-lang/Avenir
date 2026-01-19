package http

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

const requestHandleKey = "__handle"

func init() {
	registerRequest()
	registerListen()
	registerAccept()
	registerRespond()
}

func registerRequest() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTTPRequest,
			Name:       "__builtin_http_request",
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
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if env == nil {
				return value.Value{}, fmt.Errorf("runtime env is nil")
			}
			if env.HTTP() == nil {
				return value.Value{}, fmt.Errorf("http service is nil")
			}
			if len(args) != 4 {
				return value.Value{}, fmt.Errorf("http.request expects 4 arguments, got %d", len(args))
			}
			methodVal := args[0].(value.Value)
			urlVal := args[1].(value.Value)
			headersVal := args[2].(value.Value)
			bodyVal := args[3].(value.Value)

			if methodVal.Kind != value.KindString || urlVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("http.request expects method and url as strings")
			}
			headers, err := requireStringHeaders(headersVal, "http.request")
			if err != nil {
				return value.Value{}, err
			}
			body, err := optionalBytes(bodyVal, "http.request")
			if err != nil {
				return value.Value{}, err
			}

			resp, err := env.HTTP().Request(methodVal.Str, urlVal.Str, headers, body)
			if err != nil {
				return value.Value{}, err
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
		},
	})
}

func registerListen() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTTPListen,
			Name:       "__builtin_http_listen",
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
			if env == nil {
				return value.Value{}, fmt.Errorf("runtime env is nil")
			}
			if env.HTTP() == nil {
				return value.Value{}, fmt.Errorf("http service is nil")
			}
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("http.listen expects 2 arguments, got %d", len(args))
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			if hostVal.Kind != value.KindString || portVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("http.listen expects host string and port int")
			}
			handle, err := env.HTTP().Listen(hostVal.Str, int(portVal.Int))
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(handle), nil
		},
	})
}

func registerAccept() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTTPAccept,
			Name:       "__builtin_http_accept",
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
			if env == nil {
				return value.Value{}, fmt.Errorf("runtime env is nil")
			}
			if env.HTTP() == nil {
				return value.Value{}, fmt.Errorf("http service is nil")
			}
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("http.accept expects 1 argument, got %d", len(args))
			}
			handle, err := extractHandle(args[0].(value.Value), "http.accept")
			if err != nil {
				return value.Value{}, err
			}
			req, err := env.HTTP().Accept(handle)
			if err != nil {
				return value.Value{}, err
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
		},
	})
}

func registerRespond() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTTPRespond,
			Name:       "__builtin_http_respond",
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
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if env == nil {
				return value.Value{}, fmt.Errorf("runtime env is nil")
			}
			if env.HTTP() == nil {
				return value.Value{}, fmt.Errorf("http service is nil")
			}
			if len(args) != 4 {
				return value.Value{}, fmt.Errorf("http.respond expects 4 arguments, got %d", len(args))
			}
			reqHandle, err := extractHandle(args[0].(value.Value), "http.respond")
			if err != nil {
				return value.Value{}, err
			}
			statusVal := args[1].(value.Value)
			headersVal := args[2].(value.Value)
			bodyVal := args[3].(value.Value)
			if statusVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("http.respond expects status as int")
			}
			headers, err := requireStringHeaders(headersVal, "http.respond")
			if err != nil {
				return value.Value{}, err
			}
			if bodyVal.Kind != value.KindBytes {
				return value.Value{}, fmt.Errorf("http.respond expects body as bytes")
			}
			if err := env.HTTP().Respond(reqHandle, int(statusVal.Int), headers, bodyVal.Bytes); err != nil {
				return value.Value{}, err
			}
			return value.Value{}, nil
		},
	})
}

func extractHandle(v value.Value, name string) ([]byte, error) {
	switch v.Kind {
	case value.KindBytes:
		return v.Bytes, nil
	case value.KindDict:
		if v.Dict == nil {
			return nil, fmt.Errorf("%s expects request handle", name)
		}
		if handleVal, ok := v.Dict[requestHandleKey]; ok {
			if handleVal.Kind != value.KindBytes {
				return nil, fmt.Errorf("%s: handle must be bytes", name)
			}
			return handleVal.Bytes, nil
		}
		return nil, fmt.Errorf("%s expects request handle", name)
	default:
		return nil, fmt.Errorf("%s expects request handle", name)
	}
}

func requireStringHeaders(v value.Value, name string) (map[string]string, error) {
	if v.Kind != value.KindDict {
		return nil, fmt.Errorf("%s expects headers as dict<string>", name)
	}
	headers := make(map[string]string, len(v.Dict))
	for k, hv := range v.Dict {
		if hv.Kind != value.KindString {
			return nil, fmt.Errorf("%s expects header values as string", name)
		}
		headers[k] = hv.Str
	}
	return headers, nil
}

func optionalBytes(v value.Value, name string) ([]byte, error) {
	switch v.Kind {
	case value.KindBytes:
		return v.Bytes, nil
	case value.KindOptional:
		if v.Optional == nil || !v.Optional.IsSome {
			return nil, nil
		}
		if v.Optional.Value.Kind != value.KindBytes {
			return nil, fmt.Errorf("%s expects optional bytes", name)
		}
		return v.Optional.Value.Bytes, nil
	default:
		return nil, fmt.Errorf("%s expects bytes or bytes?", name)
	}
}
