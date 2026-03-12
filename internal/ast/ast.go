package ast

import "avenir/internal/token"

// Basic interfaces

type Node interface {
	Pos() token.Position
}

type Stmt interface {
	Node
	stmtNode()
}

type Expr interface {
	Node
	exprNode()
}

type TypeNode interface {
	Node
	typeNode()
}

// Program / Package

type Program struct {
	Package    *PackageDecl
	Imports    []*ImportDecl
	Vars       []*VarDeclStmt
	Funcs      []*FunDecl
	Structs    []*StructDecl
	Interfaces []*InterfaceDecl
}

func (p *Program) Pos() token.Position {
	if p.Package != nil {
		return p.Package.Pos()
	}
	if len(p.Funcs) > 0 {
		return p.Funcs[0].Pos()
	}
	return token.Position{}
}

type PackageDecl struct {
	Name    string
	NamePos token.Position
}

func (p *PackageDecl) Pos() token.Position { return p.NamePos }

type ImportDecl struct {
	ImportPos token.Position
	Path      []string // e.g. ["std", "io"]
	Alias     string   // may be "", meaning "use last segment"
}

func (d *ImportDecl) Pos() token.Position { return d.ImportPos }

// Decorator represents a decorator annotation on a function declaration.
type Decorator struct {
	AtPos token.Position
	Expr  Expr
}

func (d *Decorator) Pos() token.Position { return d.AtPos }

// FunDecl / Param

type TypeParam struct {
	Name       string
	NamePos    token.Position
	IsVariadic bool // true for variadic type pack: ...Args
}

func (tp *TypeParam) Pos() token.Position { return tp.NamePos }

// VariadicParam represents a variadic value parameter: args... | Args...
type VariadicParam struct {
	Name    string
	NamePos token.Position
	Type    TypeNode // TypePackExpansion node
}

func (vp *VariadicParam) Pos() token.Position { return vp.NamePos }

type FunDecl struct {
	Decorators    []*Decorator // decorator annotations, e.g. @log, @cache(60)
	Name          string
	NamePos       token.Position
	TypeParams    []*TypeParam // generic type parameters, e.g. <T, U>
	Receiver      *Receiver    // nil for regular functions, non-nil for methods
	Params        []*Param
	VariadicParam *VariadicParam // nil or variadic param (args... | Args...)
	Return        TypeNode
	Throws        []TypeNode // error types after "!", e.g. fun f() | int ! MyError | OtherError
	Body          *BlockStmt
	IsPublic      bool // true if declared with "pub fun"
	IsAsync       bool // true if declared with "async fun"
}

// ReceiverKind distinguishes between instance and static methods.
type ReceiverKind int

const (
	ReceiverInstance ReceiverKind = iota // Instance method: (name | Type).method
	ReceiverStatic                       // Static method: Type.method
)

// Receiver represents a method receiver.
// For instance methods: (name | Type) - Name is the receiver variable name.
// For static methods: Type - Name is empty, only Type is set.
type Receiver struct {
	Kind    ReceiverKind // Instance or Static
	Name    string       // Empty for static methods
	NamePos token.Position
	Type    TypeNode
}

func (f *FunDecl) Pos() token.Position { return f.NamePos }

type Param struct {
	Name    string
	NamePos token.Position
	Type    TypeNode
	Default Expr // nil if no default value
}

func (p *Param) Pos() token.Position { return p.NamePos }

// ---------- Structs ----------

type StructDecl struct {
	Name       string
	NamePos    token.Position
	TypeParams []*TypeParam // generic type parameters, e.g. <T, U>
	Fields     []*FieldDecl
	IsPublic   bool
	IsMutable  bool // true if declared with "mut struct"
}

func (s *StructDecl) Pos() token.Position { return s.NamePos }

type FieldDecl struct {
	Name        string
	NamePos     token.Position
	Type        TypeNode
	IsPublic    bool // true if declared with "pub"
	IsMutable   bool // true if declared with "mut" (overrides struct default)
	DefaultExpr Expr // nil if no default value, non-nil if default is provided
}

func (f *FieldDecl) Pos() token.Position { return f.NamePos }

// ---------- Interfaces ----------

type InterfaceDecl struct {
	Name     string
	NamePos  token.Position
	Methods  []*InterfaceMethod
	IsPublic bool
}

func (i *InterfaceDecl) Pos() token.Position { return i.NamePos }

type InterfaceMethod struct {
	Name       string
	NamePos    token.Position
	ParamTypes []TypeNode
	Return     TypeNode
}

func (m *InterfaceMethod) Pos() token.Position { return m.NamePos }

// ---------- Types ----------

type SimpleType struct {
	Name    string
	NamePos token.Position
}

func (t *SimpleType) Pos() token.Position { return t.NamePos }
func (t *SimpleType) typeNode()           {}

// QualifiedType represents a type name with a dotted path (e.g. net.Socket).
type QualifiedType struct {
	Path    []string
	PathPos token.Position
}

func (t *QualifiedType) Pos() token.Position { return t.PathPos }
func (t *QualifiedType) typeNode()           {}

type ListType struct {
	ListPos      token.Position
	ElementTypes []TypeNode
}

func (t *ListType) Pos() token.Position { return t.ListPos }
func (t *ListType) typeNode()           {}

type DictType struct {
	DictPos   token.Position
	KeyType   TypeNode // nil means string (backward compat dict<V>)
	ValueType TypeNode
}

func (t *DictType) Pos() token.Position { return t.DictPos }
func (t *DictType) typeNode()           {}

type FuncType struct {
	FunPos     token.Position
	ParamTypes []TypeNode
	Result     TypeNode
}

func (t *FuncType) Pos() token.Position { return t.FunPos }
func (t *FuncType) typeNode()           {}

type UnionType struct {
	UnionPos token.Position
	Variants []TypeNode
}

func (t *UnionType) Pos() token.Position { return t.UnionPos }
func (t *UnionType) typeNode()           {}

type OptionalType struct {
	Inner    TypeNode
	QMarkPos token.Position
}

func (t *OptionalType) Pos() token.Position { return t.Inner.Pos() }
func (t *OptionalType) typeNode()           {}

type GenericInstanceType struct {
	Name     string
	NamePos  token.Position
	TypeArgs []TypeNode
}

func (t *GenericInstanceType) Pos() token.Position { return t.NamePos }
func (t *GenericInstanceType) typeNode()           {}

// TypePackExpansion represents a variadic type pack expansion in type position: Args...
type TypePackExpansion struct {
	Name    string
	NamePos token.Position
}

func (t *TypePackExpansion) Pos() token.Position { return t.NamePos }
func (t *TypePackExpansion) typeNode()           {}

// ---------- Statements ----------

type BlockStmt struct {
	LBrace token.Position
	Stmts  []Stmt
	RBrace token.Position
}

func (b *BlockStmt) Pos() token.Position { return b.LBrace }
func (b *BlockStmt) stmtNode()           {}

type VarDeclStmt struct {
	VarPos  token.Position
	Name    string
	NamePos token.Position
	Type    TypeNode
	Value   Expr
}

func (s *VarDeclStmt) Pos() token.Position { return s.VarPos }
func (s *VarDeclStmt) stmtNode()           {}

type AssignStmt struct {
	Name    string
	NamePos token.Position
	Value   Expr
}

func (s *AssignStmt) Pos() token.Position { return s.NamePos }
func (s *AssignStmt) stmtNode()           {}

type StructFieldAssignStmt struct {
	Struct    Expr   // struct variable expression (e.g., `p`)
	Field     string // field name (e.g., `x`)
	FieldPos  token.Position
	Value     Expr // value to assign
	AssignPos token.Position
}

func (s *StructFieldAssignStmt) Pos() token.Position { return s.AssignPos }
func (s *StructFieldAssignStmt) stmtNode()           {}

type ExprStmt struct {
	Expression Expr
}

func (s *ExprStmt) Pos() token.Position { return s.Expression.Pos() }
func (s *ExprStmt) stmtNode()           {}

type IfStmt struct {
	IfPos token.Position
	Cond  Expr
	Then  *BlockStmt
	Else  Stmt // either *BlockStmt or *IfStmt (else-if)
}

func (s *IfStmt) Pos() token.Position { return s.IfPos }
func (s *IfStmt) stmtNode()           {}

type ReturnStmt struct {
	ReturnPos token.Position
	Result    Expr // may be nil for `return;`
}

func (s *ReturnStmt) Pos() token.Position { return s.ReturnPos }
func (s *ReturnStmt) stmtNode()           {}

type ThrowStmt struct {
	ThrowPos token.Position
	Expr     Expr
}

func (s *ThrowStmt) Pos() token.Position { return s.ThrowPos }
func (s *ThrowStmt) stmtNode()           {}

type BreakStmt struct {
	BreakPos token.Position
}

func (s *BreakStmt) Pos() token.Position { return s.BreakPos }
func (s *BreakStmt) stmtNode()           {}

type ContinueStmt struct {
	ContinuePos token.Position
}

func (s *ContinueStmt) Pos() token.Position { return s.ContinuePos }
func (s *ContinueStmt) stmtNode()           {}

type CaseClause struct {
	CasePos token.Position
	Pattern Expr
	Body    []Stmt
}

func (c *CaseClause) Pos() token.Position { return c.CasePos }

type SwitchStmt struct {
	SwitchPos token.Position
	Expr      Expr
	Cases     []*CaseClause
	Default   []Stmt
}

func (s *SwitchStmt) Pos() token.Position { return s.SwitchPos }
func (s *SwitchStmt) stmtNode()           {}

type DeferStmt struct {
	DeferPos token.Position
	Call     *CallExpr
}

func (s *DeferStmt) Pos() token.Position { return s.DeferPos }
func (s *DeferStmt) stmtNode()           {}

type CatchClause struct {
	CatchPos token.Position
	VarName  string         // catch variable name, e.g., "e"
	VarPos   token.Position // position of the variable name
	Type     TypeNode       // catch type (error, or struct type)
	Body     *BlockStmt
}

func (c *CatchClause) Pos() token.Position { return c.CatchPos }

type TryStmt struct {
	TryPos    token.Position
	Body      *BlockStmt
	CatchName string         // identifier name, e.g., "e" (legacy single catch)
	CatchPos  token.Position // position of the identifier (legacy single catch)
	CatchType TypeNode       // currently expected to be SimpleType "error" (legacy single catch)
	CatchBody *BlockStmt     // nil if no catch (legacy single catch)
	Catches   []*CatchClause // typed catch clauses
}

func (s *TryStmt) Pos() token.Position { return s.TryPos }
func (s *TryStmt) stmtNode()           {}

type WhileStmt struct {
	WhilePos token.Position
	Cond     Expr
	Body     *BlockStmt
}

func (s *WhileStmt) Pos() token.Position { return s.WhilePos }
func (s *WhileStmt) stmtNode()           {}

type ForStmt struct {
	ForPos token.Position
	Init   Stmt // may be nil
	Cond   Expr // may be nil
	Post   Stmt // may be nil
	Body   *BlockStmt
}

func (s *ForStmt) Pos() token.Position { return s.ForPos }
func (s *ForStmt) stmtNode()           {}

type ForEachStmt struct {
	ForPos   token.Position
	VarName  string
	VarPos   token.Position
	ListExpr Expr
	Body     *BlockStmt
}

func (s *ForEachStmt) Pos() token.Position { return s.ForPos }
func (s *ForEachStmt) stmtNode()           {}

// ---------- Expressions ----------

type IdentExpr struct {
	Name    string
	NamePos token.Position
}

func (e *IdentExpr) Pos() token.Position { return e.NamePos }
func (e *IdentExpr) exprNode()           {}

type IntLiteral struct {
	Value  int64
	LitPos token.Position
	Raw    string
}

func (e *IntLiteral) Pos() token.Position { return e.LitPos }
func (e *IntLiteral) exprNode()           {}

type FloatLiteral struct {
	Value  float64
	LitPos token.Position
	Raw    string
}

func (e *FloatLiteral) Pos() token.Position { return e.LitPos }
func (e *FloatLiteral) exprNode()           {}

type StringLiteral struct {
	Value  string
	LitPos token.Position
}

func (e *StringLiteral) Pos() token.Position { return e.LitPos }
func (e *StringLiteral) exprNode()           {}

type StringPart interface {
	Node
	stringPart()
}

type StringTextPart struct {
	Value   string
	PartPos token.Position
}

func (p *StringTextPart) Pos() token.Position { return p.PartPos }
func (p *StringTextPart) stringPart()         {}

type StringExprPart struct {
	Expr Expr
}

func (p *StringExprPart) Pos() token.Position { return p.Expr.Pos() }
func (p *StringExprPart) stringPart()         {}

type InterpolatedString struct {
	Parts  []StringPart
	LitPos token.Position
}

func (e *InterpolatedString) Pos() token.Position { return e.LitPos }
func (e *InterpolatedString) exprNode()           {}

type BytesLiteral struct {
	Value  []byte
	LitPos token.Position
}

func (e *BytesLiteral) Pos() token.Position { return e.LitPos }
func (e *BytesLiteral) exprNode()           {}

type BoolLiteral struct {
	Value  bool
	LitPos token.Position
}

func (e *BoolLiteral) Pos() token.Position { return e.LitPos }
func (e *BoolLiteral) exprNode()           {}

type NoneLiteral struct {
	LitPos token.Position
}

func (e *NoneLiteral) Pos() token.Position { return e.LitPos }
func (e *NoneLiteral) exprNode()           {}

type SomeLiteral struct {
	SomePos token.Position
	Value   Expr
}

func (e *SomeLiteral) Pos() token.Position { return e.SomePos }
func (e *SomeLiteral) exprNode()           {}

type ListLiteral struct {
	LBracket token.Position
	Elements []Expr
	RBracket token.Position
}

func (e *ListLiteral) Pos() token.Position { return e.LBracket }
func (e *ListLiteral) exprNode()           {}

type DictEntry struct {
	Key    string
	KeyPos token.Position
	Value  Expr
}

func (e *DictEntry) Pos() token.Position { return e.KeyPos }

type DictLiteral struct {
	LBrace  token.Position
	Entries []*DictEntry
	RBrace  token.Position
}

func (e *DictLiteral) Pos() token.Position { return e.LBrace }
func (e *DictLiteral) exprNode()           {}

type StructLiteral struct {
	TypeName    string
	TypeNamePos token.Position
	TypeArgs    []TypeNode // generic type arguments, e.g. Box<int>
	LBrace      token.Position
	Fields      []*FieldInit
	RBrace      token.Position
}

func (e *StructLiteral) Pos() token.Position { return e.TypeNamePos }
func (e *StructLiteral) exprNode()           {}

type FieldInit struct {
	Name    string
	NamePos token.Position
	Value   Expr
}

func (f *FieldInit) Pos() token.Position { return f.NamePos }

type FuncLiteral struct {
	FunPos        token.Position
	Params        []*Param
	VariadicParam *VariadicParam // nil or variadic param (args... | Args...)
	Return        TypeNode
	Throws        []TypeNode // error types after "!"
	Body          *BlockStmt
}

func (e *FuncLiteral) Pos() token.Position { return e.FunPos }
func (e *FuncLiteral) exprNode()           {}

type CallExpr struct {
	Callee   Expr
	TypeArgs []TypeNode // generic type arguments, e.g. identity<int>(x)
	LParen   token.Position
	Args     []Expr
	RParen   token.Position
}

func (e *CallExpr) Pos() token.Position { return e.Callee.Pos() }
func (e *CallExpr) exprNode()           {}

type IndexExpr struct {
	X        Expr
	LBracket token.Position
	Index    Expr
	RBracket token.Position
}

func (e *IndexExpr) Pos() token.Position { return e.X.Pos() }
func (e *IndexExpr) exprNode()           {}

type MemberExpr struct {
	X       Expr
	Name    string
	NamePos token.Position
}

func (e *MemberExpr) Pos() token.Position { return e.X.Pos() }
func (e *MemberExpr) exprNode()           {}

type OptionalMemberExpr struct {
	X       Expr
	Name    string
	NamePos token.Position
}

func (e *OptionalMemberExpr) Pos() token.Position { return e.X.Pos() }
func (e *OptionalMemberExpr) exprNode()           {}

type OptionalCallExpr struct {
	Callee Expr
	LParen token.Position
	Args   []Expr
	RParen token.Position
}

func (e *OptionalCallExpr) Pos() token.Position { return e.Callee.Pos() }
func (e *OptionalCallExpr) exprNode()           {}

type BinaryExpr struct {
	OpPos token.Position
	Op    token.Kind
	Left  Expr
	Right Expr
}

func (e *BinaryExpr) Pos() token.Position { return e.OpPos }
func (e *BinaryExpr) exprNode()           {}

type UnaryExpr struct {
	OpPos token.Position
	Op    token.Kind
	X     Expr
}

func (e *UnaryExpr) Pos() token.Position { return e.OpPos }
func (e *UnaryExpr) exprNode()           {}

type NamedArg struct {
	Name    string
	NamePos token.Position
	Value   Expr
}

func (a *NamedArg) Pos() token.Position { return a.NamePos }
func (a *NamedArg) exprNode()           {}

type AwaitExpr struct {
	AwaitPos token.Position
	Expr     Expr
}

func (e *AwaitExpr) Pos() token.Position { return e.AwaitPos }
func (e *AwaitExpr) exprNode()           {}

// ValuePackExpansion represents a variadic value pack expansion in expression position: args...
type ValuePackExpansion struct {
	Name    string
	NamePos token.Position
}

func (e *ValuePackExpansion) Pos() token.Position { return e.NamePos }
func (e *ValuePackExpansion) exprNode()           {}
