package types

import (
	"fmt"

	"avenir/internal/ast"
)

type Type interface {
	String() string
	equal(Type) bool
}

// Basic types

type BasicKind int

const (
	BasicInvalid BasicKind = iota
	BasicInt
	BasicFloat
	BasicString
	BasicBool
	BasicVoid
	BasicAny
	BasicError
	BasicBytes
)

type Basic struct {
	Kind BasicKind
	Name string
}

func (b *Basic) String() string { return b.Name }

func (b *Basic) equal(other Type) bool {
	o, ok := other.(*Basic)
	if !ok {
		return false
	}
	return b.Kind == o.Kind
}

var (
	Invalid   = &Basic{Kind: BasicInvalid, Name: "invalid"}
	Int       = &Basic{Kind: BasicInt, Name: "int"}
	Float     = &Basic{Kind: BasicFloat, Name: "float"}
	String    = &Basic{Kind: BasicString, Name: "string"}
	Bool      = &Basic{Kind: BasicBool, Name: "bool"}
	Void      = &Basic{Kind: BasicVoid, Name: "void"}
	Any       = &Basic{Kind: BasicAny, Name: "any"}
	ErrorType = &Basic{Kind: BasicError, Name: "error"}
	Bytes     = &Basic{Kind: BasicBytes, Name: "bytes"}
)

func IsInvalid(t Type) bool {
	if b, ok := t.(*Basic); ok {
		return b.Kind == BasicInvalid
	}

	return false
}

func IsVoid(t Type) bool {
	if b, ok := t.(*Basic); ok {
		return b.Kind == BasicVoid
	}

	return false
}

// Lists Element Types - a list of valid element types
// list<int> => ElementTypes = [int]
// list<string, int> => [string, int]

type List struct {
	ElementTypes []Type
}

func (l *List) String() string {
	if len(l.ElementTypes) == 0 {
		return "list<?>"
	}
	s := "list<"
	for i, et := range l.ElementTypes {
		if i > 0 {
			s += ", "
		}
		s += et.String()
	}
	s += ">"
	return s
}

func (l *List) equal(other Type) bool {
	o, ok := other.(*List)
	if !ok {
		return false
	}

	if len(l.ElementTypes) != len(o.ElementTypes) {
		return false
	}
	for i, t := range l.ElementTypes {
		if !t.equal(o.ElementTypes[i]) {
			return false
		}
	}
	return true
}

// Dict represents a dictionary type: dict<T>
type Dict struct {
	ValueType Type
}

func (d *Dict) String() string {
	if d.ValueType == nil {
		return "dict<?>"
	}
	return "dict<" + d.ValueType.String() + ">"
}

func (d *Dict) equal(other Type) bool {
	o, ok := other.(*Dict)
	if !ok {
		return false
	}
	if d.ValueType == nil || o.ValueType == nil {
		return d.ValueType == o.ValueType
	}
	return d.ValueType.equal(o.ValueType)
}

// Functions

// Func - function type: (T1, T2, ...) -> R
type Func struct {
	ParamTypes []Type
	Result     Type
}

func (f *Func) String() string {
	s := "func("
	for i, pt := range f.ParamTypes {
		if i > 0 {
			s += ", "
		}
		s += pt.String()
	}
	s += ") | "
	if f.Result != nil {
		s += f.Result.String()
	} else {
		s += "void"
	}
	return s
}

func (f *Func) equal(other Type) bool {
	o, ok := other.(*Func)
	if !ok {
		return false
	}
	if len(f.ParamTypes) != len(o.ParamTypes) {
		return false
	}
	for i, pt := range f.ParamTypes {
		if !pt.equal(o.ParamTypes[i]) {
			return false
		}
	}
	if f.Result == nil || o.Result == nil {
		return f.Result == o.Result
	}
	return f.Result.equal(o.Result)
}

// Union is a union type: <T1|T2|...>.
type Union struct {
	Variants []Type
}

func (u *Union) String() string {
	if len(u.Variants) == 0 {
		return "<>"
	}
	s := "<"
	for i, v := range u.Variants {
		if i > 0 {
			s += "|"
		}
		s += v.String()
	}
	s += ">"
	return s
}

func (u *Union) equal(other Type) bool {
	o, ok := other.(*Union)
	if !ok {
		return false
	}
	// Order-insensitive equality: check if both unions contain the same set of variants
	if len(u.Variants) != len(o.Variants) {
		return false
	}
	// For each variant in u, check if it exists in o
	for _, v := range u.Variants {
		found := false
		for _, ov := range o.Variants {
			if v.equal(ov) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Optional represents an optional type: T?
type Optional struct {
	Inner Type
}

func (o *Optional) String() string {
	return o.Inner.String() + "?"
}

func (o *Optional) equal(other Type) bool {
	otherOpt, ok := other.(*Optional)
	if !ok {
		return false
	}
	return o.Inner.equal(otherOpt.Inner)
}

// Struct represents a struct type with named fields.
// Structs use nominal typing: two structs are equal only if they have the same name.
type Struct struct {
	Name            string
	Fields          []Field
	IsPublic        bool               // true if struct is public, false if private
	IsMutable       bool               // true if struct is mutable (mut struct), false if immutable
	InstanceMethods map[string]*Method // instance method name -> method signature
	StaticMethods   map[string]*Method // static method name -> method signature
}

// Method represents a method signature on a type.
// For instance methods: Receiver is the receiver type, and it's the first parameter.
// For static methods: Receiver is the struct type (for identification), but it's NOT a parameter.
type Method struct {
	Name       string
	Receiver   Type
	ParamTypes []Type
	Result     Type
	IsStatic   bool // true for static methods, false for instance methods
	IsPublic   bool // true if method is public (pub fun), false if private
}

func (s *Struct) String() string {
	return s.Name
}

func (s *Struct) equal(other Type) bool {
	otherStruct, ok := other.(*Struct)
	if !ok {
		return false
	}
	// Nominal typing: same name required
	if s.Name != otherStruct.Name {
		return false
	}
	// Also check field count matches (defensive)
	if len(s.Fields) != len(otherStruct.Fields) {
		return false
	}
	return true
}

// Field represents a struct field with a name, type, visibility, mutability, and optional default value.
type Field struct {
	Name         string
	Type         Type
	IsPublic     bool        // true if field is public, false if private
	IsMutable    bool        // true if field is explicitly mutable (mut field), overrides struct default
	DefaultExpr  ast.Expr    // nil if no default, non-nil if default is provided (compile-time constant)
	DefaultValue interface{} // materialized default value (for IR compiler)
}

// Interface represents an interface type with method signatures.
// Interfaces use structural typing: a type satisfies an interface if it has all required methods.
type Interface struct {
	Name           string
	Methods        []InterfaceMethod
	IsPublic       bool
	DefiningModule string // module where interface is defined (for visibility checks)
}

// InterfaceMethod represents a method signature required by an interface.
type InterfaceMethod struct {
	Name       string
	ParamTypes []Type
	Return     Type
}

func (i *Interface) String() string {
	return i.Name
}

func (i *Interface) equal(other Type) bool {
	otherInterface, ok := other.(*Interface)
	if !ok {
		return false
	}
	// Nominal typing: same name required
	if i.Name != otherInterface.Name {
		return false
	}
	// Also check method count matches (defensive)
	if len(i.Methods) != len(otherInterface.Methods) {
		return false
	}
	return true
}

// Equal - public function for type checking
func Equal(a, b Type) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.equal(b)
}

// Debug helper
func DebugType(t Type) string {
	if t == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%T(%s)", t, t.String())
}
