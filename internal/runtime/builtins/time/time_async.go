package time

import (
	"fmt"
	stdtime "time"

	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

func init() {
	registerAsyncSleep()
	registerAsyncWithTimeout()
}

func registerAsyncWithTimeout() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.AsyncWithTimeout,
			Name:         "__builtin_async_with_timeout",
			Arity:        2,
			ParamNames:   []string{"future", "nanos"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeAny}, {Kind: builtins.TypeInt}},
			Result:       builtins.TypeRef{Kind: builtins.TypeAny},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			return nil, fmt.Errorf("__builtin_async_with_timeout is handled by the VM directly")
		},
	})
}

func registerAsyncSleep() {
	builtins.Register(builtins.Builtin{
		Meta: builtins.Meta{
			ID:           builtins.AsyncTimeSleep,
			Name:         "__builtin_async_time_sleep",
			Arity:        1,
			ParamNames:   []string{"nanos"},
			Params:       []builtins.TypeRef{{Kind: builtins.TypeInt}},
			Result:       builtins.TypeRef{Kind: builtins.TypeVoid},
			ReceiverType: builtins.TypeVoid,
			MethodName:   "",
		},
		CallAsync: func(env builtins.Env, args []interface{}) (builtins.AsyncHandle, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("__builtin_async_time_sleep expects 1 argument, got %d", len(args))
			}
			nanosVal := args[0].(value.Value)
			if nanosVal.Kind != value.KindInt {
				return nil, fmt.Errorf("__builtin_async_time_sleep expects nanos as int")
			}
			if nanosVal.Int < 0 {
				return nil, fmt.Errorf("__builtin_async_time_sleep expects non-negative nanos, got %d", nanosVal.Int)
			}

			dur := stdtime.Duration(nanosVal.Int)
			return builtins.RunAsync(func() (interface{}, error) {
				stdtime.Sleep(dur)
				return value.Value{}, nil
			}), nil
		},
	})
}
