package http_test

import (
	"io"
	"net"
	nethttp "net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"avenir/internal/runtime"
	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func callBuiltin(t *testing.T, env *runtime.Env, name string, args ...value.Value) (value.Value, error) {
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
		return value.Value{}, err
	}
	val, ok := res.(value.Value)
	if !ok {
		t.Fatalf("builtin %q returned non-value %T", name, res)
	}
	return val, nil
}

func TestHTTPRequestBuiltin(t *testing.T) {
	server := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if r.Method != "POST" {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("X-Test") != "yes" {
			t.Fatalf("expected header X-Test=yes")
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "hello" {
			t.Fatalf("expected body hello, got %q", string(body))
		}
		w.Header().Set("X-Reply", "ok")
		w.WriteHeader(201)
		w.Write([]byte("done"))
	}))
	defer server.Close()

	env := runtime.DefaultEnv()
	headers := value.Dict(map[string]value.Value{"X-Test": value.Str("yes")})
	resp, err := callBuiltin(t, env, "__builtin_http_request",
		value.Str("POST"),
		value.Str(server.URL),
		headers,
		value.Bytes([]byte("hello")),
	)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	if resp.Kind != value.KindDict {
		t.Fatalf("expected dict response, got %v", resp.Kind)
	}
	if resp.Dict["status"].Kind != value.KindInt || resp.Dict["status"].Int != 201 {
		t.Fatalf("expected status 201, got %v", resp.Dict["status"].String())
	}
	if resp.Dict["body"].Kind != value.KindBytes || string(resp.Dict["body"].Bytes) != "done" {
		t.Fatalf("expected body done, got %v", resp.Dict["body"].String())
	}
	reply := resp.Dict["headers"]
	if reply.Kind != value.KindDict {
		t.Fatalf("expected headers dict, got %v", reply.Kind)
	}
	if reply.Dict["X-Reply"].Str != "ok" {
		t.Fatalf("expected X-Reply ok")
	}
}

func TestHTTPServerBuiltins(t *testing.T) {
	env := runtime.DefaultEnv()
	port := pickFreePort(t)

	serverHandle, err := callBuiltin(t, env, "__builtin_http_listen", value.Str("127.0.0.1"), value.Int(int64(port)))
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		req, err := callBuiltin(t, env, "__builtin_http_accept", serverHandle)
		if err != nil {
			t.Errorf("accept error: %v", err)
			return
		}
		if req.Kind != value.KindDict {
			t.Errorf("expected dict request, got %v", req.Kind)
			return
		}
		if req.Dict["path"].Str != "/ping" {
			t.Errorf("expected path /ping, got %q", req.Dict["path"].Str)
			return
		}
		headers := value.Dict(map[string]value.Value{
			"Content-Type": value.Str("text/plain"),
		})
		_, err = callBuiltin(t, env, "__builtin_http_respond",
			req.Dict["__handle"],
			value.Int(200),
			headers,
			value.Bytes([]byte("pong")),
		)
		if err != nil {
			t.Errorf("respond error: %v", err)
		}
	}()

	client := &nethttp.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/ping")
	if err != nil {
		t.Fatalf("client get error: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(body) != "pong" {
		t.Fatalf("expected pong, got %q", string(body))
	}
	<-done
}

func pickFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to pick free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

