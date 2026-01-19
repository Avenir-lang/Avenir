package token

import "fmt"

type Kind int

const (
	Illegal Kind = iota
	EOF

	Ident       // Identifier
	Int         // Integer
	Float       // Floating-point number
	String      // String literal
	StringPart  // String literal segment (for interpolation)
	InterpStart // ${
	InterpEnd   // }
	StringEnd   // end of interpolated string
	Bytes       // Bytes literal (b"...")

	// Keywords
	Pckg
	Fun
	Var
	If
	Else
	Return
	While
	For
	In
	Try
	Catch
	Throw
	Break
	True
	False
	None
	Some
	Struct    // struct
	Import    // import
	Pub       // pub
	Mut       // mut
	Interface // interface

	// Type keywords
	IntType    // int
	FloatType  // float
	StringType // string
	BoolType   // bool
	VoidType   // void
	AnyType    // any
	ListType   // list
	ErrorType  // error
	BytesType  // bytes
	DictType   // dict

	// Operators
	Assign // =

	Plus    // +
	Minus   // -
	Star    // *
	Slash   // /
	Percent // %

	Bang   // !
	AndAnd // &&
	OrOr   // ||

	Eq    // ==
	NotEq // !=
	Lt    // <
	LtEq  // <=
	Gt    // >
	GtEq  // >=

	// Symbols
	Pipe      // |
	Comma     // ,
	Semicolon // ;
	Dot       // .
	Colon     // :

	LParen   // (
	RParen   // )
	LBrace   // {
	RBrace   // }
	LBracket // [
	RBracket // ]
	Question // ?
)

type Position struct {
	Line   int
	Column int
}

type Token struct {
	Kind   Kind
	Lexeme string
	Pos    Position
}

func (k Kind) String() string {
	switch k {
	case Illegal:
		return "Illegal"
	case EOF:
		return "EOF"
	case Ident:
		return "Ident"
	case Int:
		return "Int"
	case Float:
		return "Float"
	case String:
		return "String"
	case StringPart:
		return "StringPart"
	case InterpStart:
		return "InterpStart"
	case InterpEnd:
		return "InterpEnd"
	case StringEnd:
		return "StringEnd"
	case Bytes:
		return "Bytes"
	case Pckg:
		return "Pckg"
	case Fun:
		return "Fun"
	case Var:
		return "Var"
	case If:
		return "If"
	case Else:
		return "Else"
	case Return:
		return "Return"
	case While:
		return "While"
	case For:
		return "For"
	case In:
		return "In"
	case Try:
		return "Try"
	case Catch:
		return "Catch"
	case Throw:
		return "Throw"
	case Break:
		return "Break"
	case True:
		return "True"
	case False:
		return "False"
	case None:
		return "None"
	case Some:
		return "Some"
	case Struct:
		return "Struct"
	case Import:
		return "Import"
	case Pub:
		return "Pub"
	case Mut:
		return "Mut"
	case Interface:
		return "Interface"
	case IntType:
		return "IntType"
	case FloatType:
		return "FloatType"
	case StringType:
		return "StringType"
	case BoolType:
		return "BoolType"
	case VoidType:
		return "VoidType"
	case AnyType:
		return "AnyType"
	case ListType:
		return "ListType"
	case ErrorType:
		return "ErrorType"
	case BytesType:
		return "BytesType"
	case DictType:
		return "DictType"
	case Assign:
		return "Assign"
	case Plus:
		return "Plus"
	case Minus:
		return "Minus"
	case Star:
		return "Star"
	case Slash:
		return "Slash"
	case Percent:
		return "Percent"
	case Bang:
		return "Bang"
	case AndAnd:
		return "AndAnd"
	case OrOr:
		return "OrOr"
	case Eq:
		return "Eq"
	case NotEq:
		return "NotEq"
	case Lt:
		return "Lt"
	case LtEq:
		return "LtEq"
	case Gt:
		return "Gt"
	case GtEq:
		return "GtEq"
	case Pipe:
		return "Pipe"
	case Comma:
		return "Comma"
	case Semicolon:
		return "Semicolon"
	case LParen:
		return "LParen"
	case RParen:
		return "RParen"
	case LBrace:
		return "LBrace"
	case RBrace:
		return "RBrace"
	case LBracket:
		return "LBracket"
	case RBracket:
		return "RBracket"
	case Question:
		return "Question"
	case Dot:
		return "Dot"
	case Colon:
		return "Colon"
	default:
		return fmt.Sprintf("Kind(%d)", int(k))
	}
}

var keywords = map[string]Kind{
	"pckg":      Pckg,
	"fun":       Fun,
	"var":       Var,
	"if":        If,
	"else":      Else,
	"return":    Return,
	"while":     While,
	"for":       For,
	"in":        In,
	"try":       Try,
	"catch":     Catch,
	"throw":     Throw,
	"break":     Break,
	"true":      True,
	"false":     False,
	"none":      None,
	"some":      Some,
	"struct":    Struct,
	"import":    Import,
	"pub":       Pub,
	"mut":       Mut,
	"interface": Interface,

	"int":    IntType,
	"float":  FloatType,
	"string": StringType,
	"bool":   BoolType,
	"void":   VoidType,
	"any":    AnyType,
	"list":   ListType,
	"error":  ErrorType,
	"bytes":  BytesType,
	"dict":   DictType,
}

func LookupIdent(lit string) Kind {
	if kind, ok := keywords[lit]; ok {
		return kind
	}
	return Ident
}
