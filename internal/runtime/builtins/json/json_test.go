package json_test

import (
	"testing"

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

func TestJSONParseObject(t *testing.T) {
	env := runtime.DefaultEnv()
	src := `{"name":"Alex","age":30,"tags":["dev","go"],"active":true,"meta":null}`
	val, err := callBuiltin(t, env, "__builtin_json_parse", value.Str(src))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if val.Kind != value.KindDict {
		t.Fatalf("expected dict, got %v", val.Kind)
	}
	if got := val.Dict["name"]; got.Kind != value.KindString || got.Str != "Alex" {
		t.Fatalf("expected name=Alex, got %v", got.String())
	}
	if got := val.Dict["age"]; got.Kind != value.KindInt || got.Int != 30 {
		t.Fatalf("expected age=30, got %v", got.String())
	}
	tags := val.Dict["tags"]
	if tags.Kind != value.KindList || len(tags.List) != 2 {
		t.Fatalf("expected tags list, got %v", tags.String())
	}
	if meta := val.Dict["meta"]; meta.Kind != value.KindOptional || meta.Optional == nil || meta.Optional.IsSome {
		t.Fatalf("expected meta=null, got %v", meta.String())
	}
}

func TestJSONParseInvalid(t *testing.T) {
	env := runtime.DefaultEnv()
	_, err := callBuiltin(t, env, "__builtin_json_parse", value.Str(`{"a":`))
	if err == nil {
		t.Fatalf("expected parse error, got nil")
	}
}

func TestJSONStringifyBasics(t *testing.T) {
	env := runtime.DefaultEnv()
	input := value.Dict(map[string]value.Value{
		"b": value.Int(2),
		"a": value.Str("x"),
	})
	val, err := callBuiltin(t, env, "__builtin_json_stringify", input)
	if err != nil {
		t.Fatalf("stringify error: %v", err)
	}
	if val.Kind != value.KindString {
		t.Fatalf("expected string result, got %v", val.Kind)
	}
	expected := `{"a":"x","b":2}`
	if val.Str != expected {
		t.Fatalf("expected %q, got %q", expected, val.Str)
	}
}

func TestJSONStringifyUnsupported(t *testing.T) {
	env := runtime.DefaultEnv()
	_, err := callBuiltin(t, env, "__builtin_json_stringify", value.Bytes([]byte{1, 2}))
	if err == nil {
		t.Fatalf("expected stringify error for unsupported type, got nil")
	}
}
