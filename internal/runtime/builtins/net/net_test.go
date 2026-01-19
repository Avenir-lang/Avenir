package net_test

import (
	"net"
	"strconv"
	"testing"
	"time"

	"avenir/internal/runtime"
	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func callBuiltin(t *testing.T, env *runtime.Env, name string, args ...value.Value) value.Value {
	t.Helper()
	b := builtins.LookupByName(name)
	if b == nil {
		t.Fatalf("builtin %q not found", name)
	}
	argsIface := make([]interface{}, len(args))
	for i, arg := range args {
		argsIface[i] = arg
	}
	res, err := b.Call(env, argsIface)
	if err != nil {
		t.Fatalf("builtin %q error: %v", name, err)
	}
	val, ok := res.(value.Value)
	if !ok {
		t.Fatalf("builtin %q returned non-value %T", name, res)
	}
	return val
}

func TestSocketConnectClose(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().(*net.TCPAddr)
	done := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			_ = conn.Close()
		}
		close(done)
	}()

	env := runtime.DefaultEnv()
	handle := callBuiltin(t, env, "__builtin_socket_connect", value.Str("127.0.0.1"), value.Int(int64(addr.Port)))
	callBuiltin(t, env, "__builtin_socket_close", handle)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("server did not accept connection")
	}
}

func TestServerListenAccept(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	addr := ln.Addr().(*net.TCPAddr)
	ln.Close()

	env := runtime.DefaultEnv()
	server := callBuiltin(t, env, "__builtin_socket_listen", value.Str("127.0.0.1"), value.Int(int64(addr.Port)))

	done := make(chan struct{})
	go func() {
		conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(addr.Port)))
		if err == nil {
			_ = conn.Close()
		}
		close(done)
	}()

	client := callBuiltin(t, env, "__builtin_socket_accept", server)
	callBuiltin(t, env, "__builtin_socket_close", client)
	callBuiltin(t, env, "__builtin_socket_close", server)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("client did not connect")
	}
}

func TestSocketReadWrite(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()
	addr := ln.Addr().(*net.TCPAddr)

	serverDone := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			close(serverDone)
			return
		}
		buf := make([]byte, 4)
		_, _ = conn.Read(buf)
		_, _ = conn.Write([]byte("pong"))
		_ = conn.Close()
		close(serverDone)
	}()

	env := runtime.DefaultEnv()
	handle := callBuiltin(t, env, "__builtin_socket_connect", value.Str("127.0.0.1"), value.Int(int64(addr.Port)))
	callBuiltin(t, env, "__builtin_socket_write", handle, value.Bytes([]byte("ping")))
	resp := callBuiltin(t, env, "__builtin_socket_read", handle, value.Int(4))
	callBuiltin(t, env, "__builtin_socket_close", handle)

	if resp.Kind != value.KindBytes || string(resp.Bytes) != "pong" {
		t.Fatalf("expected response 'pong', got %v", resp.String())
	}

	select {
	case <-serverDone:
	case <-time.After(2 * time.Second):
		t.Fatalf("server did not finish")
	}
}
