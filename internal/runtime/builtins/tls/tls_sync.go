package tls

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	registerConnect()
	registerConnectConfig()
	registerListen()
	registerListenConfig()
	registerAccept()
	registerRead()
	registerWrite()
	registerClose()
	registerListenerClose()
	registerLoadCert()
	registerLoadCertChain()
	registerLoadCertPEM()
	registerCertInfo()
	registerPeerCerts()
	registerNegotiatedProtocol()
	registerTLSVersion()
	registerHTTPSListen()
	registerHTTPSListenConfig()
	registerHTTPSListenAuto()
}

func requireHandle(val value.Value) ([]byte, error) {
	if val.Kind != value.KindBytes {
		return nil, fmt.Errorf("tls handle must be bytes")
	}
	if len(val.Bytes) == 0 {
		return nil, fmt.Errorf("tls handle is empty")
	}
	return val.Bytes, nil
}

func extractTLSConfig(val value.Value) (*builtins.TLSConfigData, error) {
	cfg := &builtins.TLSConfigData{}
	if val.Kind == value.KindDict {
		d := val.Dict
		if d == nil {
			return cfg, nil
		}
		if v, ok := d["certFile"]; ok && v.Kind == value.KindString {
			cfg.CertFile = v.Str
		}
		if v, ok := d["keyFile"]; ok && v.Kind == value.KindString {
			cfg.KeyFile = v.Str
		}
		if v, ok := d["minVersion"]; ok && v.Kind == value.KindString {
			cfg.MinVersion = v.Str
		}
		if v, ok := d["maxVersion"]; ok && v.Kind == value.KindString {
			cfg.MaxVersion = v.Str
		}
		if v, ok := d["clientAuth"]; ok && v.Kind == value.KindString {
			cfg.ClientAuth = v.Str
		}
		if v, ok := d["serverName"]; ok && v.Kind == value.KindString {
			cfg.ServerName = v.Str
		}
		if v, ok := d["insecureSkipVerify"]; ok && v.Kind == value.KindBool {
			cfg.InsecureSkipVerify = v.Bool
		}
		if v, ok := d["alpnProtocols"]; ok && v.Kind == value.KindList {
			for _, item := range v.List {
				if item.Kind == value.KindString {
					cfg.ALPNProtocols = append(cfg.ALPNProtocols, item.Str)
				}
			}
		}
		if v, ok := d["clientCAs"]; ok && v.Kind == value.KindList {
			for _, item := range v.List {
				if item.Kind == value.KindString {
					cfg.ClientCAs = append(cfg.ClientCAs, item.Str)
				}
			}
		}
		return cfg, nil
	}
	return nil, fmt.Errorf("expected TLS Config dict, got %v", val.Kind)
}

func registerConnect() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TLSConnect,
			Name:       "__builtin_tls_connect",
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
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 3 {
				return value.Value{}, fmt.Errorf("__builtin_tls_connect expects 3 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			snVal := args[2].(value.Value)
			if hostVal.Kind != value.KindString || portVal.Kind != value.KindInt || snVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_tls_connect: invalid argument types")
			}
			handle, err := env.TLS().Connect(hostVal.Str, int(portVal.Int), snVal.Str)
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(append([]byte(nil), handle...)), nil
		},
	})
}

func registerConnectConfig() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TLSConnectConfig,
			Name:       "__builtin_tls_connect_config",
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
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 3 {
				return value.Value{}, fmt.Errorf("__builtin_tls_connect_config expects 3 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			cfgVal := args[2].(value.Value)
			if hostVal.Kind != value.KindString || portVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("__builtin_tls_connect_config: invalid argument types")
			}
			cfg, err := extractTLSConfig(cfgVal)
			if err != nil {
				return value.Value{}, err
			}
			handle, err := env.TLS().ConnectConfig(hostVal.Str, int(portVal.Int), cfg)
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
			ID:         builtins.TLSListen,
			Name:       "__builtin_tls_listen",
			Arity:      4,
			ParamNames: []string{"host", "port", "certFile", "keyFile"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeInt},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 4 {
				return value.Value{}, fmt.Errorf("__builtin_tls_listen expects 4 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			certVal := args[2].(value.Value)
			keyVal := args[3].(value.Value)
			if hostVal.Kind != value.KindString || portVal.Kind != value.KindInt ||
				certVal.Kind != value.KindString || keyVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_tls_listen: invalid argument types")
			}
			handle, err := env.TLS().Listen(hostVal.Str, int(portVal.Int), certVal.Str, keyVal.Str)
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(append([]byte(nil), handle...)), nil
		},
	})
}

func registerListenConfig() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TLSListenConfig,
			Name:       "__builtin_tls_listen_config",
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
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 3 {
				return value.Value{}, fmt.Errorf("__builtin_tls_listen_config expects 3 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			cfgVal := args[2].(value.Value)
			if hostVal.Kind != value.KindString || portVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("__builtin_tls_listen_config: invalid argument types")
			}
			cfg, err := extractTLSConfig(cfgVal)
			if err != nil {
				return value.Value{}, err
			}
			handle, err := env.TLS().ListenConfig(hostVal.Str, int(portVal.Int), cfg)
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
			ID:         builtins.TLSAccept,
			Name:       "__builtin_tls_accept",
			Arity:      1,
			ParamNames: []string{"listener"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_tls_accept expects 1 argument, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			connHandle, err := env.TLS().Accept(handle)
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
			ID:         builtins.TLSRead,
			Name:       "__builtin_tls_read",
			Arity:      2,
			ParamNames: []string{"conn", "n"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeBytes},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("__builtin_tls_read expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			nVal := args[1].(value.Value)
			if nVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("__builtin_tls_read expects n as int")
			}
			data, err := env.TLS().Read(handle, int(nVal.Int))
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
			ID:         builtins.TLSWrite,
			Name:       "__builtin_tls_write",
			Arity:      2,
			ParamNames: []string{"conn", "data"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
				{Kind: builtins.TypeBytes},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("__builtin_tls_write expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			dataVal := args[1].(value.Value)
			if dataVal.Kind != value.KindBytes {
				return value.Value{}, fmt.Errorf("__builtin_tls_write expects data as bytes")
			}
			n, err := env.TLS().Write(handle, dataVal.Bytes)
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
			ID:         builtins.TLSClose,
			Name:       "__builtin_tls_close",
			Arity:      1,
			ParamNames: []string{"conn"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_tls_close expects 1 argument, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			if err := env.TLS().Close(handle); err != nil {
				return value.Value{}, err
			}
			return value.Value{}, nil
		},
	})
}

func registerListenerClose() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TLSListenerClose,
			Name:       "__builtin_tls_listener_close",
			Arity:      1,
			ParamNames: []string{"listener"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_tls_listener_close expects 1 argument, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			if err := env.TLS().CloseListener(handle); err != nil {
				return value.Value{}, err
			}
			return value.Value{}, nil
		},
	})
}

func registerLoadCert() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TLSLoadCert,
			Name:       "__builtin_tls_load_cert",
			Arity:      2,
			ParamNames: []string{"certFile", "keyFile"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("__builtin_tls_load_cert expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			certVal := args[0].(value.Value)
			keyVal := args[1].(value.Value)
			if certVal.Kind != value.KindString || keyVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_tls_load_cert: expects string arguments")
			}
			handle, err := env.TLS().LoadCert(certVal.Str, keyVal.Str)
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(append([]byte(nil), handle...)), nil
		},
	})
}

func registerLoadCertChain() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TLSLoadCertChain,
			Name:       "__builtin_tls_load_cert_chain",
			Arity:      1,
			ParamNames: []string{"files"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeString}}},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_tls_load_cert_chain expects 1 argument, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			listVal := args[0].(value.Value)
			if listVal.Kind != value.KindList {
				return value.Value{}, fmt.Errorf("__builtin_tls_load_cert_chain: expects list argument")
			}
			var files []string
			for _, elem := range listVal.List {
				if elem.Kind != value.KindString {
					return value.Value{}, fmt.Errorf("__builtin_tls_load_cert_chain: list elements must be strings")
				}
				files = append(files, elem.Str)
			}
			handle, err := env.TLS().LoadCertChain(files)
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(append([]byte(nil), handle...)), nil
		},
	})
}

func registerLoadCertPEM() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TLSLoadCertPEM,
			Name:       "__builtin_tls_load_cert_pem",
			Arity:      2,
			ParamNames: []string{"certPEM", "keyPEM"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeBytes},
				{Kind: builtins.TypeBytes},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("__builtin_tls_load_cert_pem expects 2 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			certVal := args[0].(value.Value)
			keyVal := args[1].(value.Value)
			if certVal.Kind != value.KindBytes || keyVal.Kind != value.KindBytes {
				return value.Value{}, fmt.Errorf("__builtin_tls_load_cert_pem: expects bytes arguments")
			}
			handle, err := env.TLS().LoadCertPEM(certVal.Bytes, keyVal.Bytes)
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(append([]byte(nil), handle...)), nil
		},
	})
}

func registerCertInfo() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TLSCertInfo,
			Name:       "__builtin_tls_cert_info",
			Arity:      1,
			ParamNames: []string{"cert"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeDict, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_tls_cert_info expects 1 argument, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			info, err := env.TLS().CertInfo(handle)
			if err != nil {
				return value.Value{}, err
			}
			return certInfoToValue(info), nil
		},
	})
}

func registerPeerCerts() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TLSPeerCerts,
			Name:       "__builtin_tls_peer_certs",
			Arity:      1,
			ParamNames: []string{"conn"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeList, Elem: []builtins.TypeRef{{Kind: builtins.TypeAny}}},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_tls_peer_certs expects 1 argument, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			certs, err := env.TLS().PeerCerts(handle)
			if err != nil {
				return value.Value{}, err
			}
			var elements []value.Value
			for _, cert := range certs {
				elements = append(elements, certInfoToValue(cert))
			}
			return value.List(elements), nil
		},
	})
}

func registerNegotiatedProtocol() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TLSNegotiatedProtocol,
			Name:       "__builtin_tls_negotiated_protocol",
			Arity:      1,
			ParamNames: []string{"conn"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_tls_negotiated_protocol expects 1 argument, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			proto, err := env.TLS().NegotiatedProtocol(handle)
			if err != nil {
				return value.Value{}, err
			}
			return value.Str(proto), nil
		},
	})
}

func registerTLSVersion() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TLSVersion,
			Name:       "__builtin_tls_version",
			Arity:      1,
			ParamNames: []string{"conn"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_tls_version expects 1 argument, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			handle, err := requireHandle(args[0].(value.Value))
			if err != nil {
				return value.Value{}, err
			}
			ver, err := env.TLS().TLSVersion(handle)
			if err != nil {
				return value.Value{}, err
			}
			return value.Str(ver), nil
		},
	})
}

func registerHTTPSListen() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTTPSListen,
			Name:       "__builtin_https_listen",
			Arity:      4,
			ParamNames: []string{"host", "port", "certFile", "keyFile"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeInt},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 4 {
				return value.Value{}, fmt.Errorf("__builtin_https_listen expects 4 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			certVal := args[2].(value.Value)
			keyVal := args[3].(value.Value)
			if hostVal.Kind != value.KindString || portVal.Kind != value.KindInt ||
				certVal.Kind != value.KindString || keyVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_https_listen: invalid argument types")
			}
			handle, err := env.TLS().Listen(hostVal.Str, int(portVal.Int), certVal.Str, keyVal.Str)
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(append([]byte(nil), handle...)), nil
		},
	})
}

func registerHTTPSListenConfig() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTTPSListenConfig,
			Name:       "__builtin_https_listen_config",
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
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 3 {
				return value.Value{}, fmt.Errorf("__builtin_https_listen_config expects 3 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			cfgVal := args[2].(value.Value)
			if hostVal.Kind != value.KindString || portVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("__builtin_https_listen_config: invalid argument types")
			}
			cfg, err := extractTLSConfig(cfgVal)
			if err != nil {
				return value.Value{}, err
			}
			handle, err := env.TLS().ListenConfig(hostVal.Str, int(portVal.Int), cfg)
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(append([]byte(nil), handle...)), nil
		},
	})
}

func registerHTTPSListenAuto() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.HTTPSListenAuto,
			Name:       "__builtin_https_listen_auto",
			Arity:      4,
			ParamNames: []string{"host", "port", "domain", "email"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeInt},
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 4 {
				return value.Value{}, fmt.Errorf("__builtin_https_listen_auto expects 4 arguments, got %d", len(args))
			}
			if env == nil || env.TLS() == nil {
				return value.Value{}, fmt.Errorf("runtime tls service is nil")
			}
			hostVal := args[0].(value.Value)
			portVal := args[1].(value.Value)
			domainVal := args[2].(value.Value)
			emailVal := args[3].(value.Value)
			if hostVal.Kind != value.KindString || portVal.Kind != value.KindInt ||
				domainVal.Kind != value.KindString || emailVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_https_listen_auto: invalid argument types")
			}
			handle, err := env.TLS().ListenAutoTLS(hostVal.Str, int(portVal.Int), domainVal.Str, emailVal.Str)
			if err != nil {
				return value.Value{}, err
			}
			return value.Bytes(append([]byte(nil), handle...)), nil
		},
	})
}

func certInfoToValue(info map[string]interface{}) value.Value {
	result := make(map[string]value.Value)
	if v, ok := info["subject"].(string); ok {
		result["subject"] = value.Str(v)
	}
	if v, ok := info["issuer"].(string); ok {
		result["issuer"] = value.Str(v)
	}
	if v, ok := info["notBefore"].(int64); ok {
		result["notBefore"] = value.Int(v)
	}
	if v, ok := info["notAfter"].(int64); ok {
		result["notAfter"] = value.Int(v)
	}
	if v, ok := info["isCA"].(bool); ok {
		result["isCA"] = value.Bool(v)
	}
	if v, ok := info["dnsNames"].([]string); ok {
		var elems []value.Value
		for _, name := range v {
			elems = append(elems, value.Str(name))
		}
		result["dnsNames"] = value.List(elems)
	}
	return value.Dict(result)
}
