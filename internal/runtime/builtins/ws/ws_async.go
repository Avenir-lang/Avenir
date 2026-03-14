package ws

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	registerAsyncUpgrade()
	registerAsyncSendText()
	registerAsyncSendBytes()
	registerAsyncSendPing()
	registerAsyncReceive()
	registerAsyncClose()
	registerSetReadLimit()
	registerGetInfo()
}

func requireHandle(v value.Value) ([]byte, error) {
	if v.Kind != value.KindBytes {
		return nil, fmt.Errorf("ws: expected bytes handle, got %v", v.Kind)
	}
	return v.Bytes, nil
}

func registerAsyncUpgrade() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncWSUpgrade,
			Name:       "__builtin_async_ws_upgrade",
			Arity:      3,
			ParamNames: []string{"reqHandle", "protocols", "extraHeaders"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if env == nil || env.WS() == nil {
				return nil, fmt.Errorf("ws service is nil")
			}
			if len(args) != 3 {
				return nil, fmt.Errorf("__builtin_async_ws_upgrade expects 3 arguments, got %d", len(args))
			}
			handleVal := args[0].(value.Value)
			protocolsVal := args[1].(value.Value)
			headersVal := args[2].(value.Value)

			handle, err := requireHandle(handleVal)
			if err != nil {
				return nil, err
			}

			var protocols []string
			if protocolsVal.Kind == value.KindList {
				for _, item := range protocolsVal.List {
					if item.Kind == value.KindString {
						protocols = append(protocols, item.Str)
					}
				}
			}

			extraHeaders := make(map[string]string)
			if headersVal.Kind == value.KindDict {
				for k, v := range headersVal.Dict {
					if v.Kind == value.KindString {
						extraHeaders[k] = v.Str
					}
				}
			}

			wsSvc := env.WS()
			reqHandle := append([]byte(nil), handle...)
			return builtins.RunAsync(func() (interface{}, error) {
				result, err := wsSvc.Upgrade(reqHandle, protocols, extraHeaders)
				if err != nil {
					return nil, err
				}
				return value.Dict(map[string]value.Value{
					"handle":   value.Bytes(result.Handle),
					"protocol": value.Str(result.Protocol),
				}), nil
			}), nil
		},
	})
}

func registerAsyncSendText() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncWSSendText,
			Name:       "__builtin_async_ws_send_text",
			Arity:      2,
			ParamNames: []string{"wsHandle", "text"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if env == nil || env.WS() == nil {
				return nil, fmt.Errorf("ws service is nil")
			}
			if len(args) != 2 {
				return nil, fmt.Errorf("__builtin_async_ws_send_text expects 2 arguments, got %d", len(args))
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			textVal := args[1].(value.Value)
			if textVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_ws_send_text expects text as string")
			}

			wsHandle := append([]byte(nil), handle...)
			text := textVal.Str
			wsSvc := env.WS()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := wsSvc.SendText(wsHandle, text); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}

func registerAsyncSendBytes() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncWSSendBytes,
			Name:       "__builtin_async_ws_send_bytes",
			Arity:      2,
			ParamNames: []string{"wsHandle", "data"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeBytes},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if env == nil || env.WS() == nil {
				return nil, fmt.Errorf("ws service is nil")
			}
			if len(args) != 2 {
				return nil, fmt.Errorf("__builtin_async_ws_send_bytes expects 2 arguments, got %d", len(args))
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			dataVal := args[1].(value.Value)
			if dataVal.Kind != value.KindBytes {
				return nil, fmt.Errorf("__builtin_async_ws_send_bytes expects data as bytes")
			}

			wsHandle := append([]byte(nil), handle...)
			data := append([]byte(nil), dataVal.Bytes...)
			wsSvc := env.WS()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := wsSvc.SendBytes(wsHandle, data); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}

func registerAsyncSendPing() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncWSSendPing,
			Name:       "__builtin_async_ws_send_ping",
			Arity:      2,
			ParamNames: []string{"wsHandle", "data"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeBytes},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if env == nil || env.WS() == nil {
				return nil, fmt.Errorf("ws service is nil")
			}
			if len(args) != 2 {
				return nil, fmt.Errorf("__builtin_async_ws_send_ping expects 2 arguments, got %d", len(args))
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			dataVal := args[1].(value.Value)
			if dataVal.Kind != value.KindBytes {
				return nil, fmt.Errorf("__builtin_async_ws_send_ping expects data as bytes")
			}

			wsHandle := append([]byte(nil), handle...)
			data := append([]byte(nil), dataVal.Bytes...)
			wsSvc := env.WS()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := wsSvc.SendPing(wsHandle, data); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}

func registerAsyncReceive() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncWSReceive,
			Name:       "__builtin_async_ws_receive",
			Arity:      1,
			ParamNames: []string{"wsHandle"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if env == nil || env.WS() == nil {
				return nil, fmt.Errorf("ws service is nil")
			}
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_ws_receive expects 1 argument, got %d", len(args))
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}

			wsHandle := append([]byte(nil), handle...)
			wsSvc := env.WS()
			return builtins.RunAsync(func() (interface{}, error) {
				msg, err := wsSvc.Receive(wsHandle)
				if err != nil {
					return nil, err
				}
				text := ""
				if msg.Type == 1 {
					text = string(msg.Data)
				}
				return value.Dict(map[string]value.Value{
					"type": value.Int(int64(msg.Type)),
					"text": value.Str(text),
					"data": value.Bytes(msg.Data),
					"code": value.Int(int64(msg.Code)),
				}), nil
			}), nil
		},
	})
}

func registerAsyncClose() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.AsyncWSClose,
			Name:       "__builtin_async_ws_close",
			Arity:      3,
			ParamNames: []string{"wsHandle", "code", "reason"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeInt},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if env == nil || env.WS() == nil {
				return nil, fmt.Errorf("ws service is nil")
			}
			if len(args) != 3 {
				return nil, fmt.Errorf("__builtin_async_ws_close expects 3 arguments, got %d", len(args))
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			codeVal := args[1].(value.Value)
			reasonVal := args[2].(value.Value)
			if codeVal.Kind != value.KindInt {
				return nil, fmt.Errorf("__builtin_async_ws_close expects code as int")
			}
			if reasonVal.Kind != value.KindString {
				return nil, fmt.Errorf("__builtin_async_ws_close expects reason as string")
			}

			wsHandle := append([]byte(nil), handle...)
			code := int(codeVal.Int)
			reason := reasonVal.Str
			wsSvc := env.WS()
			return builtins.RunAsync(func() (interface{}, error) {
				if err := wsSvc.Close(wsHandle, code, reason); err != nil {
					return nil, err
				}
				return value.Value{}, nil
			}), nil
		},
	})
}

func registerSetReadLimit() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.WSSetReadLimit,
			Name:       "__builtin_ws_set_read_limit",
			Arity:      2,
			ParamNames: []string{"wsHandle", "limit"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if env == nil || env.WS() == nil {
				return nil, fmt.Errorf("ws service is nil")
			}
			if len(args) != 2 {
				return nil, fmt.Errorf("__builtin_ws_set_read_limit expects 2 arguments, got %d", len(args))
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			limitVal := args[1].(value.Value)
			if limitVal.Kind != value.KindInt {
				return nil, fmt.Errorf("__builtin_ws_set_read_limit expects limit as int")
			}
			if err := env.WS().SetReadLimit(handle, limitVal.Int); err != nil {
				return nil, err
			}
			return value.Value{}, nil
		},
	})
}

func registerGetInfo() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.WSGetInfo,
			Name:       "__builtin_ws_get_info",
			Arity:      1,
			ParamNames: []string{"wsHandle"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if env == nil || env.WS() == nil {
				return nil, fmt.Errorf("ws service is nil")
			}
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_ws_get_info expects 1 argument, got %d", len(args))
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return nil, err
			}
			info, err := env.WS().GetInfo(handle)
			if err != nil {
				return nil, err
			}
			headers := make(map[string]value.Value, len(info.Headers))
			for k, v := range info.Headers {
				headers[k] = value.Str(v)
			}
			return value.Dict(map[string]value.Value{
				"id":         value.Str(info.ID),
				"path":       value.Str(info.Path),
				"headers":    value.Dict(headers),
				"query":      value.Str(info.Query),
				"remoteAddr": value.Str(info.RemoteAddr),
			}), nil
		},
	})
}
