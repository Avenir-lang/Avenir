package runtime

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"avenir/internal/runtime/builtins"
	builtinsio "avenir/internal/runtime/builtins/io"
	"avenir/internal/value"
)

// ClosureCaller is a function that calls a closure with the given arguments.
type ClosureCaller func(clo *value.Closure, args []value.Value) (value.Value, error)

// Env aggregates host services used by builtins (IO, FS, HTTP, etc.).
// For now we only need IO; more services can be added later.
// Env implements builtins.Env to avoid import cycles.
type Env struct {
	ioService       builtinsio.IO
	closureCaller   ClosureCaller // Function to call closures (set by VM)
	structTypeNames []string
	netService      *netService
	fsService       *fsService
	httpService     *httpService
	execRoot        string
}

// IO returns the IO service. Implements builtins.Env interface.
func (e *Env) IO() builtins.IO {
	return e.ioService
}

// Net returns the networking service. Implements builtins.Env interface.
func (e *Env) Net() builtins.Net {
	return e.netService
}

// FS returns the filesystem service. Implements builtins.Env interface.
func (e *Env) FS() builtins.FS {
	return e.fsService
}

// HTTP returns the HTTP service. Implements builtins.Env interface.
func (e *Env) HTTP() builtins.HTTP {
	return e.httpService
}

// ExecRoot returns the execution root directory for relative file paths.
func (e *Env) ExecRoot() string {
	if e == nil {
		return ""
	}
	return e.execRoot
}

// StructTypeName returns the struct type name for the given index.
// Implements builtins.Env interface.
func (e *Env) StructTypeName(index int) (string, bool) {
	if e == nil || index < 0 || index >= len(e.structTypeNames) {
		return "", false
	}
	name := e.structTypeNames[index]
	if name == "" {
		return "", false
	}
	return name, true
}

// CallClosure calls a closure with the given arguments.
// Implements builtins.Env interface.
func (e *Env) CallClosure(clo interface{}, args []interface{}) (interface{}, error) {
	if e.closureCaller == nil {
		return value.Value{}, fmt.Errorf("closure caller not set in runtime.Env")
	}
	// Convert interface{} arguments to value.Value
	closure, ok := clo.(*value.Closure)
	if !ok {
		return value.Value{}, fmt.Errorf("CallClosure: expected *value.Closure, got %T", clo)
	}
	valueArgs := make([]value.Value, len(args))
	for i, arg := range args {
		val, ok := arg.(value.Value)
		if !ok {
			return value.Value{}, fmt.Errorf("CallClosure: argument %d is not value.Value, got %T", i, arg)
		}
		valueArgs[i] = val
	}
	result, err := e.closureCaller(closure, valueArgs)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SetClosureCaller sets the closure caller function.
// This is called by the VM to enable builtins to call closures.
func (e *Env) SetClosureCaller(caller ClosureCaller) {
	e.closureCaller = caller
}

// SetStructTypeNames sets the struct type name table for runtime lookups.
func (e *Env) SetStructTypeNames(names []string) {
	e.structTypeNames = names
}

// SetExecRoot sets the execution root directory for relative file paths.
func (e *Env) SetExecRoot(root string) {
	e.execRoot = root
}

// stdIO is the default IO implementation for CLI/console.
type stdIO struct {
	reader *bufio.Reader
}

func newStdIO() *stdIO {
	return &stdIO{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (s *stdIO) Println(str string) {
	fmt.Println(str)
}

func (s *stdIO) ReadLine() (string, error) {
	if s.reader == nil {
		s.reader = bufio.NewReader(os.Stdin)
	}
	line, err := s.reader.ReadString('\n')
	if err != nil {
		// Handle EOF - io.EOF is returned when stdin is closed
		if errors.Is(err, io.EOF) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	// Trim leading and trailing whitespace (including newlines)
	line = strings.TrimSpace(line)
	return line, nil
}

// DefaultEnv returns an Env with standard implementations
// (printing to stdout, real filesystem, etc.).
func DefaultEnv() *Env {
	return &Env{
		ioService:  newStdIO(),
		netService: newNetService(),
		fsService:  newFSService(),
		httpService: newHTTPService(),
	}
}

// NewEnv creates a new Env with the given IO service.
// This is useful for tests that need to provide a custom IO implementation.
func NewEnv(io builtinsio.IO) *Env {
	return &Env{
		ioService:  io,
		netService: newNetService(),
		fsService:  newFSService(),
		httpService: newHTTPService(),
	}
}
