package fs_test

import (
	"path/filepath"
	"testing"

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

func TestFSOpenWriteReadClose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hello.txt")
	env := runtime.DefaultEnv()

	handle := callBuiltin(t, env, "__builtin_fs_open", value.Str(path), value.Str("w"))
	callBuiltin(t, env, "__builtin_fs_write", handle, value.Bytes([]byte("hello")))
	callBuiltin(t, env, "__builtin_fs_close", handle)

	readHandle := callBuiltin(t, env, "__builtin_fs_open", value.Str(path), value.Str("r"))
	data := callBuiltin(t, env, "__builtin_fs_read", readHandle, value.Int(5))
	callBuiltin(t, env, "__builtin_fs_close", readHandle)

	if data.Kind != value.KindBytes || string(data.Bytes) != "hello" {
		t.Fatalf("expected %q, got %v", "hello", data.String())
	}
}

func TestFSExistsMkdirRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir")
	env := runtime.DefaultEnv()

	exists := callBuiltin(t, env, "__builtin_fs_exists", value.Str(path))
	if exists.Kind != value.KindBool || exists.Bool {
		t.Fatalf("expected exists=false, got %v", exists.String())
	}

	callBuiltin(t, env, "__builtin_fs_mkdir", value.Str(path))
	exists = callBuiltin(t, env, "__builtin_fs_exists", value.Str(path))
	if exists.Kind != value.KindBool || !exists.Bool {
		t.Fatalf("expected exists=true, got %v", exists.String())
	}

	callBuiltin(t, env, "__builtin_fs_remove", value.Str(path))
	exists = callBuiltin(t, env, "__builtin_fs_exists", value.Str(path))
	if exists.Kind != value.KindBool || exists.Bool {
		t.Fatalf("expected exists=false after remove, got %v", exists.String())
	}
}
