package time

import (
	"fmt"
	stdtime "time"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	registerNow()
	registerSleep()
	registerParseDateTime()
	registerFormatDateTime()
	registerParseDuration()
	registerYear()
	registerMonth()
	registerDay()
	registerHour()
	registerMinute()
	registerSecond()
}

func registerNow() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TimeNow,
			Name:       "__builtin_time_now",
			Arity:      0,
			ParamNames: []string{},
			Params:     []builtins.TypeRef{},
			Result:     builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 0 {
				return value.Value{}, fmt.Errorf("__builtin_time_now expects 0 arguments, got %d", len(args))
			}
			return value.Int(stdtime.Now().UTC().UnixNano()), nil
		},
	})
}

func registerSleep() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TimeSleep,
			Name:       "__builtin_time_sleep",
			Arity:      1,
			ParamNames: []string{"nanos"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_time_sleep expects 1 argument, got %d", len(args))
			}
			nanosVal := args[0].(value.Value)
			if nanosVal.Kind != value.KindInt {
				return value.Value{}, fmt.Errorf("__builtin_time_sleep expects nanos as int")
			}
			if nanosVal.Int < 0 {
				return value.Value{}, fmt.Errorf("__builtin_time_sleep expects non-negative nanos, got %d", nanosVal.Int)
			}
			stdtime.Sleep(stdtime.Duration(nanosVal.Int))
			return value.Value{}, nil
		},
	})
}

func registerParseDateTime() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TimeParseDateTime,
			Name:       "__builtin_time_parse_datetime",
			Arity:      2,
			ParamNames: []string{"text", "layout"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("__builtin_time_parse_datetime expects 2 arguments, got %d", len(args))
			}
			textVal := args[0].(value.Value)
			layoutVal := args[1].(value.Value)
			if textVal.Kind != value.KindString || layoutVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_time_parse_datetime expects text and layout as strings")
			}
			t, err := stdtime.Parse(layoutVal.Str, textVal.Str)
			if err != nil {
				return value.Value{}, err
			}
			return value.Int(t.UTC().UnixNano()), nil
		},
	})
}

func registerFormatDateTime() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TimeFormatDateTime,
			Name:       "__builtin_time_format_datetime",
			Arity:      2,
			ParamNames: []string{"timestamp", "layout"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeInt},
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeString},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return value.Value{}, fmt.Errorf("__builtin_time_format_datetime expects 2 arguments, got %d", len(args))
			}
			tsVal := args[0].(value.Value)
			layoutVal := args[1].(value.Value)
			if tsVal.Kind != value.KindInt || layoutVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_time_format_datetime expects timestamp int and layout string")
			}
			t := stdtime.Unix(0, tsVal.Int).UTC()
			return value.Str(t.Format(layoutVal.Str)), nil
		},
	})
}

func registerParseDuration() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TimeParseDuration,
			Name:       "__builtin_time_parse_duration",
			Arity:      1,
			ParamNames: []string{"text"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeString},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return value.Value{}, fmt.Errorf("__builtin_time_parse_duration expects 1 argument, got %d", len(args))
			}
			textVal := args[0].(value.Value)
			if textVal.Kind != value.KindString {
				return value.Value{}, fmt.Errorf("__builtin_time_parse_duration expects text as string")
			}
			d, err := stdtime.ParseDuration(textVal.Str)
			if err != nil {
				return value.Value{}, err
			}
			return value.Int(int64(d)), nil
		},
	})
}

func registerYear() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TimeYear,
			Name:       "__builtin_time_year",
			Arity:      1,
			ParamNames: []string{"timestamp"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			ts, err := requireTimestamp(args, "__builtin_time_year")
			if err != nil {
				return value.Value{}, err
			}
			return value.Int(int64(stdtime.Unix(0, ts).UTC().Year())), nil
		},
	})
}

func registerMonth() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TimeMonth,
			Name:       "__builtin_time_month",
			Arity:      1,
			ParamNames: []string{"timestamp"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			ts, err := requireTimestamp(args, "__builtin_time_month")
			if err != nil {
				return value.Value{}, err
			}
			return value.Int(int64(stdtime.Unix(0, ts).UTC().Month())), nil
		},
	})
}

func registerDay() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TimeDay,
			Name:       "__builtin_time_day",
			Arity:      1,
			ParamNames: []string{"timestamp"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			ts, err := requireTimestamp(args, "__builtin_time_day")
			if err != nil {
				return value.Value{}, err
			}
			return value.Int(int64(stdtime.Unix(0, ts).UTC().Day())), nil
		},
	})
}

func registerHour() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TimeHour,
			Name:       "__builtin_time_hour",
			Arity:      1,
			ParamNames: []string{"timestamp"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			ts, err := requireTimestamp(args, "__builtin_time_hour")
			if err != nil {
				return value.Value{}, err
			}
			return value.Int(int64(stdtime.Unix(0, ts).UTC().Hour())), nil
		},
	})
}

func registerMinute() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TimeMinute,
			Name:       "__builtin_time_minute",
			Arity:      1,
			ParamNames: []string{"timestamp"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			ts, err := requireTimestamp(args, "__builtin_time_minute")
			if err != nil {
				return value.Value{}, err
			}
			return value.Int(int64(stdtime.Unix(0, ts).UTC().Minute())), nil
		},
	})
}

func registerSecond() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:         builtins.TimeSecond,
			Name:       "__builtin_time_second",
			Arity:      1,
			ParamNames: []string{"timestamp"},
			Params: []builtins.TypeRef{
				{Kind: builtins.TypeInt},
			},
			Result:       builtins.TypeRef{Kind: builtins.TypeInt},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		Call: func(env builtins.Env, args []interface{}) (interface{}, error) {
			ts, err := requireTimestamp(args, "__builtin_time_second")
			if err != nil {
				return value.Value{}, err
			}
			return value.Int(int64(stdtime.Unix(0, ts).UTC().Second())), nil
		},
	})
}

func requireTimestamp(args []interface{}, name string) (int64, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("%s expects 1 argument, got %d", name, len(args))
	}
	val := args[0].(value.Value)
	if val.Kind != value.KindInt {
		return 0, fmt.Errorf("%s expects timestamp as int", name)
	}
	return val.Int, nil
}
