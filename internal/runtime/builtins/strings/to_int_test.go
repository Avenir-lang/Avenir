package strings

import (
	"testing"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func callToInt(t *testing.T, input string) (value.Value, error) {
	t.Helper()

	builtin := builtins.LookupByName("toInt")
	if builtin == nil {
		t.Fatalf("toInt builtin not registered")
	}
	result, err := builtin.Call(nil, []interface{}{value.Str(input)})
	if err != nil {
		return value.Value{}, err
	}
	val, ok := result.(value.Value)
	if !ok {
		t.Fatalf("toInt returned non-value type %T", result)
	}
	return val, nil
}

func TestToIntBuiltin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantKind value.Kind
		wantInt  int64
		wantErr  string
		wantFail bool
	}{
		{
			name:     "valid_positive",
			input:    "123",
			wantKind: value.KindInt,
			wantInt:  123,
		},
		{
			name:     "valid_negative",
			input:    "-42",
			wantKind: value.KindInt,
			wantInt:  -42,
		},
		{
			name:     "invalid_alpha",
			input:    "hello",
			wantErr:  `toInt: invalid integer "hello"`,
			wantFail: true,
		},
		{
			name:     "invalid_empty",
			input:    "",
			wantErr:  `toInt: invalid integer ""`,
			wantFail: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val, err := callToInt(t, tc.input)
			if tc.wantFail {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Error() != tc.wantErr {
					t.Fatalf("expected error %q, got %q", tc.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val.Kind != tc.wantKind {
				t.Fatalf("expected kind %v, got %v", tc.wantKind, val.Kind)
			}
			if val.Int != tc.wantInt {
				t.Fatalf("expected int %d, got %d", tc.wantInt, val.Int)
			}
		})
	}
}
