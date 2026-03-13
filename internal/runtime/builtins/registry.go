package builtins

import (
	"fmt"
	"sync"
)

// Env provides host services to builtins (IO, FS, etc.).
// This interface is implemented by runtime.Env to avoid import cycles.
type Env interface {
	IO() IO
	StructTypeName(index int) (string, bool)
	Net() Net
	FS() FS
	HTTP() HTTP
	SQL() SQL
	ExecRoot() string
	// CallClosure calls a closure with the given arguments.
	// This enables builtins to call first-class functions (e.g., in map/filter/reduce).
	// The closure and arguments are passed as interface{} to avoid import cycles.
	CallClosure(clo interface{}, args []interface{}) (interface{}, error)
}

// IO is the minimal interface needed by builtin IO functions (e.g. print, input).
// This matches the interface defined in builtins/io/io.go.
type IO interface {
	Println(string)
	ReadLine() (string, error)
}

// Net is the minimal interface needed by builtin networking functions.
type Net interface {
	Connect(host string, port int) ([]byte, error)
	Listen(host string, port int) ([]byte, error)
	Accept(serverHandle []byte) ([]byte, error)
	Read(sockHandle []byte, n int) ([]byte, error)
	Write(sockHandle []byte, data []byte) (int, error)
	Close(handle []byte) error
}

// FS is the minimal interface needed by builtin filesystem functions.
type FS interface {
	Open(path string, mode string) ([]byte, error)
	Read(handle []byte, n int) ([]byte, error)
	Write(handle []byte, data []byte) (int, error)
	Close(handle []byte) error
	Exists(path string) (bool, error)
	Remove(path string) error
	Mkdir(path string) error
}

// HTTP is the minimal interface needed by builtin HTTP functions.
type HTTP interface {
	Request(method string, url string, headers map[string]string, body []byte) (*HTTPResponseData, error)
	Listen(host string, port int) ([]byte, error)
	Accept(serverHandle []byte) (*HTTPRequestData, error)
	Respond(reqHandle []byte, status int, headers map[string]string, body []byte) error
}

// SQL is the minimal interface needed by builtin SQL functions.
type SQL interface {
	PgConnect(host, port, user, password, database string) ([]byte, error)
	PgClose(handle []byte) error
	PgQuery(handle []byte, query string, params []interface{}) (*SQLResultData, error)
	PgExec(handle []byte, query string, params []interface{}) (*SQLResultData, error)
	PgBegin(handle []byte) ([]byte, error)
	PgCommit(txHandle []byte) error
	PgRollback(txHandle []byte) error
	SqliteConnect(path string) ([]byte, error)
	SqliteClose(handle []byte) error
	SqliteQuery(handle []byte, query string, params []interface{}) (*SQLResultData, error)
	SqliteExec(handle []byte, query string, params []interface{}) (*SQLResultData, error)
	SqliteBegin(handle []byte) ([]byte, error)
	SqliteCommit(txHandle []byte) error
	SqliteRollback(txHandle []byte) error
}

// SQLResultData represents the result of an SQL query or exec operation.
type SQLResultData struct {
	Columns      []string
	Rows         []map[string]interface{}
	RowsAffected int64
	LastInsertID int64
}

// HTTPRequestData represents a parsed HTTP request returned by the runtime service.
type HTTPRequestData struct {
	Handle     []byte
	Method     string
	Path       string
	RemoteAddr string
	Headers    map[string]string
	Body       []byte
}

// HTTPResponseData represents a response returned by the runtime service.
type HTTPResponseData struct {
	Status  int
	Headers map[string]string
	Body    []byte
}

// ID is a builtin function identifier.
type ID int

const (
	Print ID = iota
	Input
	ToInt
	TypeOf
	SocketConnect
	SocketListen
	SocketAccept
	SocketRead
	SocketWrite
	SocketClose
	FSOpen
	FSRead
	FSWrite
	FSClose
	FSExists
	FSRemove
	FSMkdir
	FSExecRoot
	HTTPRequest
	HTTPListen
	HTTPAccept
	HTTPRespond
	JSONParse
	JSONStringify
	Len
	Error
	ErrorMessage
	// Built-in methods
	ListAppend
	ListLength
	ListPop
	ListInsert
	ListRemoveAt
	ListClear
	ListIsEmpty
	ListGet
	ListContains
	ListIndexOf
	ListSlice
	ListReverse
	ListCopy
	ListMap
	ListFilter
	ListReduce
	StringLength
	StringToUpper
	StringToLower
	StringTrim
	StringTrimLeft
	StringTrimRight
	StringContains
	StringStartsWith
	StringEndsWith
	StringReplace
	StringSplit
	StringIndexOf
	StringLastIndexOf
	BytesLength
	BytesAppend
	BytesConcat
	BytesSlice
	BytesToString
	BytesFromString
	DictLength
	DictKeys
	DictValues
	DictHas
	DictGet
	DictSet
	DictRemove
	TimeNow
	TimeSleep
	TimeParseDateTime
	TimeFormatDateTime
	TimeParseDuration
	TimeYear
	TimeMonth
	TimeDay
	TimeHour
	TimeMinute
	TimeSecond
	CryptoSHA256
	CryptoSHA512
	CryptoHMACSHA256
	CryptoHMACSHA256Verify
	CryptoBase64URLEncode
	CryptoBase64URLDecode
	CryptoJWTSignHS256
	CryptoJWTVerifyHS256
	CryptoJWTSignRS256
	CryptoJWTVerifyRS256
	CryptoJWTSignES256
	CryptoJWTVerifyES256
	CryptoPasswordHash
	CryptoPasswordVerify

	// Async builtins
	AsyncTimeSleep
	AsyncFSOpen
	AsyncFSRead
	AsyncFSWrite
	AsyncFSClose
	AsyncFSExists
	AsyncFSRemove
	AsyncFSMkdir
	AsyncSocketConnect
	AsyncSocketAccept
	AsyncSocketRead
	AsyncSocketWrite
	AsyncSocketClose
	AsyncHTTPRequest
	AsyncHTTPAccept
	AsyncHTTPRespond
	AsyncFSReadAll
	AsyncWithTimeout

	// Async SQL builtins
	AsyncSQLPgConnect
	AsyncSQLPgClose
	AsyncSQLPgQuery
	AsyncSQLPgExec
	AsyncSQLPgBegin
	AsyncSQLPgCommit
	AsyncSQLPgRollback

	// Async SQLite builtins
	AsyncSQLSqliteConnect
	AsyncSQLSqliteClose
	AsyncSQLSqliteQuery
	AsyncSQLSqliteExec
	AsyncSQLSqliteBegin
	AsyncSQLSqliteCommit
	AsyncSQLSqliteRollback

	// HTML builder builtins
	HTMLNewBuilder
	HTMLTag
	HTMLVoidTag
	HTMLText
	HTMLRawHTML
	HTMLDoctype
	HTMLBuilderResult
	HTMLEscape
	HTMLRaw

	// HTML template engine builtins
	HTMLNewEngine
	HTMLEngineRender
	HTMLEngineCompile
	HTMLEngineSetDevMode
	HTMLTemplateRender
)

// TypeKind represents a type in the builtin type system.
type TypeKind int

const (
	TypeInt TypeKind = iota
	TypeFloat
	TypeString
	TypeBool
	TypeVoid
	TypeAny
	TypeList
	TypeDict
	TypeError
	TypeBytes
	TypeUnion
	// extend later if needed
)

// TypeRef describes a type in the builtin type system.
// This is independent from types.Type to avoid import cycles.
type TypeRef struct {
	Kind TypeKind
	Elem []TypeRef // for list<...> or future func types, generics, etc.
}

// Meta contains metadata about a builtin function or method.
// For regular built-in functions, ReceiverType is TypeVoid and MethodName is empty.
// For built-in methods, ReceiverType indicates the receiver type and MethodName is the method name.
type Meta struct {
	ID           ID
	Name         string
	Arity        int
	ParamNames   []string // Parameter names in order (must match Arity)
	Params       []TypeRef
	Result       TypeRef
	ReceiverType TypeKind // TypeVoid for regular functions, non-Void for methods
	MethodName   string   // Empty for regular functions, method name for methods
}

// AsyncHandle is an opaque interface for an asynchronous operation handle.
// The concrete type is *runtime.AsyncHandle; we use interface{} here to avoid import cycles.
type AsyncHandle = interface{}

// AsyncRunner is a function that starts an async operation and returns an AsyncHandle.
// It is set by the runtime package at init time to avoid import cycles.
var AsyncRunner func(fn func() (interface{}, error)) AsyncHandle

// RunAsync starts an async operation using the registered AsyncRunner.
// The function fn runs in a goroutine; its result is delivered to the returned AsyncHandle.
func RunAsync(fn func() (interface{}, error)) AsyncHandle {
	if AsyncRunner == nil {
		panic("builtins.AsyncRunner not registered; runtime must call builtins.SetAsyncRunner")
	}
	return AsyncRunner(fn)
}

// Builtin represents a complete builtin function or method with both metadata and implementation.
// The Call function signature uses interface{} to avoid import cycles.
// Implementations should import the value package and cast appropriately.
type Builtin struct {
	Meta Meta
	// Call executes the builtin synchronously. For methods, args[0] is the receiver.
	// For builtins that need host services, env provides IO, FS, etc.
	// For pure builtins, env may be nil.
	// Args and return value are []interface{} and interface{} to avoid import cycles.
	// Implementations should cast to []value.Value and value.Value.
	Call func(env Env, args []interface{}) (interface{}, error)

	// CallAsync executes the builtin asynchronously, returning an AsyncHandle
	// that will be resolved/rejected when the I/O completes.
	// If non-nil, this builtin is async and the compiler should emit OpCallBuiltinAsync.
	CallAsync func(env Env, args []interface{}) (AsyncHandle, error)
}

// IsAsyncBuiltin returns true if the given builtin ID is an async builtin.
func IsAsyncBuiltin(id ID) bool {
	b := LookupByID(id)
	if b == nil {
		return false
	}
	return b.CallAsync != nil
}

// registry holds all registered builtins with fast lookup indexes.
type registry struct {
	mu sync.RWMutex

	// Index by ID for fast dispatch
	byID map[ID]*Builtin

	// Index by name for regular function lookup
	byName map[string]*Builtin

	// Index by (receiver type, method name) for method lookup
	byMethod map[TypeKind]map[string]*Builtin
}

var globalRegistry = &registry{
	byID:     make(map[ID]*Builtin),
	byName:   make(map[string]*Builtin),
	byMethod: make(map[TypeKind]map[string]*Builtin),
}

// Register registers a builtin. This is called automatically by each builtin's init() function.
// Panics if the builtin ID is already registered or if metadata is invalid.
func Register(b Builtin) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	// Validate metadata
	if len(b.Meta.ParamNames) != b.Meta.Arity {
		panic(fmt.Sprintf("builtin %s (ID %d): ParamNames length (%d) != Arity (%d)",
			b.Meta.Name, b.Meta.ID, len(b.Meta.ParamNames), b.Meta.Arity))
	}

	// For methods, ensure the first parameter is the receiver
	if b.Meta.ReceiverType != TypeVoid {
		if len(b.Meta.Params) == 0 {
			panic(fmt.Sprintf("method %s on %v has no parameters (receiver missing)", b.Meta.MethodName, b.Meta.ReceiverType))
		}
		if b.Meta.Params[0].Kind != b.Meta.ReceiverType {
			panic(fmt.Sprintf("method %s on %v has wrong receiver type in first parameter", b.Meta.MethodName, b.Meta.ReceiverType))
		}
		if b.Meta.MethodName == "" {
			panic(fmt.Sprintf("builtin %s (ID %d): ReceiverType is set but MethodName is empty", b.Meta.Name, b.Meta.ID))
		}
	} else {
		if b.Meta.MethodName != "" {
			panic(fmt.Sprintf("builtin %s (ID %d): MethodName is set but ReceiverType is TypeVoid", b.Meta.Name, b.Meta.ID))
		}
	}

	// Check for duplicate ID
	if _, exists := globalRegistry.byID[b.Meta.ID]; exists {
		panic(fmt.Sprintf("builtin ID %d (%s) is already registered", b.Meta.ID, b.Meta.Name))
	}

	// Check for duplicate name (for regular functions)
	if b.Meta.ReceiverType == TypeVoid {
		if _, exists := globalRegistry.byName[b.Meta.Name]; exists {
			panic(fmt.Sprintf("builtin name %q is already registered", b.Meta.Name))
		}
		globalRegistry.byName[b.Meta.Name] = &b
	}

	// Index by ID
	globalRegistry.byID[b.Meta.ID] = &b

	// Index by method if applicable
	if b.Meta.ReceiverType != TypeVoid {
		if globalRegistry.byMethod[b.Meta.ReceiverType] == nil {
			globalRegistry.byMethod[b.Meta.ReceiverType] = make(map[string]*Builtin)
		}
		if _, exists := globalRegistry.byMethod[b.Meta.ReceiverType][b.Meta.MethodName]; exists {
			panic(fmt.Sprintf("method %s on %v is already registered", b.Meta.MethodName, b.Meta.ReceiverType))
		}
		globalRegistry.byMethod[b.Meta.ReceiverType][b.Meta.MethodName] = &b
	}
}

// LookupByID finds a builtin by ID. Returns nil if not found.
func LookupByID(id ID) *Builtin {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	return globalRegistry.byID[id]
}

// LookupByName finds a builtin by name (for regular functions only).
// Returns nil if not found.
func LookupByName(name string) *Builtin {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	return globalRegistry.byName[name]
}

// LookupMethod finds a built-in method by receiver type and method name.
// Returns nil if not found.
func LookupMethod(receiverType TypeKind, methodName string) *Builtin {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	if methodMap, ok := globalRegistry.byMethod[receiverType]; ok {
		return methodMap[methodName]
	}
	return nil
}

// All returns all registered builtin metadata. Used for type checking and other introspection.
func All() []Meta {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	result := make([]Meta, 0, len(globalRegistry.byID))
	for _, b := range globalRegistry.byID {
		result = append(result, b.Meta)
	}
	return result
}

// TypeKindFromString converts a type name string to a TypeKind.
// Returns the TypeKind and true if found, TypeVoid and false otherwise.
func TypeKindFromString(name string) (TypeKind, bool) {
	switch name {
	case "int":
		return TypeInt, true
	case "float":
		return TypeFloat, true
	case "string":
		return TypeString, true
	case "bool":
		return TypeBool, true
	case "void":
		return TypeVoid, true
	case "any":
		return TypeAny, true
	case "dict":
		return TypeDict, true
	case "error":
		return TypeError, true
	case "bytes":
		return TypeBytes, true
	default:
		return TypeVoid, false
	}
}
