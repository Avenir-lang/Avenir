package json

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	registerParse()
	registerStringify()
}

func registerParse() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.JSONParse,
			Name:       "__builtin_json_parse",
			Arity:      1,
			ParamNames: []string{"text"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("json.parse expects 1 argument, got %d", len(args))
			}
			textVal := args[0].(value.Value)
			if textVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("json.parse expects a string")
			}
			parsed, err := parseJSON(textVal.Str)
			if err != nil {
				return value.Value{}, err
			}
			return parsed, nil
		},
	})
}

func registerStringify() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.JSONStringify,
			Name:       "__builtin_json_stringify",
			Arity:      1,
			ParamNames: []string{"value"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeAny},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("json.stringify expects 1 argument, got %d", len(args))
			}
			val := args[0].(value.Value)
			out, err := stringifyJSON(val)
			if err != nil {
				return value.Value{}, err
			}
			return value.Str(out), nil
		},
	})
}

func parseJSON(text string) (value.Value, error) {
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()

	var data interface{}
	if err := dec.Decode(&data); err != nil {
		return value.Value{}, fmt.Errorf("json.parse: %w", err)
	}
	// Ensure there is no trailing data.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return value.Value{}, fmt.Errorf("json.parse: extra data after JSON value")
		}
		return value.Value{}, fmt.Errorf("json.parse: %w", err)
	}

	return jsonToValue(data)
}

func jsonToValue(v interface{}) (value.Value, error) {
	switch val := v.(type) {
	case nil:
		return value.None(), nil
	case bool:
		return value.Bool(val), nil
	case string:
		return value.Str(val), nil
	case json.Number:
		return numberToValue(val)
	case float64:
		return value.Float(val), nil
	case []interface{}:
		items := make([]value.Value, len(val))
		for i, item := range val {
			converted, err := jsonToValue(item)
			if err != nil {
				return value.Value{}, err
			}
			items[i] = converted
		}
		return value.List(items), nil
	case map[string]interface{}:
		dict := make(map[string]value.Value, len(val))
		for k, item := range val {
			converted, err := jsonToValue(item)
			if err != nil {
				return value.Value{}, err
			}
			dict[k] = converted
		}
		return value.Dict(dict), nil
	default:
		return value.Value{}, fmt.Errorf("json.parse: unsupported JSON value %T", v)
	}
}

func numberToValue(num json.Number) (value.Value, error) {
	raw := num.String()
	if !strings.ContainsAny(raw, ".eE") {
		if i, err := num.Int64(); err == nil {
			return value.Int(i), nil
		}
	}
	f, err := num.Float64()
	if err != nil {
		return value.Value{}, fmt.Errorf("json.parse: invalid number %q", raw)
	}
	return value.Float(f), nil
}

func stringifyJSON(val value.Value) (string, error) {
	var b strings.Builder
	if err := writeJSONValue(&b, val); err != nil {
		return "", err
	}
	return b.String(), nil
}

func writeJSONValue(b *strings.Builder, val value.Value) error {
	switch val.Kind {
	case value.KindInt:
		b.WriteString(strconv.FormatInt(val.Int, 10))
		return nil
	case value.KindFloat:
		if math.IsNaN(val.Float) || math.IsInf(val.Float, 0) {
			return fmt.Errorf("json.stringify: cannot encode non-finite float")
		}
		b.WriteString(strconv.FormatFloat(val.Float, 'g', -1, 64))
		return nil
	case value.KindString:
		return writeJSONString(b, val.Str)
	case value.KindBool:
		if val.Bool {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
		return nil
	case value.KindOptional:
		if val.Optional == nil || !val.Optional.IsSome {
			b.WriteString("null")
			return nil
		}
		return writeJSONValue(b, val.Optional.Value)
	case value.KindList:
		b.WriteByte('[')
		for i, item := range val.List {
			if i > 0 {
				b.WriteByte(',')
			}
			if err := writeJSONValue(b, item); err != nil {
				return err
			}
		}
		b.WriteByte(']')
		return nil
	case value.KindDict:
		b.WriteByte('{')
		if len(val.Dict) > 0 {
			keys := make([]string, 0, len(val.Dict))
			for k := range val.Dict {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for i, k := range keys {
				if i > 0 {
					b.WriteByte(',')
				}
				if err := writeJSONString(b, k); err != nil {
					return err
				}
				b.WriteByte(':')
				if err := writeJSONValue(b, val.Dict[k]); err != nil {
					return err
				}
			}
		}
		b.WriteByte('}')
		return nil
	default:
		return fmt.Errorf("json.stringify: unsupported value type %v", val.Kind)
	}
}

func writeJSONString(b *strings.Builder, s string) error {
	encoded, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("json.stringify: invalid string: %w", err)
	}
	b.Write(encoded)
	return nil
}
