package runtime

import (
	"fmt"

	"avenir/internal/runtime/builtins"
	// Import all builtin packages to trigger their init() functions for self-registration
	_ "avenir/internal/runtime/builtins/bytes"
	_ "avenir/internal/runtime/builtins/collections"
	_ "avenir/internal/runtime/builtins/dict"
	_ "avenir/internal/runtime/builtins/errors"
	_ "avenir/internal/runtime/builtins/fs"
	_ "avenir/internal/runtime/builtins/http"
	_ "avenir/internal/runtime/builtins/io"
	_ "avenir/internal/runtime/builtins/json"
	_ "avenir/internal/runtime/builtins/meta"
	_ "avenir/internal/runtime/builtins/net"
	_ "avenir/internal/runtime/builtins/strings"
	_ "avenir/internal/runtime/builtins/time"
	"avenir/internal/value"
)

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
