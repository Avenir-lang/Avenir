package time_test

import (
	"testing"
	stdtime "time"

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

func TestTimeNow(t *testing.T) {
	env := runtime.DefaultEnv()
	val, err := callBuiltin(t, env, "__builtin_time_now")
	if err != nil {
		t.Fatalf("now error: %v", err)
	}
	if val.Kind != value.KindInt {
		t.Fatalf("expected int, got %v", val.Kind)
	}
	if val.Int <= 0 {
		t.Fatalf("expected positive timestamp, got %d", val.Int)
	}
}

func TestTimeFormatParse(t *testing.T) {
	env := runtime.DefaultEnv()
	base := stdtime.Date(2024, 2, 3, 4, 5, 6, 0, stdtime.UTC)
	ts := base.UnixNano()
	layout := "2006-01-02 15:04:05"

	formatted, err := callBuiltin(t, env, "__builtin_time_format_datetime", value.Int(ts), value.Str(layout))
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	if formatted.Kind != value.KindString {
		t.Fatalf("expected string, got %v", formatted.Kind)
	}
	if formatted.Str != "2024-02-03 04:05:06" {
		t.Fatalf("unexpected formatted value: %q", formatted.Str)
	}

	parsed, err := callBuiltin(t, env, "__builtin_time_parse_datetime", value.Str(formatted.Str), value.Str(layout))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if parsed.Kind != value.KindInt {
		t.Fatalf("expected int, got %v", parsed.Kind)
	}
	if parsed.Int != ts {
		t.Fatalf("expected %d, got %d", ts, parsed.Int)
	}
}

func TestTimeParseDuration(t *testing.T) {
	env := runtime.DefaultEnv()
	val, err := callBuiltin(t, env, "__builtin_time_parse_duration", value.Str("1h30m"))
	if err != nil {
		t.Fatalf("parse duration error: %v", err)
	}
	want := int64(stdtime.Hour + 30*stdtime.Minute)
	if val.Kind != value.KindInt || val.Int != want {
		t.Fatalf("expected %d, got %v", want, val.String())
	}
}

func TestTimeComponents(t *testing.T) {
	env := runtime.DefaultEnv()
	base := stdtime.Date(2023, 12, 31, 23, 59, 1, 0, stdtime.UTC)
	ts := value.Int(base.UnixNano())

	year, _ := callBuiltin(t, env, "__builtin_time_year", ts)
	month, _ := callBuiltin(t, env, "__builtin_time_month", ts)
	day, _ := callBuiltin(t, env, "__builtin_time_day", ts)
	hour, _ := callBuiltin(t, env, "__builtin_time_hour", ts)
	minute, _ := callBuiltin(t, env, "__builtin_time_minute", ts)
	second, _ := callBuiltin(t, env, "__builtin_time_second", ts)

	if year.Int != 2023 || month.Int != 12 || day.Int != 31 {
		t.Fatalf("unexpected date parts: %d-%d-%d", year.Int, month.Int, day.Int)
	}
	if hour.Int != 23 || minute.Int != 59 || second.Int != 1 {
		t.Fatalf("unexpected time parts: %d:%d:%d", hour.Int, minute.Int, second.Int)
	}
}

func TestTimeSleepZero(t *testing.T) {
	env := runtime.DefaultEnv()
	_, err := callBuiltin(t, env, "__builtin_time_sleep", value.Int(0))
	if err != nil {
		t.Fatalf("sleep error: %v", err)
	}
}
