package runtime

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	// Import all builtin packages to trigger their init() functions for self-registration
	_ "avenir/internal/runtime/builtins/bytes"
	_ "avenir/internal/runtime/builtins/collections"
	_ "avenir/internal/runtime/builtins/crypto"
	_ "avenir/internal/runtime/builtins/dict"
	_ "avenir/internal/runtime/builtins/errors"
	_ "avenir/internal/runtime/builtins/fs"
	_ "avenir/internal/runtime/builtins/html"
	_ "avenir/internal/runtime/builtins/http"
	_ "avenir/internal/runtime/builtins/io"
	_ "avenir/internal/runtime/builtins/json"
	_ "avenir/internal/runtime/builtins/meta"
	_ "avenir/internal/runtime/builtins/net"
	_ "avenir/internal/runtime/builtins/sql"
	_ "avenir/internal/runtime/builtins/strings"
	_ "avenir/internal/runtime/builtins/time"
	_ "avenir/internal/runtime/builtins/tls"
	"avenir/internal/value"
)

func init() {
	builtins.AsyncRunner = func(fn func() (interface{}, error)) builtins.AsyncHandle {
		return RunAsync(func() (value.Value, error) {
			res, err := fn()
			if err != nil {
				return value.Value{}, err
			}
			val, ok := res.(value.Value)
			if !ok {
				return value.Value{}, fmt.Errorf("async builtin returned non-Value type")
			}
			return val, nil
		})
	}
}

// CallBuiltin executes a builtin identified by builtins.ID with given args.
// It uses services from Env (IO, FS, etc.).
// Returns the result value and an error if the call failed.
func CallBuiltin(env *Env, id builtins.ID, args []value.Value) (value.Value, bool, error) {
	builtin := builtins.LookupByID(id)
	if builtin == nil {
		return value.Value{}, false, fmt.Errorf("unknown builtin id %d", id)
	}

	// Convert []value.Value to []interface{} for the Call function
	argsIface := make([]interface{}, len(args))
	for i, arg := range args {
		argsIface[i] = arg
	}

	// Call the builtin
	resultIface, err := builtin.Call(env, argsIface)
	if err != nil {
		return value.Value{}, false, err
	}

	// Convert result back to value.Value
	result, ok := resultIface.(value.Value)
	if !ok {
		return value.Value{}, false, fmt.Errorf("builtin %s returned non-Value type", builtin.Meta.Name)
	}

	return result, true, nil
}

// CallBuiltinAsync executes an async builtin identified by builtins.ID.
// Returns an *AsyncHandle that will be resolved/rejected when the I/O completes.
func CallBuiltinAsync(env *Env, id builtins.ID, args []value.Value) (*AsyncHandle, error) {
	builtin := builtins.LookupByID(id)
	if builtin == nil {
		return nil, fmt.Errorf("unknown async builtin id %d", id)
	}
	if builtin.CallAsync == nil {
		return nil, fmt.Errorf("builtin %s is not async", builtin.Meta.Name)
	}

	argsIface := make([]interface{}, len(args))
	for i, arg := range args {
		argsIface[i] = arg
	}

	handleIface, err := builtin.CallAsync(env, argsIface)
	if err != nil {
		return nil, err
	}

	ah, ok := handleIface.(*AsyncHandle)
	if !ok {
		return nil, fmt.Errorf("async builtin %s returned non-AsyncHandle type", builtin.Meta.Name)
	}

	return ah, nil
}
