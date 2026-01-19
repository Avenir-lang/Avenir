package ir

// OpCode is an opcode for Avenir VM bytecode
type OpCode byte

const (
	OpHalt OpCode = iota

	OpConst
	OpLoadLocal  // A = local variable index; push local[A]
	OpStoreLocal // A = local variable index; local[A] = top (no pop)
	OpPop

	// Math
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpNegate

	// Compare / logic
	OpEq
	OpNeq
	OpLt
	OpLte
	OpGt
	OpGte

	// Thread managment
	OpJump        // A = absolute ip
	OpJumpIfFalse // A = absolute ip, pop cond

	// Calls / returns
	OpCall        // A = function index, B = number of arguments
	OpCallValue   // A = number of arguments, callee = value on stack
	OpCallBuiltin // A = builtin id, B = number of arguments
	OpReturn      // B = 0 (without result) or 1 (with result returning)

	// Built-in operations
	OpMakeList     // A = number of elements; pop A values, create list, push list
	OpMakeDict     // A = number of entries; pop 2*A values (key, value), create dict, push dict
	OpIndex        // pop index, pop list, push element
	OpMakeSome     // pop value, wrap in some, push optional
	OpMakeStruct   // A = struct type index, B = number of fields; pop B values, create struct, push struct
	OpLoadField    // A = struct value (on stack), B = field index; pop struct, push field value
	OpStoreField   // A = struct value (on stack), B = field index; pop value, pop struct, push updated struct
	OpStringify    // pop value, push string representation
	OpConcatString // pop two strings, push concatenated string

	// Exceptions
	OpBeginTry // A = handler IP (absolute)
	OpEndTry   // remove handler for current frame
	OpThrow    // pop exception value and throw

	// Closures
	OpClosure      // A = function index, B = number of upvalues
	OpLoadUpvalue  // A = upvalue index
	OpStoreUpvalue // A = upvalue index
)

// Instruction is one bytecode instruction
// A and B are operands (semantics depend on Op)
type Instruction struct {
	Op OpCode
	A  int
	B  int
}

type ConstKind int

const (
	ConstInt ConstKind = iota
	ConstFloat
	ConstString
	ConstBool
	ConstBytes
	ConstNone
)

// Constant is written to the module's constant table
type Constant struct {
	Kind   ConstKind
	Int    int64
	Float  float64
	String string
	Bool   bool
	Bytes  []byte
}

// Chunk is a sequence of instructions plus a constant table
type Chunk struct {
	Code      []Instruction
	Consts    []Constant
	NumLocals int // Number of local slots, including parameters
}

// UpvalueInfo describes a captured variable (upvalue).
type UpvalueInfo struct {
	IsLocal bool // true: captures a local from immediately enclosing function
	Index   int  // slot index in enclosing function's locals OR "upvalue index" of enclosing function
}

// Function represents a single function in a module.
type Function struct {
	Name      string
	NumParams int
	Chunk     Chunk
	Upvalues  []UpvalueInfo // NEW: upvalues for closures
}

// Module represents a compiled Avenir program.
type Module struct {
	Functions   []*Function
	StructTypes []StructTypeInfo
	MainIndex   int // Index of the main function in the Functions array
}

// AddConstInt adds an integer constant and returns its index.
func (c *Chunk) AddConstInt(v int64) int {
	c.Consts = append(c.Consts, Constant{
		Kind: ConstInt,
		Int:  v,
	})
	return len(c.Consts) - 1
}

// AddConstFloat adds a float constant and returns its index.
func (c *Chunk) AddConstFloat(v float64) int {
	c.Consts = append(c.Consts, Constant{
		Kind:  ConstFloat,
		Float: v,
	})
	return len(c.Consts) - 1
}

// AddConstString adds a string constant and returns its index.
func (c *Chunk) AddConstString(s string) int {
	c.Consts = append(c.Consts, Constant{
		Kind:   ConstString,
		String: s,
	})
	return len(c.Consts) - 1
}

// AddConstBool adds a boolean constant and returns its index.
func (c *Chunk) AddConstBool(b bool) int {
	c.Consts = append(c.Consts, Constant{
		Kind: ConstBool,
		Bool: b,
	})
	return len(c.Consts) - 1
}

// AddConstBytes adds a bytes constant and returns its index.
func (c *Chunk) AddConstBytes(b []byte) int {
	c.Consts = append(c.Consts, Constant{
		Kind:  ConstBytes,
		Bytes: b,
	})
	return len(c.Consts) - 1
}

// AddConstNone adds a none constant and returns its index.
func (c *Chunk) AddConstNone() int {
	c.Consts = append(c.Consts, Constant{
		Kind: ConstNone,
	})
	return len(c.Consts) - 1
}

// Emit appends an instruction to the end of the chunk.
func (c *Chunk) Emit(op OpCode, a, b int) int {
	c.Code = append(c.Code, Instruction{
		Op: op,
		A:  a,
		B:  b,
	})
	return len(c.Code) - 1
}
