package types

import (
	"fmt"
	"strings"

	"avenir/internal/ast"
	"avenir/internal/runtime/builtins"
	"avenir/internal/token"
)

// ----- Type Checking Errors -----

type Error struct {
	Pos token.Position
	Msg string
}

func (e Error) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.Pos.Line, e.Pos.Column, e.Msg)
}

// ----- Symbols and Scopes -----

type SymbolKind int

const (
	SymVar SymbolKind = iota
	SymFunc
	SymModule
	SymType
)

type Symbol struct {
	Name     string
	Kind     SymbolKind
	Type     Type
	Node     ast.Node
	Module   *ModuleInfo // only set for SymModule
	IsPublic bool        // true if function is public (pub fun)
	IsGlobal bool        // true for module-level variable declarations
}

// ModuleInfo holds type-checking information for a module.
type ModuleInfo struct {
	Name  string // e.g. "std.io"
	Prog  *ast.Program
	Scope *Scope // top-level symbols (functions, etc.)
}

// World represents all modules in a program.
type World struct {
	Modules map[string]*ModuleInfo
	Entry   string // Entry module name
}

// Bindings stores static resolution info for expressions:
//
//   - Idents: all ast.IdentExpr resolved to some symbol in scope
//   - Members: all ast.MemberExpr that refer to module members (e.g. math.pow)
//   - ExprTypes: expression -> its type (for struct field access and other uses)
//
// DecoratorInfo holds resolved decorator information for the IR compiler.
type DecoratorInfo struct {
	Expr     ast.Expr // decorator expression (compiled by IR to produce the decorator closure)
	FuncType *Func    // type of the final decorator: fun(fn) | fn
}

type Bindings struct {
	Idents    map[*ast.IdentExpr]*Symbol
	Members   map[*ast.MemberExpr]*Symbol
	ExprTypes map[ast.Expr]Type

	MonomorphizedStructs map[string]*Struct                // monoName -> instantiated struct type
	MonomorphizedFuncs   map[string]*ast.FunDecl           // monoName -> synthetic FunDecl
	Decorators           map[*ast.FunDecl][]*DecoratorInfo // decorated func -> resolved decorators
}

func NewBindings() *Bindings {
	return &Bindings{
		Idents:               make(map[*ast.IdentExpr]*Symbol),
		Members:              make(map[*ast.MemberExpr]*Symbol),
		ExprTypes:            make(map[ast.Expr]Type),
		MonomorphizedStructs: make(map[string]*Struct),
		MonomorphizedFuncs:   make(map[string]*ast.FunDecl),
		Decorators:           make(map[*ast.FunDecl][]*DecoratorInfo),
	}
}

type Scope struct {
	parent  *Scope
	symbols map[string]*Symbol
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent:  parent,
		symbols: make(map[string]*Symbol),
	}
}

func (s *Scope) Insert(sym *Symbol) error {
	if _, exists := s.symbols[sym.Name]; exists {
		return fmt.Errorf("redefinition of %q", sym.Name)
	}
	s.symbols[sym.Name] = sym
	return nil
}

func (s *Scope) Lookup(name string) *Symbol {
	for sc := s; sc != nil; sc = sc.parent {
		if sym, ok := sc.symbols[name]; ok {
			return sym
		}
	}
	return nil
}

// ----- Checker -----

type Checker struct {
	global *Scope
	scope  *Scope

	errors []error

	currentReturn Type
	currentThrows []Type // error types the current function can throw

	bindings *Bindings // optional binding info sink

	loopDepth      int                   // tracks nesting level of loops for break validation
	structTypes    map[string]*Struct    // struct name -> Struct type
	interfaceTypes map[string]*Interface // interface name -> Interface type

	// Visibility tracking
	currentModule   string  // name of the module currently being checked
	currentReceiver *Struct // struct type of the current method receiver (if any)

	// Generics tracking
	genericStructs map[string]*GenericStruct // generic struct name -> definition
	genericFuncs   map[string]*GenericFunc   // generic func name -> definition
	monomorphized  map[string]bool           // monomorph key -> already generated
	typeParamScope map[string]Type           // current type parameter bindings (T -> int)
}

// CheckProgram type-checks a program and returns a list of errors (if any).
// For single-file programs without imports.
func CheckProgram(prog *ast.Program) []error {
	_, errs := CheckProgramWithBindings(prog)
	return errs
}

// CheckProgramWithBindings type-checks a single program and returns bindings.
func CheckProgramWithBindings(prog *ast.Program) (*Bindings, []error) {
	// Build a minimal world with just this program
	world := &World{
		Modules: make(map[string]*ModuleInfo),
	}
	modName := "main"
	if prog.Package != nil {
		modName = prog.Package.Name
	}
	world.Modules[modName] = &ModuleInfo{
		Name:  modName,
		Prog:  prog,
		Scope: NewScope(nil),
	}
	world.Entry = modName
	return CheckWorldWithBindings(world)
}

// CheckWorld type-checks all modules in a world.
func CheckWorld(world *World) []error {
	_, errs := CheckWorldWithBindings(world)
	return errs
}

// CheckWorldWithBindings type-checks all modules in a world and returns bindings.
func CheckWorldWithBindings(world *World) (*Bindings, []error) {
	bindings := NewBindings()
	var allErrors []error

	// Phase 1a: Create scopes and register builtins for each module
	for _, modInfo := range world.Modules {
		modInfo.Scope = NewScope(nil)
		c := &Checker{
			global:        modInfo.Scope,
			bindings:      bindings,
			currentModule: modInfo.Name,
		}
		c.declareBuiltins()

		// Register all top-level struct names (forward declarations only)
		for _, st := range modInfo.Prog.Structs {
			c.forwardDeclareStruct(st)
		}

		allErrors = append(allErrors, c.errors...)
	}

	// Phase 1b: Process imports so that cross-module types are available
	for _, modInfo := range world.Modules {
		c := &Checker{
			global:        modInfo.Scope,
			bindings:      bindings,
			currentModule: modInfo.Name,
		}
		c.scope = c.global
		for _, imp := range modInfo.Prog.Imports {
			importFQN := strings.Join(imp.Path, ".")
			importedMod, ok := world.Modules[importFQN]
			if !ok {
				continue
			}
			localAlias := imp.Alias
			if localAlias == "" {
				if len(imp.Path) > 0 {
					localAlias = imp.Path[len(imp.Path)-1]
				} else {
					localAlias = importFQN
				}
			}
			_ = c.scope.Insert(&Symbol{
				Name:   localAlias,
				Kind:   SymModule,
				Type:   nil,
				Node:   imp,
				Module: importedMod,
			})
			if localAlias != importFQN {
				_ = c.scope.Insert(&Symbol{
					Name:   importFQN,
					Kind:   SymModule,
					Type:   nil,
					Node:   imp,
					Module: importedMod,
				})
			}
		}
		allErrors = append(allErrors, c.errors...)
	}

	// Phase 1c: Resolve struct fields, interfaces, functions, and top-level vars
	for _, modInfo := range world.Modules {
		c := &Checker{
			global:        modInfo.Scope,
			bindings:      bindings,
			currentModule: modInfo.Name,
		}
		c.scope = c.global

		for _, st := range modInfo.Prog.Structs {
			c.resolveStructFields(st)
		}

		for _, iface := range modInfo.Prog.Interfaces {
			c.declareInterface(iface)
		}

		for _, fn := range modInfo.Prog.Funcs {
			c.declareFunc(fn)
		}

		for _, v := range modInfo.Prog.Vars {
			var varType Type
			if v.Type != nil {
				varType = c.typeOfTypeNode(v.Type)
			} else {
				varType = Any
			}
			_ = c.global.Insert(&Symbol{
				Name:     v.Name,
				Kind:     SymVar,
				Type:     varType,
				Node:     v,
				IsGlobal: true,
			})
		}

		allErrors = append(allErrors, c.errors...)
	}

	// Phase 2: Process imports and type-check each module
	for _, modInfo := range world.Modules {
		c := &Checker{
			global:        modInfo.Scope,
			bindings:      bindings,
			currentModule: modInfo.Name,
		}
		c.scope = c.global

		// Process imports: add module symbols to scope
		for _, imp := range modInfo.Prog.Imports {
			importFQN := strings.Join(imp.Path, ".")
			importedMod, ok := world.Modules[importFQN]
			if !ok {
				c.addError(imp.ImportPos, "cannot find module %q", importFQN)
				continue
			}

			// Determine local alias
			localAlias := imp.Alias
			if localAlias == "" {
				// Use last segment of path
				if len(imp.Path) > 0 {
					localAlias = imp.Path[len(imp.Path)-1]
				} else {
					localAlias = importFQN
				}
			}

			// Insert module symbol
			_ = c.scope.Insert(&Symbol{
				Name:   localAlias,
				Kind:   SymModule,
				Type:   nil,
				Node:   imp,
				Module: importedMod,
			})

			// Also insert full module path to support qualified types like std.net.Socket.
			if localAlias != importFQN {
				_ = c.scope.Insert(&Symbol{
					Name:   importFQN,
					Kind:   SymModule,
					Type:   nil,
					Node:   imp,
					Module: importedMod,
				})
			}
		}

		// Type-check all top-level variables in this module
		for _, v := range modInfo.Prog.Vars {
			c.checkTopLevelVar(v)
		}

		// Type-check all functions in this module
		for _, fn := range modInfo.Prog.Funcs {
			c.checkFunc(fn)
		}

		allErrors = append(allErrors, c.errors...)
	}

	return bindings, allErrors
}

func (c *Checker) addError(pos token.Position, msg string, args ...interface{}) {
	e := Error{
		Pos: pos,
		Msg: fmt.Sprintf(msg, args...),
	}
	c.errors = append(c.errors, e)
}

func (c *Checker) declareBuiltins() {
	for _, meta := range builtins.All() {
		if meta.ReceiverType != builtins.TypeVoid {
			continue
		}
		fnType := c.builtinFuncTypeFromMeta(meta)

		_ = c.global.Insert(&Symbol{
			Name:     meta.Name,
			Kind:     SymFunc,
			Type:     fnType,
			Node:     nil,
			IsPublic: false, // builtins are always accessible
		})
	}
}

func (c *Checker) builtinFuncTypeFromMeta(meta builtins.Meta) *Func {
	paramTypes := make([]Type, 0, len(meta.Params))
	for _, p := range meta.Params {
		paramTypes = append(paramTypes, c.typeFromBuiltinTypeRef(p))
	}
	res := c.typeFromBuiltinTypeRef(meta.Result)
	if builtins.IsAsyncBuiltin(meta.ID) {
		res = &Future{Inner: res}
	}
	return &Func{
		ParamTypes: paramTypes,
		Result:     res,
	}
}

func (c *Checker) typeFromBuiltinTypeRef(tr builtins.TypeRef) Type {
	switch tr.Kind {
	case builtins.TypeInt:
		return Int
	case builtins.TypeFloat:
		return Float
	case builtins.TypeString:
		return String
	case builtins.TypeBool:
		return Bool
	case builtins.TypeVoid:
		return Void
	case builtins.TypeAny:
		return Any
	case builtins.TypeList:
		elemTypes := make([]Type, 0, len(tr.Elem))
		for _, e := range tr.Elem {
			elemTypes = append(elemTypes, c.typeFromBuiltinTypeRef(e))
		}
		return &List{ElementTypes: elemTypes}
	case builtins.TypeDict:
		var keyType Type
		var valueType Type = Any
		if len(tr.Elem) == 2 {
			keyType = c.typeFromBuiltinTypeRef(tr.Elem[0])
			valueType = c.typeFromBuiltinTypeRef(tr.Elem[1])
		} else if len(tr.Elem) == 1 {
			valueType = c.typeFromBuiltinTypeRef(tr.Elem[0])
		}
		return &Dict{KeyType: keyType, ValueType: valueType}
	case builtins.TypeError:
		return ErrorType
	case builtins.TypeBytes:
		return Bytes
	case builtins.TypeUnion:
		variants := make([]Type, 0, len(tr.Elem))
		for _, v := range tr.Elem {
			vt := c.typeFromBuiltinTypeRef(v)
			if !IsInvalid(vt) {
				variants = append(variants, vt)
			}
		}
		if len(variants) == 0 {
			return Invalid
		}
		return &Union{Variants: variants}
	default:
		// for now, treat unknown as Invalid
		return Invalid
	}
}

// ----- Interfaces -----

func (c *Checker) declareInterface(iface *ast.InterfaceDecl) {
	// Build method list
	methods := make([]InterfaceMethod, 0, len(iface.Methods))
	methodNames := make(map[string]bool)

	for _, m := range iface.Methods {
		if methodNames[m.Name] {
			c.addError(m.Pos(), "duplicate method name %q in interface %q", m.Name, iface.Name)
			continue
		}
		methodNames[m.Name] = true

		// Parse parameter types
		paramTypes := make([]Type, 0, len(m.ParamTypes))
		for _, pt := range m.ParamTypes {
			paramType := c.typeOfTypeNode(pt)
			if IsInvalid(paramType) {
				continue
			}
			paramTypes = append(paramTypes, paramType)
		}

		// Parse return type
		returnType := c.typeOfTypeNode(m.Return)
		if IsInvalid(returnType) {
			continue
		}

		methods = append(methods, InterfaceMethod{
			Name:       m.Name,
			ParamTypes: paramTypes,
			Return:     returnType,
		})
	}

	// Create interface type
	interfaceType := &Interface{
		Name:           iface.Name,
		Methods:        methods,
		IsPublic:       iface.IsPublic,
		DefiningModule: c.currentModule,
	}

	// Store in checker's interface registry
	if c.interfaceTypes == nil {
		c.interfaceTypes = make(map[string]*Interface)
	}
	c.interfaceTypes[iface.Name] = interfaceType

	// Insert into scope
	if err := c.global.Insert(&Symbol{
		Name:     iface.Name,
		Kind:     SymType,
		Type:     interfaceType,
		Node:     iface,
		IsPublic: iface.IsPublic,
	}); err != nil {
		c.addError(iface.Pos(), "interface %q: %v", iface.Name, err)
	}
}

// ----- Structs -----

func (c *Checker) forwardDeclareStruct(st *ast.StructDecl) {
	if len(st.TypeParams) > 0 {
		if c.genericStructs == nil {
			c.genericStructs = make(map[string]*GenericStruct)
		}
		paramNames := make([]string, len(st.TypeParams))
		for i, tp := range st.TypeParams {
			paramNames[i] = tp.Name
		}
		gs := &GenericStruct{Decl: st, TypeParams: paramNames}
		c.genericStructs[st.Name] = gs

		if err := c.global.Insert(&Symbol{
			Name:     st.Name,
			Kind:     SymType,
			Type:     gs,
			Node:     st,
			IsPublic: st.IsPublic,
		}); err != nil {
			c.addError(st.Pos(), "struct %q: %v", st.Name, err)
		}
		return
	}

	structType := &Struct{
		Name:            st.Name,
		Fields:          nil,
		IsPublic:        st.IsPublic,
		IsMutable:       st.IsMutable,
		InstanceMethods: make(map[string]*Method),
		StaticMethods:   make(map[string]*Method),
	}

	if c.structTypes == nil {
		c.structTypes = make(map[string]*Struct)
	}
	c.structTypes[st.Name] = structType

	if err := c.global.Insert(&Symbol{
		Name:     st.Name,
		Kind:     SymType,
		Type:     structType,
		Node:     st,
		IsPublic: st.IsPublic,
	}); err != nil {
		c.addError(st.Pos(), "struct %q: %v", st.Name, err)
	}
}

func (c *Checker) resolveStructFields(st *ast.StructDecl) {
	if len(st.TypeParams) > 0 {
		return
	}

	// Look up the struct type from scope (may have been registered by a different checker instance)
	var structType *Struct
	if c.structTypes != nil {
		structType = c.structTypes[st.Name]
	}
	if structType == nil {
		sym := c.global.Lookup(st.Name)
		if sym == nil || sym.Kind != SymType {
			return
		}
		var ok bool
		structType, ok = sym.Type.(*Struct)
		if !ok {
			return
		}
	}

	fields := make([]Field, 0, len(st.Fields))
	fieldNames := make(map[string]bool)

	for _, f := range st.Fields {
		if fieldNames[f.Name] {
			c.addError(f.Pos(), "duplicate field name %q in struct %q", f.Name, st.Name)
			continue
		}
		fieldNames[f.Name] = true

		if !st.IsPublic && f.IsPublic {
			c.addError(f.Pos(), "public field %q declared in private struct %q", f.Name, st.Name)
			f.IsPublic = false
		}

		fieldType := c.typeOfTypeNode(f.Type)
		if IsInvalid(fieldType) {
			continue
		}

		if f.DefaultExpr != nil {
			if !c.isCompileTimeConstant(f.DefaultExpr) {
				c.addError(f.DefaultExpr.Pos(), "default value for field %q must be a compile-time constant", f.Name)
				continue
			}

			defaultType := c.checkExpr(f.DefaultExpr)
			if !c.assignable(fieldType, defaultType) {
				c.addError(f.DefaultExpr.Pos(), "default value for field %q has type %s, expected %s",
					f.Name, defaultType.String(), fieldType.String())
				continue
			}
		}

		fieldMutable := st.IsMutable
		if f.IsMutable {
			fieldMutable = true
		}

		fields = append(fields, Field{
			Name:        f.Name,
			Type:        fieldType,
			IsPublic:    f.IsPublic,
			IsMutable:   fieldMutable,
			DefaultExpr: f.DefaultExpr,
		})
	}

	structType.Fields = fields
}

func (c *Checker) declareStruct(st *ast.StructDecl) {
	c.forwardDeclareStruct(st)
	c.resolveStructFields(st)
}

// ----- Functions -----

func (c *Checker) declareFunc(fn *ast.FunDecl) {
	// If the function has type parameters, register it as a generic func template
	if len(fn.TypeParams) > 0 && fn.Receiver == nil {
		if c.genericFuncs == nil {
			c.genericFuncs = make(map[string]*GenericFunc)
		}
		paramNames := make([]string, len(fn.TypeParams))
		for i, tp := range fn.TypeParams {
			paramNames[i] = tp.Name
		}
		gf := &GenericFunc{Decl: fn, TypeParams: paramNames}
		c.genericFuncs[fn.Name] = gf

		if err := c.global.Insert(&Symbol{
			Name:     fn.Name,
			Kind:     SymFunc,
			Type:     gf,
			Node:     fn,
			IsPublic: fn.IsPublic,
		}); err != nil {
			c.addError(fn.Pos(), "function %q: %v", fn.Name, err)
		}
		return
	}

	// Collect parameter types
	var params []Type
	for _, p := range fn.Params {
		pt := c.typeOfTypeNode(p.Type)
		params = append(params, pt)
	}

	retType := c.typeOfTypeNode(fn.Return)

	// If this is a method, register it with the receiver type
	if fn.Receiver != nil {
		receiverType := c.typeOfTypeNode(fn.Receiver.Type)
		if IsInvalid(receiverType) {
			return
		}

		// Only struct types can have methods for now
		structType, ok := receiverType.(*Struct)
		if !ok {
			c.addError(fn.Receiver.Type.Pos(), "methods can only be defined on struct types, got %s", receiverType.String())
			return
		}

		// Initialize methods maps if needed
		if structType.InstanceMethods == nil {
			structType.InstanceMethods = make(map[string]*Method)
		}
		if structType.StaticMethods == nil {
			structType.StaticMethods = make(map[string]*Method)
		}

		isStatic := fn.Receiver.Kind == ast.ReceiverStatic

		// Check for duplicate method name (check both maps)
		var methodMap map[string]*Method
		if isStatic {
			methodMap = structType.StaticMethods
			if _, exists := structType.StaticMethods[fn.Name]; exists {
				c.addError(fn.Pos(), "duplicate static method %q on type %s", fn.Name, structType.Name)
				return
			}
			// Also check instance methods to prevent name collision
			if _, exists := structType.InstanceMethods[fn.Name]; exists {
				c.addError(fn.Pos(), "static method %q conflicts with instance method on type %s", fn.Name, structType.Name)
				return
			}
		} else {
			methodMap = structType.InstanceMethods
			if _, exists := structType.InstanceMethods[fn.Name]; exists {
				c.addError(fn.Pos(), "duplicate instance method %q on type %s", fn.Name, structType.Name)
				return
			}
			// Also check static methods to prevent name collision
			if _, exists := structType.StaticMethods[fn.Name]; exists {
				c.addError(fn.Pos(), "instance method %q conflicts with static method on type %s", fn.Name, structType.Name)
				return
			}
		}

		// For instance methods, receiver is the first parameter
		// For static methods, receiver is NOT a parameter (only for identification)
		var methodParams []Type
		if isStatic {
			// Static method: parameters are just the function parameters
			methodParams = params
		} else {
			// Instance method: receiver is the first parameter
			methodParams = append([]Type{receiverType}, params...)
		}

		methodResult := retType
		if fn.IsAsync {
			methodResult = &Future{Inner: retType}
		}

		// Register the method
		methodMap[fn.Name] = &Method{
			Name:       fn.Name,
			Receiver:   receiverType,
			ParamTypes: methodParams,
			Result:     methodResult,
			IsStatic:   isStatic,
			IsPublic:   fn.IsPublic,
			Decl:       fn,
		}
		return // Methods are not inserted into global scope
	}

	// For async functions, wrap the return type in Future<T>
	resultType := retType
	if fn.IsAsync {
		resultType = &Future{Inner: retType}
	}

	// Resolve throws types
	var throwsTypes []Type
	for _, tn := range fn.Throws {
		tt := c.typeOfTypeNode(tn)
		throwsTypes = append(throwsTypes, tt)
	}

	// Regular function: insert into global scope
	fnType := &Func{
		ParamTypes: params,
		Result:     resultType,
		Throws:     throwsTypes,
	}

	if err := c.global.Insert(&Symbol{
		Name:     fn.Name,
		Kind:     SymFunc,
		Type:     fnType,
		Node:     fn,
		IsPublic: fn.IsPublic,
	}); err != nil {
		c.addError(fn.Pos(), "function %q: %v", fn.Name, err)
	}
}

func (c *Checker) checkFunc(fn *ast.FunDecl) {
	// Skip uninstantiated generic functions - they are checked upon instantiation
	if len(fn.TypeParams) > 0 {
		return
	}

	var fnType *Func

	// For methods, look up the method in the struct's method set
	if fn.Receiver != nil {
		receiverType := c.typeOfTypeNode(fn.Receiver.Type)
		if IsInvalid(receiverType) {
			return
		}
		structType, ok := receiverType.(*Struct)
		if !ok {
			c.addError(fn.Receiver.Type.Pos(), "methods can only be defined on struct types, got %s", receiverType.String())
			return
		}

		// Look up method in appropriate map
		var method *Method
		var found bool
		if fn.Receiver.Kind == ast.ReceiverStatic {
			if structType.StaticMethods != nil {
				method, found = structType.StaticMethods[fn.Name]
			}
		} else {
			if structType.InstanceMethods != nil {
				method, found = structType.InstanceMethods[fn.Name]
			}
		}

		if !found || method == nil {
			c.addError(fn.Pos(), "internal error: method %q not found on struct %s", fn.Name, structType.Name)
			return
		}

		// For instance methods, receiver is already in ParamTypes as first parameter
		// For static methods, ParamTypes doesn't include receiver
		fnType = &Func{
			ParamTypes: method.ParamTypes,
			Result:     method.Result,
		}
	} else {
		// Regular function: look up in global scope
		sym := c.global.Lookup(fn.Name)
		if sym == nil {
			c.addError(fn.Pos(), "internal error: function %q not found in global scope", fn.Name)
			return
		}
		var ok bool
		fnType, ok = sym.Type.(*Func)
		if !ok {
			c.addError(fn.Pos(), "internal error: symbol for function %q is not Func", fn.Name)
			return
		}
	}

	// Function scope
	prevScope := c.scope
	c.scope = NewScope(c.global)
	defer func() { c.scope = prevScope }()

	prevRet := c.currentReturn
	// For async functions, the declared result is Future<T> but return statements
	// should check against T (the inner type), not Future<T>.
	if fn.IsAsync {
		if fut, ok := fnType.Result.(*Future); ok {
			c.currentReturn = fut.Inner
		} else {
			c.currentReturn = fnType.Result
		}
	} else {
		c.currentReturn = fnType.Result
	}
	defer func() { c.currentReturn = prevRet }()

	prevThrows := c.currentThrows
	c.currentThrows = fnType.Throws
	defer func() { c.currentThrows = prevThrows }()

	// For instance methods, insert receiver into scope and track it for visibility checks
	// Static methods don't have a receiver variable
	var prevReceiver *Struct
	if fn.Receiver != nil && fn.Receiver.Kind == ast.ReceiverInstance {
		receiverType := c.typeOfTypeNode(fn.Receiver.Type)
		if !IsInvalid(receiverType) {
			if structType, ok := receiverType.(*Struct); ok {
				// Track receiver for visibility checks (methods can access private fields)
				prevReceiver = c.currentReceiver
				c.currentReceiver = structType
			}
			if err := c.scope.Insert(&Symbol{
				Name: fn.Receiver.Name,
				Kind: SymVar,
				Type: receiverType,
				Node: fn, // Use FunDecl as the node since Receiver doesn't implement Node
			}); err != nil {
				c.addError(fn.Receiver.NamePos, "receiver %q: %v", fn.Receiver.Name, err)
			}
		}
	}
	defer func() {
		if fn.Receiver != nil && fn.Receiver.Kind == ast.ReceiverInstance {
			c.currentReceiver = prevReceiver
		}
	}()

	// Parameters
	// First, enforce required before optional rule
	sawDefault := false
	for _, param := range fn.Params {
		if param.Default != nil {
			sawDefault = true
		} else if sawDefault {
			c.addError(param.Pos(), "parameter %q without default cannot follow parameter with default", param.Name)
		}
	}

	// Insert parameters into scope
	// For instance methods, skip the first parameter type (receiver) when indexing into ParamTypes
	// For static methods, ParamTypes doesn't include receiver, so no offset needed
	paramTypeOffset := 0
	if fn.Receiver != nil && fn.Receiver.Kind == ast.ReceiverInstance {
		paramTypeOffset = 1 // Skip receiver type for instance methods
	}
	for i, param := range fn.Params {
		pt := fnType.ParamTypes[paramTypeOffset+i]
		if err := c.scope.Insert(&Symbol{
			Name: param.Name,
			Kind: SymVar,
			Type: pt,
			Node: param,
		}); err != nil {
			c.addError(param.Pos(), "parameter %q: %v", param.Name, err)
		}
	}

	// Type-check default expressions
	for i, param := range fn.Params {
		if param.Default != nil {
			defaultType := c.checkExpr(param.Default)
			paramType := fnType.ParamTypes[paramTypeOffset+i]
			if !c.assignable(paramType, defaultType) {
				c.addError(param.Default.Pos(), "default value for parameter %q has type %s, expected %s",
					param.Name, defaultType.String(), paramType.String())
			}
		}
	}

	// Body
	c.checkBlock(fn.Body)

	// Decorators
	c.checkDecorators(fn, fnType)
}

func (c *Checker) checkDecorators(fn *ast.FunDecl, fnType *Func) {
	if len(fn.Decorators) == 0 {
		return
	}

	var infos []*DecoratorInfo
	for _, dec := range fn.Decorators {
		info := c.checkDecorator(dec, fnType, fn)
		if info != nil {
			infos = append(infos, info)
		}
	}

	if c.bindings != nil && len(infos) > 0 {
		c.bindings.Decorators[fn] = infos
	}
}

func (c *Checker) checkDecorator(dec *ast.Decorator, fnType *Func, fn *ast.FunDecl) *DecoratorInfo {
	exprType := c.checkExpr(dec.Expr)
	if IsInvalid(exprType) {
		return nil
	}

	if gf, ok := exprType.(*GenericFunc); ok {
		return c.checkGenericDecorator(dec, gf, fnType)
	}

	decFunc, ok := exprType.(*Func)
	if !ok {
		c.addError(dec.AtPos, "decorator expression must be a function, got %s", exprType.String())
		return nil
	}

	return c.validateDecoratorFunc(dec, decFunc, fnType)
}

func (c *Checker) validateDecoratorFunc(dec *ast.Decorator, decFunc *Func, fnType *Func) *DecoratorInfo {
	if len(decFunc.ParamTypes) != 1 {
		c.addError(dec.AtPos, "decorator must accept exactly 1 argument (the function), got %d", len(decFunc.ParamTypes))
		return nil
	}
	paramFn, ok := decFunc.ParamTypes[0].(*Func)
	if !ok {
		c.addError(dec.AtPos, "decorator parameter must be a function type, got %s", decFunc.ParamTypes[0].String())
		return nil
	}
	if !Equal(paramFn, fnType) && !c.asyncFuncCompatible(paramFn, fnType) {
		c.addError(dec.AtPos, "decorator expects function of type %s, but decorated function has type %s", paramFn.String(), fnType.String())
		return nil
	}
	retFn, ok := decFunc.Result.(*Func)
	if !ok {
		c.addError(dec.AtPos, "decorator must return a function type, got %s", decFunc.Result.String())
		return nil
	}
	if !Equal(retFn, fnType) && !c.asyncFuncCompatible(retFn, fnType) {
		c.addError(dec.AtPos, "decorator returns function of type %s, but decorated function has type %s", retFn.String(), fnType.String())
		return nil
	}
	return &DecoratorInfo{
		Expr:     dec.Expr,
		FuncType: decFunc,
	}
}

func (c *Checker) asyncFuncCompatible(expected *Func, actual *Func) bool {
	if len(expected.ParamTypes) != len(actual.ParamTypes) {
		return false
	}
	for i, p := range expected.ParamTypes {
		if !Equal(p, actual.ParamTypes[i]) {
			return false
		}
	}
	fut, ok := actual.Result.(*Future)
	if !ok {
		return false
	}
	return Equal(expected.Result, fut.Inner)
}

func (c *Checker) checkGenericDecorator(dec *ast.Decorator, gf *GenericFunc, fnType *Func) *DecoratorInfo {
	typeArgs := c.inferDecoratorTypeArgs(gf, fnType, dec)
	if typeArgs == nil {
		return nil
	}

	instantiated, monoName := c.instantiateGenericFunc(gf, typeArgs)
	if instantiated == nil {
		return nil
	}

	if len(instantiated.ParamTypes) != 1 {
		c.addError(dec.AtPos, "instantiated decorator must accept exactly 1 argument, got %d", len(instantiated.ParamTypes))
		return nil
	}
	paramFn, ok := instantiated.ParamTypes[0].(*Func)
	if !ok {
		c.addError(dec.AtPos, "instantiated decorator parameter must be a function type")
		return nil
	}
	if !Equal(paramFn, fnType) {
		c.addError(dec.AtPos, "instantiated decorator expects %s, decorated function is %s", paramFn.String(), fnType.String())
		return nil
	}
	retFn, ok := instantiated.Result.(*Func)
	if !ok {
		c.addError(dec.AtPos, "instantiated decorator must return a function type")
		return nil
	}
	if !Equal(retFn, fnType) {
		c.addError(dec.AtPos, "instantiated decorator returns %s, but decorated function is %s", retFn.String(), fnType.String())
		return nil
	}

	syntheticIdent := &ast.IdentExpr{Name: monoName, NamePos: dec.AtPos}
	if c.bindings != nil {
		sym := &Symbol{Name: monoName, Kind: SymFunc, Type: instantiated}
		c.bindings.Idents[syntheticIdent] = sym
	}

	return &DecoratorInfo{
		Expr:     syntheticIdent,
		FuncType: instantiated,
	}
}

func (c *Checker) inferDecoratorTypeArgs(gf *GenericFunc, fnType *Func, dec *ast.Decorator) []Type {
	decl := gf.Decl
	if len(decl.Params) != 1 {
		c.addError(dec.AtPos, "generic decorator must have exactly 1 parameter")
		return nil
	}

	paramTypeNode := decl.Params[0].Type
	funcTypeNode, ok := paramTypeNode.(*ast.FuncType)
	if !ok {
		c.addError(dec.AtPos, "generic decorator parameter must be a function type")
		return nil
	}

	typeArgs := make([]Type, len(gf.TypeParams))
	matched := make([]bool, len(gf.TypeParams))

	paramIndex := func(name string) int {
		for i, tp := range gf.TypeParams {
			if tp == name {
				return i
			}
		}
		return -1
	}

	resultTypeNode := funcTypeNode.Result
	if simple, ok := resultTypeNode.(*ast.SimpleType); ok {
		idx := paramIndex(simple.Name)
		if idx >= 0 {
			typeArgs[idx] = fnType.Result
			matched[idx] = true
		}
	}

	concreteParamIdx := 0
	for _, ptNode := range funcTypeNode.ParamTypes {
		switch pt := ptNode.(type) {
		case *ast.TypePackExpansion:
			idx := paramIndex(pt.Name)
			if idx >= 0 {
				remaining := fnType.ParamTypes[concreteParamIdx:]
				pack := &TypePack{Types: make([]Type, len(remaining))}
				copy(pack.Types, remaining)
				typeArgs[idx] = pack
				matched[idx] = true
				concreteParamIdx = len(fnType.ParamTypes)
			}
		case *ast.SimpleType:
			idx := paramIndex(pt.Name)
			if idx >= 0 && concreteParamIdx < len(fnType.ParamTypes) {
				typeArgs[idx] = fnType.ParamTypes[concreteParamIdx]
				matched[idx] = true
				concreteParamIdx++
			} else {
				concreteParamIdx++
			}
		default:
			concreteParamIdx++
		}
	}

	for i, m := range matched {
		if !m {
			c.addError(dec.AtPos, "cannot infer type argument %q for generic decorator", gf.TypeParams[i])
			return nil
		}
	}

	return typeArgs
}

func (c *Checker) checkFuncLiteral(lit *ast.FuncLiteral) Type {
	// Build function type from parameters and return type
	var paramTypes []Type
	for _, param := range lit.Params {
		pt := c.typeOfTypeNode(param.Type)
		paramTypes = append(paramTypes, pt)
	}
	res := c.typeOfTypeNode(lit.Return)
	fnType := &Func{
		ParamTypes: paramTypes,
		Result:     res,
	}

	// Create a new scope with current scope as parent
	prevScope := c.scope
	c.scope = NewScope(prevScope)
	defer func() { c.scope = prevScope }()

	// Insert parameters into scope
	for i, param := range lit.Params {
		pt := paramTypes[i]
		if err := c.scope.Insert(&Symbol{
			Name: param.Name,
			Kind: SymVar,
			Type: pt,
			Node: param,
		}); err != nil {
			c.addError(param.Pos(), "parameter %q: %v", param.Name, err)
		}
	}

	// Type-check default expressions
	for i, param := range lit.Params {
		if param.Default != nil {
			defaultType := c.checkExpr(param.Default)
			paramType := paramTypes[i]
			if !c.assignable(paramType, defaultType) {
				c.addError(param.Default.Pos(), "default value for parameter %q has type %s, expected %s",
					param.Name, defaultType.String(), paramType.String())
			}
		}
	}

	// Save current return type and set it for the function body
	prevReturn := c.currentReturn
	c.currentReturn = res
	defer func() { c.currentReturn = prevReturn }()

	// Check body
	c.checkBlock(lit.Body)

	return fnType
}

// ----- Types from AST -----

func (c *Checker) typeOfTypeNode(tn ast.TypeNode) Type {
	switch t := tn.(type) {
	case *ast.SimpleType:
		switch t.Name {
		case "int":
			return Int
		case "float":
			return Float
		case "string":
			return String
		case "bool":
			return Bool
		case "void":
			return Void
		case "any":
			return Any
		case "error":
			return ErrorType
		case "bytes":
			return Bytes
		default:
			// Check if it's a type parameter in scope
			if c.typeParamScope != nil {
				if concrete, ok := c.typeParamScope[t.Name]; ok {
					return concrete
				}
			}
			// Check if it's a user-defined struct type
			if c.structTypes != nil {
				if structType, ok := c.structTypes[t.Name]; ok {
					return structType
				}
			}
			// Check if it's a user-defined interface type
			if c.interfaceTypes != nil {
				if interfaceType, ok := c.interfaceTypes[t.Name]; ok {
					return interfaceType
				}
			}
			// Also check in scope (for imported types)
			if sym := c.scope.Lookup(t.Name); sym != nil && sym.Kind == SymType {
				return sym.Type
			}
			c.addError(t.Pos(), "unknown type %q", t.Name)
			return Invalid
		}

	case *ast.QualifiedType:
		if len(t.Path) < 2 {
			c.addError(t.Pos(), "invalid qualified type")
			return Invalid
		}
		moduleName := strings.Join(t.Path[:len(t.Path)-1], ".")
		typeName := t.Path[len(t.Path)-1]
		modSym := c.scope.Lookup(moduleName)
		if modSym == nil || modSym.Kind != SymModule {
			c.addError(t.Pos(), "module %q not imported", moduleName)
			return Invalid
		}
		if modSym.Module == nil || modSym.Module.Scope == nil {
			c.addError(t.Pos(), "internal error: module %q has no scope", moduleName)
			return Invalid
		}
		typeSym := modSym.Module.Scope.Lookup(typeName)
		if typeSym == nil || typeSym.Kind != SymType {
			c.addError(t.Pos(), "type %q not found in module %q", typeName, moduleName)
			return Invalid
		}
		if !typeSym.IsPublic && modSym.Module.Name != c.currentModule {
			c.addError(t.Pos(), "type %q is not public in module %q", typeName, moduleName)
			return Invalid
		}
		return typeSym.Type

	case *ast.ListType:
		elemTypes := make([]Type, 0, len(t.ElementTypes))
		for _, etn := range t.ElementTypes {
			et := c.typeOfTypeNode(etn)
			elemTypes = append(elemTypes, et)
		}
		return &List{ElementTypes: elemTypes}
	case *ast.DictType:
		var keyType Type
		if t.KeyType != nil {
			keyType = c.typeOfTypeNode(t.KeyType)
		}
		valueType := c.typeOfTypeNode(t.ValueType)
		return &Dict{KeyType: keyType, ValueType: valueType}
	case *ast.OptionalType:
		innerType := c.typeOfTypeNode(t.Inner)
		return &Optional{Inner: innerType}

	case *ast.FuncType:
		paramTypes := make([]Type, 0, len(t.ParamTypes))
		for _, ptNode := range t.ParamTypes {
			pt := c.typeOfTypeNode(ptNode)
			paramTypes = append(paramTypes, pt)
		}
		res := c.typeOfTypeNode(t.Result)
		return &Func{
			ParamTypes: paramTypes,
			Result:     res,
		}

	case *ast.UnionType:
		var variants []Type
		for _, vtn := range t.Variants {
			vt := c.typeOfTypeNode(vtn)
			if IsInvalid(vt) {
				continue // skip invalid variants to avoid cascading errors
			}
			// Flatten nested unions
			if u, ok := vt.(*Union); ok {
				variants = append(variants, u.Variants...)
			} else {
				variants = append(variants, vt)
			}
		}
		if len(variants) == 0 {
			c.addError(t.Pos(), "union type must have at least one variant")
			return Invalid
		}
		return &Union{Variants: variants}

	case *ast.GenericInstanceType:
		return c.instantiateGenericType(t)

	case *ast.TypePackExpansion:
		if c.typeParamScope != nil {
			if concrete, ok := c.typeParamScope[t.Name]; ok {
				return concrete
			}
		}
		return &TypeVar{Name: t.Name}

	default:
		// Safety check
		c.addError(tn.Pos(), "unsupported type node %T", tn)
		return Invalid
	}
}

// ----- Generics Instantiation -----

func (c *Checker) instantiateGenericType(git *ast.GenericInstanceType) Type {
	if git.Name == "Future" {
		if len(git.TypeArgs) != 1 {
			c.addError(git.Pos(), "Future expects exactly 1 type argument, got %d", len(git.TypeArgs))
			return Invalid
		}
		inner := c.typeOfTypeNode(git.TypeArgs[0])
		if IsInvalid(inner) {
			return Invalid
		}
		return &Future{Inner: inner}
	}

	// Look up the generic struct by name from internal map
	if c.genericStructs != nil {
		if gs, ok := c.genericStructs[git.Name]; ok {
			return c.instantiateGenericStruct(gs, git)
		}
	}
	// Also check scope (generic structs are registered as symbols in Phase 1)
	if sym := c.scope.Lookup(git.Name); sym != nil && sym.Kind == SymType {
		if gs, ok := sym.Type.(*GenericStruct); ok {
			return c.instantiateGenericStruct(gs, git)
		}
	}
	c.addError(git.Pos(), "unknown generic type %q", git.Name)
	return Invalid
}

func (c *Checker) instantiateGenericStruct(gs *GenericStruct, git *ast.GenericInstanceType) Type {
	if len(git.TypeArgs) != len(gs.TypeParams) {
		c.addError(git.Pos(), "generic struct %q expects %d type arguments, got %d",
			gs.Decl.Name, len(gs.TypeParams), len(git.TypeArgs))
		return Invalid
	}

	// Resolve type arguments
	concreteArgs := make([]Type, len(git.TypeArgs))
	for i, ta := range git.TypeArgs {
		concreteArgs[i] = c.typeOfTypeNode(ta)
		if IsInvalid(concreteArgs[i]) {
			return Invalid
		}
	}

	// Build monomorphized name
	monoName := MonomorphKey(gs.Decl.Name, concreteArgs)

	// Check if already monomorphized
	if c.structTypes != nil {
		if existing, ok := c.structTypes[monoName]; ok {
			return existing
		}
	}

	// Build type param -> concrete type mapping
	mapping := make(map[string]Type, len(gs.TypeParams))
	for i, name := range gs.TypeParams {
		mapping[name] = concreteArgs[i]
	}

	// Save and set type param scope
	prevScope := c.typeParamScope
	c.typeParamScope = mapping
	defer func() { c.typeParamScope = prevScope }()

	// Build fields with substituted types
	st := gs.Decl
	fields := make([]Field, 0, len(st.Fields))
	for _, f := range st.Fields {
		fieldType := c.typeOfTypeNode(f.Type)
		if IsInvalid(fieldType) {
			continue
		}

		fieldMutable := st.IsMutable
		if f.IsMutable {
			fieldMutable = true
		}

		fields = append(fields, Field{
			Name:        f.Name,
			Type:        fieldType,
			IsPublic:    f.IsPublic,
			IsMutable:   fieldMutable,
			DefaultExpr: f.DefaultExpr,
		})
	}

	// Create the monomorphized struct type
	structType := &Struct{
		Name:            monoName,
		Fields:          fields,
		IsPublic:        st.IsPublic,
		IsMutable:       st.IsMutable,
		InstanceMethods: make(map[string]*Method),
		StaticMethods:   make(map[string]*Method),
	}

	if c.structTypes == nil {
		c.structTypes = make(map[string]*Struct)
	}
	c.structTypes[monoName] = structType

	if c.bindings != nil {
		c.bindings.MonomorphizedStructs[monoName] = structType
	}

	return structType
}

func (c *Checker) inferTypeArgs(gf *GenericFunc, argTypes []Type) ([]Type, bool) {
	typeParamSet := make(map[string]bool, len(gf.TypeParams))
	for _, tp := range gf.TypeParams {
		typeParamSet[tp] = true
	}

	inferred := make(map[string]Type)

	var unify func(tn ast.TypeNode, concrete Type) bool
	unify = func(tn ast.TypeNode, concrete Type) bool {
		switch t := tn.(type) {
		case *ast.SimpleType:
			if typeParamSet[t.Name] {
				if prev, exists := inferred[t.Name]; exists {
					return prev.equal(concrete)
				}
				inferred[t.Name] = concrete
				return true
			}
			return true
		case *ast.ListType:
			lt, ok := concrete.(*List)
			if !ok {
				return false
			}
			if len(t.ElementTypes) == 1 && len(lt.ElementTypes) >= 1 {
				return unify(t.ElementTypes[0], lt.ElementTypes[0])
			}
			return true
		case *ast.DictType:
			dt, ok := concrete.(*Dict)
			if !ok {
				return false
			}
			if t.ValueType != nil && dt.ValueType != nil {
				return unify(t.ValueType, dt.ValueType)
			}
			return true
		case *ast.OptionalType:
			ot, ok := concrete.(*Optional)
			if !ok {
				return false
			}
			return unify(t.Inner, ot.Inner)
		case *ast.FuncType:
			ft, ok := concrete.(*Func)
			if !ok {
				return false
			}
			if len(t.ParamTypes) != len(ft.ParamTypes) {
				return false
			}
			for i, pt := range t.ParamTypes {
				if !unify(pt, ft.ParamTypes[i]) {
					return false
				}
			}
			if t.Result != nil && ft.Result != nil {
				return unify(t.Result, ft.Result)
			}
			return true
		case *ast.GenericInstanceType:
			switch t.Name {
			case "Future":
				fut, ok := concrete.(*Future)
				if !ok {
					return false
				}
				if len(t.TypeArgs) == 1 {
					return unify(t.TypeArgs[0], fut.Inner)
				}
			}
			st, ok := concrete.(*Struct)
			if !ok {
				return true
			}
			_ = st
			return true
		default:
			return true
		}
	}

	params := gf.Decl.Params
	minLen := len(params)
	if minLen > len(argTypes) {
		minLen = len(argTypes)
	}

	for i := 0; i < minLen; i++ {
		if !unify(params[i].Type, argTypes[i]) {
			return nil, false
		}
	}

	result := make([]Type, len(gf.TypeParams))
	for i, tp := range gf.TypeParams {
		t, ok := inferred[tp]
		if !ok {
			return nil, false
		}
		result[i] = t
	}
	return result, true
}

func (c *Checker) instantiateGenericFunc(gf *GenericFunc, typeArgs []Type) (*Func, string) {
	if len(typeArgs) != len(gf.TypeParams) {
		c.addError(gf.Decl.Pos(), "generic function %q expects %d type arguments, got %d",
			gf.Decl.Name, len(gf.TypeParams), len(typeArgs))
		return nil, ""
	}

	monoName := MonomorphKey(gf.Decl.Name, typeArgs)

	// Build type param mapping
	mapping := make(map[string]Type, len(gf.TypeParams))
	for i, name := range gf.TypeParams {
		mapping[name] = typeArgs[i]
	}

	// Save and set type param scope
	prevScope := c.typeParamScope
	c.typeParamScope = mapping
	defer func() { c.typeParamScope = prevScope }()

	// Resolve parameter types with substitution
	fn := gf.Decl
	var paramTypes []Type
	for _, p := range fn.Params {
		pt := c.typeOfTypeNode(p.Type)
		paramTypes = append(paramTypes, pt)
	}
	retType := c.typeOfTypeNode(fn.Return)

	fnType := &Func{
		ParamTypes: paramTypes,
		Result:     retType,
	}

	// Register the monomorphized function in scope if not already done
	if c.monomorphized == nil {
		c.monomorphized = make(map[string]bool)
	}
	if !c.monomorphized[monoName] {
		c.monomorphized[monoName] = true

		// Insert monomorphized function as a new symbol
		monoDecl := &ast.FunDecl{
			Name:     monoName,
			NamePos:  fn.NamePos,
			Params:   fn.Params,
			Return:   fn.Return,
			Body:     fn.Body,
			IsPublic: fn.IsPublic,
		}

		_ = c.global.Insert(&Symbol{
			Name:     monoName,
			Kind:     SymFunc,
			Type:     fnType,
			Node:     monoDecl,
			IsPublic: fn.IsPublic,
		})

		if c.bindings != nil {
			c.bindings.MonomorphizedFuncs[monoName] = monoDecl
		}

		// Type-check the body with substituted types
		prevScopeCheck := c.scope
		c.scope = NewScope(c.global)
		prevRet := c.currentReturn
		c.currentReturn = retType

		for i, param := range fn.Params {
			_ = c.scope.Insert(&Symbol{
				Name: param.Name,
				Kind: SymVar,
				Type: paramTypes[i],
				Node: param,
			})
		}

		c.checkBlock(fn.Body)

		c.currentReturn = prevRet
		c.scope = prevScopeCheck
	}

	return fnType, monoName
}

// ----- Statements -----

func (c *Checker) checkBlock(b *ast.BlockStmt) {
	prevScope := c.scope
	c.scope = NewScope(prevScope)
	defer func() { c.scope = prevScope }()

	for _, st := range b.Stmts {
		c.checkStmt(st)
	}
}

func (c *Checker) checkStmt(s ast.Stmt) {
	switch st := s.(type) {
	case *ast.VarDeclStmt:
		c.checkVarDecl(st)
	case *ast.AssignStmt:
		c.checkAssign(st)
	case *ast.StructFieldAssignStmt:
		c.checkStructFieldAssign(st)
	case *ast.ExprStmt:
		_ = c.checkExpr(st.Expression)
	case *ast.IfStmt:
		c.checkIf(st)
	case *ast.WhileStmt:
		c.checkWhile(st)
	case *ast.ForStmt:
		c.checkFor(st)
	case *ast.ForEachStmt:
		c.checkForEach(st)
	case *ast.SwitchStmt:
		c.checkSwitch(st)
	case *ast.ThrowStmt:
		c.checkThrow(st)
	case *ast.TryStmt:
		c.checkTry(st)
	case *ast.BreakStmt:
		c.checkBreak(st)
	case *ast.ContinueStmt:
		c.checkContinue(st)
	case *ast.DeferStmt:
		c.checkDefer(st)
	case *ast.ReturnStmt:
		c.checkReturn(st)
	case *ast.BlockStmt:
		c.checkBlock(st)
	default:
		// Other statements
	}
}

func (c *Checker) checkTopLevelVar(s *ast.VarDeclStmt) {
	valType := c.checkExpr(s.Value)

	if s.Type == nil {
		if IsInvalid(valType) {
			c.addError(s.Pos(), "cannot infer type for variable %q", s.Name)
			return
		}
		if IsVoid(valType) {
			c.addError(s.Pos(), "cannot infer type void for variable %q", s.Name)
			return
		}
		sym := c.scope.Lookup(s.Name)
		if sym != nil {
			sym.Type = valType
		}
		return
	}

	typ := c.typeOfTypeNode(s.Type)

	if !c.assignable(typ, valType) {
		if iface, ok := typ.(*Interface); ok {
			c.addInterfaceSatisfactionError(s.Value.Pos(), valType, iface, s.Name)
		} else {
			c.addError(s.Pos(), "cannot assign expression of type %s to variable %q of type %s",
				valType.String(), s.Name, typ.String())
		}
	}
}

func (c *Checker) checkVarDecl(s *ast.VarDeclStmt) {
	valType := c.checkExpr(s.Value)

	var typ Type
	if s.Type == nil {
		if IsInvalid(valType) {
			c.addError(s.Pos(), "cannot infer type for variable %q", s.Name)
			return
		}
		if IsVoid(valType) {
			c.addError(s.Pos(), "cannot infer type void for variable %q", s.Name)
			return
		}
		typ = valType
	} else {
		typ = c.typeOfTypeNode(s.Type)
		if !c.assignable(typ, valType) {
			if iface, ok := typ.(*Interface); ok {
				c.addInterfaceSatisfactionError(s.Value.Pos(), valType, iface, s.Name)
			} else {
				c.addError(s.Pos(), "cannot assign expression of type %s to variable %q of type %s",
					valType.String(), s.Name, typ.String())
			}
			return
		}
	}

	if err := c.scope.Insert(&Symbol{
		Name: s.Name,
		Kind: SymVar,
		Type: typ,
		Node: s,
	}); err != nil {
		c.addError(s.Pos(), "variable %q: %v", s.Name, err)
	}
}

func (c *Checker) checkAssign(s *ast.AssignStmt) {
	sym := c.scope.Lookup(s.Name)
	if sym == nil {
		c.addError(s.Pos(), "undefined variable %q", s.Name)
		return
	}
	valType := c.checkExpr(s.Value)
	if !c.assignable(sym.Type, valType) {
		c.addError(s.Pos(), "cannot assign expression of type %s to variable %q of type %s",
			valType.String(), s.Name, sym.Type.String())
	}
}

func (c *Checker) checkStructFieldAssign(s *ast.StructFieldAssignStmt) {
	// Check that struct expression evaluates to a struct type
	structTypeExpr := c.checkExpr(s.Struct)
	structType, ok := structTypeExpr.(*Struct)
	if !ok {
		c.addError(s.Pos(), "cannot assign to field %q: target is not a struct, got %s", s.Field, structTypeExpr.String())
		return
	}

	// Find the field
	var field *Field
	for i := range structType.Fields {
		if structType.Fields[i].Name == s.Field {
			field = &structType.Fields[i]
			break
		}
	}

	if field == nil {
		c.addError(s.FieldPos, "struct %q has no field %q", structType.Name, s.Field)
		return
	}

	// Check field mutability
	// Field assignment is allowed only if the field is mutable
	// Field mutability is computed from struct default and field override
	if !field.IsMutable {
		c.addError(s.FieldPos, "cannot assign to immutable field %q", s.Field)
		return
	}

	// Check field visibility
	// Private fields can only be assigned from:
	// 1. Within methods of the same struct (via receiver)
	// 2. Within the same module where struct is defined
	if !field.IsPublic {
		// Check if we're in a method with matching receiver
		if c.currentReceiver == nil || c.currentReceiver.Name != structType.Name {
			// For same-module access, we allow it (visibility is per-module).
			// Cross-module checks would go here if we had module tracking for structs.
		}
	}

	// Check type compatibility
	valType := c.checkExpr(s.Value)
	if !c.assignable(field.Type, valType) {
		c.addError(s.Value.Pos(), "cannot assign expression of type %s to field %q of type %s",
			valType.String(), s.Field, field.Type.String())
	}
}

func (c *Checker) checkIf(s *ast.IfStmt) {
	condType := c.checkExpr(s.Cond)
	if !Equal(condType, Bool) {
		c.addError(s.Cond.Pos(), "if condition must be bool, got %s", condType.String())
	}
	c.checkBlock(s.Then)
	if s.Else != nil {
		c.checkStmt(s.Else)
	}
}

func (c *Checker) checkReturn(s *ast.ReturnStmt) {
	if c.currentReturn == nil || IsVoid(c.currentReturn) {
		if s.Result != nil {
			c.addError(s.Pos(), "function with void return cannot return a value")
		}
		return
	}

	if s.Result == nil {
		c.addError(s.Pos(), "function must return a value of type %s", c.currentReturn.String())
		return
	}

	resType := c.checkExpr(s.Result)
	if !c.assignable(c.currentReturn, resType) {
		c.addError(s.Pos(), "cannot use expression of type %s as return value of type %s",
			resType.String(), c.currentReturn.String())
	}
}

func (c *Checker) checkThrow(s *ast.ThrowStmt) {
	t := c.checkExpr(s.Expr)
	if Equal(t, ErrorType) {
		return
	}
	if _, ok := t.(*Struct); ok {
		if len(c.currentThrows) > 0 {
			for _, allowed := range c.currentThrows {
				if Equal(t, allowed) {
					return
				}
			}
			c.addError(s.Pos(), "thrown type %s is not declared in function's throws list", t.String())
			return
		}
		return
	}
	c.addError(s.Pos(), "throw expression must be of type error or a struct type, got %s", t.String())
}

func (c *Checker) isValidCatchType(t Type) bool {
	if Equal(t, ErrorType) {
		return true
	}
	if _, ok := t.(*Struct); ok {
		return true
	}
	return false
}

func (c *Checker) checkTry(s *ast.TryStmt) {
	c.checkBlock(s.Body)

	if len(s.Catches) > 1 {
		for _, clause := range s.Catches {
			catchType := c.typeOfTypeNode(clause.Type)

			if !c.isValidCatchType(catchType) {
				c.addError(clause.Type.Pos(), "catch variable must be of type error or a struct type, got %s", catchType.String())
				continue
			}

			prevScope := c.scope
			c.scope = NewScope(prevScope)

			_ = c.scope.Insert(&Symbol{
				Name: clause.VarName,
				Kind: SymVar,
				Type: catchType,
				Node: s,
			})

			c.checkBlock(clause.Body)
			c.scope = prevScope
		}
		return
	}

	if s.CatchBody != nil {
		catchType := c.typeOfTypeNode(s.CatchType)

		if !c.isValidCatchType(catchType) {
			c.addError(s.CatchType.Pos(), "catch variable must be of type error or a struct type, got %s", catchType.String())
			return
		}

		prevScope := c.scope
		c.scope = NewScope(prevScope)
		defer func() { c.scope = prevScope }()

		_ = c.scope.Insert(&Symbol{
			Name: s.CatchName,
			Kind: SymVar,
			Type: catchType,
			Node: s,
		})

		c.checkBlock(s.CatchBody)
	}
}

func (c *Checker) checkWhile(s *ast.WhileStmt) {
	condType := c.checkExpr(s.Cond)
	if !Equal(condType, Bool) {
		c.addError(s.Cond.Pos(), "while condition must be bool, got %s", condType.String())
	}
	c.loopDepth++
	c.checkBlock(s.Body)
	c.loopDepth--
}

func (c *Checker) checkFor(s *ast.ForStmt) {
	// Create a new scope for the for loop
	prevScope := c.scope
	c.scope = NewScope(prevScope)
	defer func() { c.scope = prevScope }()

	// Check init (if present)
	if s.Init != nil {
		c.checkStmt(s.Init)
	}

	// Check cond (if present)
	if s.Cond != nil {
		condType := c.checkExpr(s.Cond)
		if !Equal(condType, Bool) {
			c.addError(s.Cond.Pos(), "for condition must be bool, got %s", condType.String())
		}
	}

	// Check post (if present)
	if s.Post != nil {
		c.checkStmt(s.Post)
	}

	// Check body
	c.loopDepth++
	c.checkBlock(s.Body)
	c.loopDepth--
}

func (c *Checker) checkForEach(s *ast.ForEachStmt) {
	listType := c.checkExpr(s.ListExpr)
	list, ok := listType.(*List)
	if !ok {
		c.addError(s.ListExpr.Pos(), "foreach requires a list type, got %s", listType.String())
		return
	}

	// Create a new scope for the loop body
	prevScope := c.scope
	c.scope = NewScope(prevScope)
	defer func() { c.scope = prevScope }()

	// Determine the type of the loop variable
	var varType Type
	if len(list.ElementTypes) == 1 {
		varType = list.ElementTypes[0]
	} else {
		// Multiple element types - use any
		varType = Any
	}

	// Bind the loop variable
	if err := c.scope.Insert(&Symbol{
		Name: s.VarName,
		Kind: SymVar,
		Type: varType,
		Node: s,
	}); err != nil {
		c.addError(s.VarPos, "variable %q: %v", s.VarName, err)
	}

	// Check body
	c.loopDepth++
	c.checkBlock(s.Body)
	c.loopDepth--
}

func (c *Checker) checkBreak(s *ast.BreakStmt) {
	if c.loopDepth == 0 {
		c.addError(s.Pos(), "'break' is only allowed inside loops")
	}
}

func (c *Checker) checkContinue(s *ast.ContinueStmt) {
	if c.loopDepth == 0 {
		c.addError(s.Pos(), "'continue' is only allowed inside loops")
	}
}

func (c *Checker) checkSwitch(s *ast.SwitchStmt) {
	switchType := c.checkExpr(s.Expr)

	for _, clause := range s.Cases {
		patternType := c.checkExpr(clause.Pattern)
		if !c.assignable(switchType, patternType) && !c.assignable(patternType, switchType) {
			c.addError(clause.Pattern.Pos(), "switch case type %s is not compatible with %s", patternType.String(), switchType.String())
		}

		prevScope := c.scope
		c.scope = NewScope(prevScope)
		for _, st := range clause.Body {
			c.checkStmt(st)
		}
		c.scope = prevScope
	}

	if s.Default != nil {
		prevScope := c.scope
		c.scope = NewScope(prevScope)
		for _, st := range s.Default {
			c.checkStmt(st)
		}
		c.scope = prevScope
	}
}

func (c *Checker) checkDefer(s *ast.DeferStmt) {
	if s.Call == nil {
		c.addError(s.Pos(), "defer expects a call expression")
		return
	}
	_ = c.checkCall(s.Call)
}

// ----- Expressions -----

func (c *Checker) checkExpr(e ast.Expr) Type {
	var resultType Type
	switch ex := e.(type) {
	case *ast.IdentExpr:
		sym := c.scope.Lookup(ex.Name)
		if sym == nil {
			c.addError(ex.Pos(), "undefined identifier %q", ex.Name)
			resultType = Invalid
		} else {
			// Record binding for IR compiler
			if c.bindings != nil {
				c.bindings.Idents[ex] = sym
			}
			resultType = sym.Type
		}

	case *ast.IntLiteral:
		resultType = Int

	case *ast.FloatLiteral:
		resultType = Float

	case *ast.StringLiteral:
		resultType = String

	case *ast.InterpolatedString:
		for _, part := range ex.Parts {
			if exprPart, ok := part.(*ast.StringExprPart); ok {
				pt := c.checkExpr(exprPart.Expr)
				if IsInvalid(pt) {
					c.addError(exprPart.Expr.Pos(), "invalid interpolation expression")
				}
			}
		}
		resultType = String

	case *ast.BytesLiteral:
		resultType = Bytes

	case *ast.BoolLiteral:
		resultType = Bool

	case *ast.NoneLiteral:
		// none has type any? (we don't know which optional type)
		resultType = &Optional{Inner: Any}

	case *ast.SomeLiteral:
		innerType := c.checkExpr(ex.Value)
		resultType = &Optional{Inner: innerType}

	case *ast.StructLiteral:
		resultType = c.checkStructLiteral(ex)

	case *ast.ListLiteral:
		resultType = c.checkListLiteral(ex)
	case *ast.DictLiteral:
		resultType = c.checkDictLiteral(ex)

	case *ast.CallExpr:
		resultType = c.checkCall(ex)

	case *ast.IndexExpr:
		resultType = c.checkIndex(ex)

	case *ast.UnaryExpr:
		resultType = c.checkUnary(ex)

	case *ast.BinaryExpr:
		resultType = c.checkBinary(ex)

	case *ast.FuncLiteral:
		resultType = c.checkFuncLiteral(ex)

	case *ast.MemberExpr:
		resultType = c.checkMember(ex)

	case *ast.OptionalMemberExpr:
		resultType = c.checkOptionalMember(ex)

	case *ast.OptionalCallExpr:
		resultType = c.checkOptionalCall(ex)

	case *ast.AwaitExpr:
		resultType = c.checkAwait(ex)

	case *ast.ValuePackExpansion:
		sym := c.scope.Lookup(ex.Name)
		if sym == nil {
			c.addError(ex.Pos(), "undefined identifier %q", ex.Name)
			resultType = Invalid
		} else {
			resultType = sym.Type
		}

	default:
		resultType = Invalid
	}

	// Record expression type for IR compiler
	if c.bindings != nil && resultType != nil {
		c.bindings.ExprTypes[e] = resultType
	}

	return resultType
}

func (c *Checker) unwrapOptionalForChain(t Type, pos token.Position) Type {
	if opt, ok := t.(*Optional); ok {
		return opt.Inner
	}
	if Equal(t, Any) {
		return Any
	}
	c.addError(pos, "optional chaining requires optional value, got %s", t.String())
	return t
}

func (c *Checker) resolveValueMemberType(xType Type, name string, pos token.Position, xExpr ast.Expr) Type {
	tempMember := &ast.MemberExpr{X: xExpr, Name: name, NamePos: pos}
	if builtinMethodType := c.checkBuiltinMethod(tempMember, xType); builtinMethodType != nil {
		return builtinMethodType
	}

	if dictType, ok := xType.(*Dict); ok {
		if dictType.ValueType == nil {
			return Any
		}
		return dictType.ValueType
	}

	if interfaceType, ok := xType.(*Interface); ok {
		for _, method := range interfaceType.Methods {
			if method.Name == name {
				return &Func{
					ParamTypes: append([]Type{interfaceType}, method.ParamTypes...),
					Result:     method.Return,
				}
			}
		}
		c.addError(pos, "interface %s has no method %q", interfaceType.Name, name)
		return Invalid
	}

	structType, ok := xType.(*Struct)
	if !ok {
		c.addError(pos, "member access on non-struct type %s", xType.String())
		return Invalid
	}

	for _, field := range structType.Fields {
		if field.Name == name {
			return field.Type
		}
	}

	if structType.InstanceMethods != nil {
		if method, ok := structType.InstanceMethods[name]; ok {
			return &Func{
				ParamTypes: method.ParamTypes,
				Result:     method.Result,
			}
		}
	}

	c.addError(pos, "struct %s has no field or instance method %q", structType.Name, name)
	return Invalid
}

func (c *Checker) checkOptionalMember(m *ast.OptionalMemberExpr) Type {
	leftType := c.checkExpr(m.X)
	inner := c.unwrapOptionalForChain(leftType, m.Pos())
	memberType := c.resolveValueMemberType(inner, m.Name, m.NamePos, m.X)
	if IsInvalid(memberType) {
		return Invalid
	}
	return &Optional{Inner: memberType}
}

func (c *Checker) checkOptionalCall(call *ast.OptionalCallExpr) Type {
	checkArgs := func(paramTypes []Type, args []ast.Expr, pos token.Position) {
		if len(args) != len(paramTypes) {
			c.addError(pos, "function expects %d arguments, got %d", len(paramTypes), len(args))
		}
		limit := len(args)
		if len(paramTypes) < limit {
			limit = len(paramTypes)
		}
		for i := 0; i < len(args); i++ {
			argExpr := args[i]
			if namedArg, ok := argExpr.(*ast.NamedArg); ok {
				argExpr = namedArg.Value
			}
			argType := c.checkExpr(argExpr)
			if i < limit && !c.assignable(paramTypes[i], argType) {
				c.addError(argExpr.Pos(), "cannot use expression of type %s as argument %d of type %s",
					argType.String(), i+1, paramTypes[i].String())
			}
		}
	}

	if member, ok := call.Callee.(*ast.MemberExpr); ok {
		leftType := c.checkExpr(member.X)
		inner := c.unwrapOptionalForChain(leftType, call.Pos())
		memberType := c.resolveValueMemberType(inner, member.Name, member.NamePos, member.X)
		fnType, ok := memberType.(*Func)
		if !ok {
			c.addError(call.Pos(), "optional call requires callable member, got %s", memberType.String())
			for _, arg := range call.Args {
				if namedArg, ok := arg.(*ast.NamedArg); ok {
					_ = c.checkExpr(namedArg.Value)
				} else {
					_ = c.checkExpr(arg)
				}
			}
			return Invalid
		}

		params := fnType.ParamTypes
		if len(params) > 0 && c.assignable(params[0], inner) {
			params = params[1:]
		}
		checkArgs(params, call.Args, call.Pos())
		return &Optional{Inner: fnType.Result}
	}

	calleeType := c.checkExpr(call.Callee)
	innerCallee := c.unwrapOptionalForChain(calleeType, call.Pos())
	fnType, ok := innerCallee.(*Func)
	if !ok {
		c.addError(call.Pos(), "optional call requires callable value, got %s", innerCallee.String())
		for _, arg := range call.Args {
			if namedArg, ok := arg.(*ast.NamedArg); ok {
				_ = c.checkExpr(namedArg.Value)
			} else {
				_ = c.checkExpr(arg)
			}
		}
		return Invalid
	}

	checkArgs(fnType.ParamTypes, call.Args, call.Pos())
	return &Optional{Inner: fnType.Result}
}

func (c *Checker) checkMember(m *ast.MemberExpr) Type {
	// Case 1: moduleAlias.name
	if ident, ok := m.X.(*ast.IdentExpr); ok {
		sym := c.scope.Lookup(ident.Name)
		if sym != nil && sym.Kind == SymModule {
			// Look into imported module
			imported := sym.Module
			if imported == nil {
				c.addError(m.Pos(), "internal error: module symbol %q has nil Module", ident.Name)
				return Invalid
			}

			// Find public function in imported.Scope
			target := imported.Scope.Lookup(m.Name)
			if target == nil || target.Kind != SymFunc {
				c.addError(m.Pos(), "module %s has no function %q", imported.Name, m.Name)
				return Invalid
			}

			// Only allow public functions
			if !target.IsPublic {
				c.addError(m.Pos(), "function %q in module %s is not public", m.Name, imported.Name)
				return Invalid
			}

			// Record binding
			if c.bindings != nil {
				c.bindings.Members[m] = target
			}

			// Type of member access is function type
			return target.Type
		}

		// Case 2: Type.method (static method call)
		// Check if ident refers to a type (SymType)
		// If not found in current scope, try global scope
		if sym == nil || sym.Kind != SymType {
			sym = c.global.Lookup(ident.Name)
		}
		if sym != nil && sym.Kind == SymType {
			structType, ok := sym.Type.(*Struct)
			if !ok {
				c.addError(m.Pos(), "static methods can only be called on struct types, got %s", sym.Type.String())
				return Invalid
			}

			// Check struct visibility when accessed from outside its module
			// For same-module access, private structs are allowed.
			// For cross-module access (when accessing via imported module), struct must be public.
			// Since we're in the same module (same global scope), allow access to private structs.
			// Cross-module visibility checks would go here if we had module tracking for structs.

			// Look up static method
			if structType.StaticMethods != nil {
				if method, ok := structType.StaticMethods[m.Name]; ok {
					// Record binding for IR compiler
					if c.bindings != nil {
						// For static methods, ParamTypes already doesn't include receiver
						methodFuncType := &Func{
							ParamTypes: method.ParamTypes,
							Result:     method.Result,
						}
						c.bindings.Members[m] = &Symbol{
							Name: m.Name,
							Kind: SymFunc,
							Type: methodFuncType,
						}
					}
					// Return function type
					return &Func{
						ParamTypes: method.ParamTypes,
						Result:     method.Result,
					}
				}
			}

			// Check instance methods to provide better error message
			if structType.InstanceMethods != nil {
				if _, ok := structType.InstanceMethods[m.Name]; ok {
					c.addError(m.Pos(), "cannot call instance method %q on type %s, use a value instead", m.Name, structType.Name)
					return Invalid
				}
			}

			c.addError(m.Pos(), "type %s has no static method %q", structType.Name, m.Name)
			return Invalid
		}
	}

	// Case 3: value.field, value.method, or built-in method
	xType := c.checkExpr(m.X)

	// Check for built-in methods on basic types (list, string, bytes, etc.)
	if builtinMethodType := c.checkBuiltinMethod(m, xType); builtinMethodType != nil {
		return builtinMethodType
	}

	// Dict key access (after built-in methods)
	if dictType, ok := xType.(*Dict); ok {
		if lit, ok2 := m.X.(*ast.DictLiteral); ok2 {
			found := false
			for _, entry := range lit.Entries {
				if entry.Key == m.Name {
					found = true
					break
				}
			}
			if !found {
				c.addError(m.Pos(), "dict literal has no key %q", m.Name)
			}
		}
		if dictType.ValueType == nil {
			return Any
		}
		return dictType.ValueType
	}

	// Check if it's an interface type - method calls on interfaces are allowed
	if interfaceType, ok := xType.(*Interface); ok {
		// Find the method in the interface
		for _, method := range interfaceType.Methods {
			if method.Name == m.Name {
				// Return the method's function type
				return &Func{
					ParamTypes: append([]Type{interfaceType}, method.ParamTypes...), // Interface type as first param (receiver)
					Result:     method.Return,
				}
			}
		}
		c.addError(m.Pos(), "interface %s has no method %q", interfaceType.Name, m.Name)
		return Invalid
	}

	structType, ok := xType.(*Struct)
	if !ok {
		c.addError(m.Pos(), "member access on non-struct type %s", xType.String())
		return Invalid
	}

	// First check for fields (fields take precedence)
	for _, field := range structType.Fields {
		if field.Name == m.Name {
			// Check field visibility
			// Private fields can be accessed:
			// 1. From within methods of the same struct (via receiver) - always allowed
			// 2. From within the same module where struct is defined - allowed
			// 3. From other modules - only public fields are accessible
			// For now, we allow same-module access by default (single-module programs).
			// Cross-module visibility is enforced when accessing imported structs.
			if !field.IsPublic {
				// Check if we're in a method with matching receiver (always allowed)
				if c.currentReceiver == nil || c.currentReceiver.Name != structType.Name {
					// For same-module access, we allow it (visibility is per-module).
					// Cross-module checks would go here if we had module tracking for structs.
					// For now, allow private field access within the same module.
				}
			}

			// Record binding for IR compiler
			if c.bindings != nil {
				// Store field index for efficient access
				c.bindings.Members[m] = &Symbol{
					Name: m.Name,
					Kind: SymVar, // Field is like a variable
					Type: field.Type,
				}
			}
			return field.Type
		}
	}

	// Then check for instance methods
	if structType.InstanceMethods != nil {
		if method, ok := structType.InstanceMethods[m.Name]; ok {
			if c.bindings != nil {
				methodFuncType := &Func{
					ParamTypes: method.ParamTypes,
					Result:     method.Result,
				}
				c.bindings.Members[m] = &Symbol{
					Name: m.Name,
					Kind: SymFunc,
					Type: methodFuncType,
					Node: method.Decl,
				}
			}
			// Return function type (includes receiver as first parameter for instance methods)
			return &Func{
				ParamTypes: method.ParamTypes,
				Result:     method.Result,
			}
		}
	}

	// Check static methods to provide better error message
	if structType.StaticMethods != nil {
		if _, ok := structType.StaticMethods[m.Name]; ok {
			c.addError(m.Pos(), "cannot call static method %q on value of type %s, use the type name instead", m.Name, structType.Name)
			return Invalid
		}
	}

	c.addError(m.Pos(), "struct %s has no field or instance method %q", structType.Name, m.Name)
	return Invalid
}

// checkBuiltinMethod checks if a member access is a built-in method call on a basic type.
// Returns the method's function type if found, nil otherwise.
func (c *Checker) checkBuiltinMethod(m *ast.MemberExpr, receiverType Type) Type {
	// Convert type to builtin TypeKind
	var typeKind builtins.TypeKind
	var found bool
	var dictType *Dict

	switch t := receiverType.(type) {
	case *Basic:
		typeKind, found = builtins.TypeKindFromString(t.Name)
	case *List:
		typeKind, found = builtins.TypeList, true
	case *Dict:
		typeKind, found = builtins.TypeDict, true
		dictType = t
	default:
		return nil
	}

	if !found {
		return nil
	}

	// Look up built-in method
	methodBuiltin := builtins.LookupMethod(typeKind, m.Name)
	if methodBuiltin == nil {
		return nil
	}
	methodMeta := methodBuiltin.Meta

	// Build function type for the method
	// For built-in methods, the receiver is the first parameter
	var paramTypes []Type
	var resultType Type
	if dictType != nil {
		keyType := dictType.keyType()
		valueType := dictType.ValueType
		if valueType == nil {
			valueType = Any
		}
		switch m.Name {
		case "length":
			paramTypes = []Type{dictType}
			resultType = Int
		case "keys":
			paramTypes = []Type{dictType}
			resultType = &List{ElementTypes: []Type{keyType}}
		case "values":
			paramTypes = []Type{dictType}
			resultType = &List{ElementTypes: []Type{valueType}}
		case "has":
			paramTypes = []Type{dictType, keyType}
			resultType = Bool
		case "get":
			paramTypes = []Type{dictType, keyType}
			resultType = &Optional{Inner: valueType}
		case "set":
			paramTypes = []Type{dictType, keyType, valueType}
			resultType = Void
		case "remove":
			paramTypes = []Type{dictType, keyType}
			resultType = Bool
		default:
			return nil
		}
	} else {
		paramTypes = make([]Type, len(methodMeta.Params))
		for i, p := range methodMeta.Params {
			paramTypes[i] = c.typeRefToType(p)
		}
		resultType = c.typeRefToType(methodMeta.Result)
	}

	// Record binding for IR compiler
	if c.bindings != nil {
		methodFuncType := &Func{
			ParamTypes: paramTypes,
			Result:     resultType,
		}
		// Use a special marker symbol for built-in methods (no Node field)
		c.bindings.Members[m] = &Symbol{
			Name: m.Name,
			Kind: SymFunc,
			Type: methodFuncType,
			// Node is nil for built-in methods - IR compiler will detect this
		}
	}

	// Return function type (includes receiver as first parameter)
	return &Func{
		ParamTypes: paramTypes,
		Result:     resultType,
	}
}

// typeRefToType converts a builtin TypeRef to a types.Type.
func (c *Checker) typeRefToType(ref builtins.TypeRef) Type {
	switch ref.Kind {
	case builtins.TypeInt:
		return Int
	case builtins.TypeFloat:
		return Float
	case builtins.TypeString:
		return String
	case builtins.TypeBool:
		return Bool
	case builtins.TypeVoid:
		return Void
	case builtins.TypeAny:
		return Any
	case builtins.TypeError:
		return ErrorType
	case builtins.TypeBytes:
		return Bytes
	case builtins.TypeList:
		// Convert element types
		elemTypes := make([]Type, len(ref.Elem))
		for i, elem := range ref.Elem {
			elemTypes[i] = c.typeRefToType(elem)
		}
		return &List{ElementTypes: elemTypes}
	case builtins.TypeDict:
		var valueType Type = Any
		if len(ref.Elem) > 0 {
			valueType = c.typeRefToType(ref.Elem[0])
		}
		return &Dict{ValueType: valueType}
	case builtins.TypeUnion:
		variants := make([]Type, 0, len(ref.Elem))
		for _, elem := range ref.Elem {
			vt := c.typeRefToType(elem)
			if !IsInvalid(vt) {
				variants = append(variants, vt)
			}
		}
		if len(variants) == 0 {
			return Invalid
		}
		return &Union{Variants: variants}
	default:
		return Invalid
	}
}

func (c *Checker) checkListLiteral(lit *ast.ListLiteral) Type {
	// List type is determined by elements: list<T1, T2, ...>
	elemTypes := []Type{}
	for _, e := range lit.Elements {
		t := c.checkExpr(e)
		if IsInvalid(t) {
			continue
		}
		// Remove duplicate types
		already := false
		for _, et := range elemTypes {
			if Equal(et, t) {
				already = true
				break
			}
		}
		if !already {
			elemTypes = append(elemTypes, t)
		}
	}
	return &List{ElementTypes: elemTypes}
}

func (c *Checker) checkDictLiteral(lit *ast.DictLiteral) Type {
	var valueTypes []Type
	for _, entry := range lit.Entries {
		t := c.checkExpr(entry.Value)
		if IsInvalid(t) {
			continue
		}
		already := false
		for _, vt := range valueTypes {
			if Equal(vt, t) {
				already = true
				break
			}
		}
		if !already {
			valueTypes = append(valueTypes, t)
		}
	}

	var valueType Type
	switch len(valueTypes) {
	case 0:
		valueType = Any
	case 1:
		valueType = valueTypes[0]
	default:
		valueType = &Union{Variants: valueTypes}
	}

	return &Dict{ValueType: valueType}
}

func (c *Checker) checkStructLiteral(lit *ast.StructLiteral) Type {
	// Handle generic struct literals: Box<int>{value = 10}
	if len(lit.TypeArgs) > 0 {
		var gs *GenericStruct
		if c.genericStructs != nil {
			gs = c.genericStructs[lit.TypeName]
		}
		if gs == nil {
			if sym := c.scope.Lookup(lit.TypeName); sym != nil && sym.Kind == SymType {
				gs, _ = sym.Type.(*GenericStruct)
			}
		}
		if gs == nil {
			c.addError(lit.Pos(), "unknown generic struct type %q", lit.TypeName)
			return Invalid
		}
		git := &ast.GenericInstanceType{
			Name:     lit.TypeName,
			NamePos:  lit.TypeNamePos,
			TypeArgs: lit.TypeArgs,
		}
		structType, ok := c.instantiateGenericStruct(gs, git).(*Struct)
		if !ok {
			return Invalid
		}
		return c.checkStructLiteralFields(lit, structType)
	}

	// Look up struct type
	var structType *Struct
	if c.structTypes != nil {
		if st, ok := c.structTypes[lit.TypeName]; ok {
			structType = st
		}
	}
	if structType == nil {
		// Check in scope (for imported structs)
		if sym := c.scope.Lookup(lit.TypeName); sym != nil && sym.Kind == SymType {
			if st, ok := sym.Type.(*Struct); ok {
				structType = st
			}
		}
	}

	if structType == nil {
		c.addError(lit.Pos(), "unknown struct type %q", lit.TypeName)
		return Invalid
	}

	return c.checkStructLiteralFields(lit, structType)
}

func (c *Checker) checkStructLiteralFields(lit *ast.StructLiteral, structType *Struct) Type {
	// Build map of provided fields
	provided := make(map[string]bool)
	fieldMap := make(map[string]Field)
	for _, f := range structType.Fields {
		fieldMap[f.Name] = f
	}

	// Check each field initialization
	for _, fieldInit := range lit.Fields {
		if provided[fieldInit.Name] {
			c.addError(fieldInit.Pos(), "duplicate field %q in struct literal", fieldInit.Name)
			continue
		}

		field, ok := fieldMap[fieldInit.Name]
		if !ok {
			c.addError(fieldInit.Pos(), "unknown field %q in struct %q", fieldInit.Name, structType.Name)
			continue
		}

		provided[fieldInit.Name] = true

		// Check field value type
		valueType := c.checkExpr(fieldInit.Value)
		if !c.assignable(field.Type, valueType) {
			c.addError(fieldInit.Value.Pos(), "cannot assign expression of type %s to field %q of type %s",
				valueType.String(), fieldInit.Name, field.Type.String())
		}
	}

	// Check all required fields are provided (fields without defaults)
	for _, field := range structType.Fields {
		if !provided[field.Name] {
			if field.DefaultExpr == nil {
				c.addError(lit.Pos(), "missing required field %q in struct literal", field.Name)
			}
		}
	}

	return structType
}

// isCompileTimeConstant checks if an expression is a compile-time constant.
// Compile-time constants are literals (int, float, string, bool, bytes) or constant composite literals.
// They must not reference variables, functions, or any runtime values.
func (c *Checker) isCompileTimeConstant(e ast.Expr) bool {
	switch expr := e.(type) {
	case *ast.IntLiteral, *ast.FloatLiteral, *ast.StringLiteral, *ast.BytesLiteral, *ast.BoolLiteral:
		return true
	case *ast.NoneLiteral, *ast.SomeLiteral:
		// Optional literals are compile-time constants if their inner value is
		if some, ok := expr.(*ast.SomeLiteral); ok {
			return c.isCompileTimeConstant(some.Value)
		}
		return true
	case *ast.ListLiteral:
		// List literals are compile-time constants if all elements are compile-time constants
		for _, elem := range expr.Elements {
			if !c.isCompileTimeConstant(elem) {
				return false
			}
		}
		return true
	case *ast.DictLiteral:
		// Dict literals are compile-time constants if all values are compile-time constants
		for _, entry := range expr.Entries {
			if !c.isCompileTimeConstant(entry.Value) {
				return false
			}
		}
		return true
	case *ast.UnaryExpr:
		// Unary operations on constants are compile-time constants
		return c.isCompileTimeConstant(expr.X)
	case *ast.BinaryExpr:
		// Binary operations on constants are compile-time constants
		return c.isCompileTimeConstant(expr.Left) && c.isCompileTimeConstant(expr.Right)
	default:
		// All other expressions (identifiers, calls, member access, etc.) are not compile-time constants
		return false
	}
}

func (c *Checker) checkGenericCall(call *ast.CallExpr) Type {
	ident, ok := call.Callee.(*ast.IdentExpr)
	if !ok {
		c.addError(call.Pos(), "generic type arguments are only supported on named function calls")
		return Invalid
	}

	sym := c.scope.Lookup(ident.Name)
	if sym == nil {
		c.addError(ident.Pos(), "undefined function %q", ident.Name)
		return Invalid
	}

	gf, ok := sym.Type.(*GenericFunc)
	if !ok {
		c.addError(call.Pos(), "function %q is not generic", ident.Name)
		return Invalid
	}

	// Resolve type arguments
	concreteArgs := make([]Type, len(call.TypeArgs))
	for i, ta := range call.TypeArgs {
		concreteArgs[i] = c.typeOfTypeNode(ta)
		if IsInvalid(concreteArgs[i]) {
			return Invalid
		}
	}

	fnType, monoName := c.instantiateGenericFunc(gf, concreteArgs)
	if fnType == nil {
		return Invalid
	}

	// Store the monomorphized name in bindings for the compiler
	if c.bindings != nil {
		var monoNode *ast.FunDecl
		if decl, ok := c.bindings.MonomorphizedFuncs[monoName]; ok {
			monoNode = decl
		}
		c.bindings.Idents[ident] = &Symbol{
			Name: monoName,
			Kind: SymFunc,
			Type: fnType,
			Node: monoNode,
		}
	}

	// Type-check arguments (positional only for now)
	if len(call.Args) != len(fnType.ParamTypes) {
		c.addError(call.Pos(), "function %s expects %d arguments, got %d",
			ident.Name, len(fnType.ParamTypes), len(call.Args))
		return fnType.Result
	}

	for i, arg := range call.Args {
		argType := c.checkExpr(arg)
		paramType := fnType.ParamTypes[i]
		if !c.assignable(paramType, argType) {
			c.addError(arg.Pos(), "cannot use expression of type %s as argument %d of type %s",
				argType.String(), i+1, paramType.String())
		}
	}

	return fnType.Result
}

func (c *Checker) checkInferredGenericCall(call *ast.CallExpr, gf *GenericFunc) Type {
	argTypes := make([]Type, len(call.Args))
	for i, arg := range call.Args {
		argTypes[i] = c.checkExpr(arg)
	}

	inferredArgs, ok := c.inferTypeArgs(gf, argTypes)
	if !ok {
		c.addError(call.Pos(), "cannot infer type arguments for generic function %q", gf.Decl.Name)
		return Invalid
	}

	fnType, monoName := c.instantiateGenericFunc(gf, inferredArgs)
	if fnType == nil {
		return Invalid
	}

	if c.bindings != nil {
		if ident, ok := call.Callee.(*ast.IdentExpr); ok {
			var monoNode *ast.FunDecl
			if decl, ok2 := c.bindings.MonomorphizedFuncs[monoName]; ok2 {
				monoNode = decl
			}
			c.bindings.Idents[ident] = &Symbol{
				Name: monoName,
				Kind: SymFunc,
				Type: fnType,
				Node: monoNode,
			}
		}
	}

	if len(call.Args) != len(fnType.ParamTypes) {
		c.addError(call.Pos(), "function %s expects %d arguments, got %d",
			gf.Decl.Name, len(fnType.ParamTypes), len(call.Args))
		return fnType.Result
	}

	for i, argType := range argTypes {
		paramType := fnType.ParamTypes[i]
		if !c.assignable(paramType, argType) {
			c.addError(call.Args[i].Pos(), "cannot use expression of type %s as argument %d of type %s",
				argType.String(), i+1, paramType.String())
		}
	}

	return fnType.Result
}

func (c *Checker) checkCall(call *ast.CallExpr) Type {
	// Handle generic function calls: ident<Type>(args)
	if len(call.TypeArgs) > 0 {
		return c.checkGenericCall(call)
	}

	calleeType := c.checkExpr(call.Callee)

	if gf, isGeneric := calleeType.(*GenericFunc); isGeneric {
		return c.checkInferredGenericCall(call, gf)
	}

	fnType, ok := calleeType.(*Func)
	if !ok {
		c.addError(call.Pos(), "called expression is not a function, got %s", calleeType.String())
		return Invalid
	}

	// Try to get function declaration if callee is an identifier
	var fnDecl *ast.FunDecl
	if ident, ok := call.Callee.(*ast.IdentExpr); ok {
		if sym := c.scope.Lookup(ident.Name); sym != nil && sym.Kind == SymFunc {
			if fn, ok2 := sym.Node.(*ast.FunDecl); ok2 {
				fnDecl = fn
			}
		}
	}

	// If callee is a module function or method (member expr), we can also get its FunDecl via bindings.
	if fnDecl == nil && c.bindings != nil {
		if member, ok := call.Callee.(*ast.MemberExpr); ok {
			if sym, ok2 := c.bindings.Members[member]; ok2 && sym.Kind == SymFunc {
				if fn, ok3 := sym.Node.(*ast.FunDecl); ok3 {
					fnDecl = fn
				}
			}
			// For methods, the type checking is already done in checkMember
			// which returns the method's function type. No need to find FunDecl here.
		}
	}

	// If we don't have a function declaration (builtin or function value)
	if fnDecl == nil {
		// Check if this is a builtin call (by identifier name)
		var builtinParamNames []string
		var builtinName string
		if ident, ok := call.Callee.(*ast.IdentExpr); ok {
			if builtin := builtins.LookupByName(ident.Name); builtin != nil {
				builtinParamNames = builtin.Meta.ParamNames
				builtinName = ident.Name
			}
		}

		// If this is a builtin, handle named arguments
		if builtinParamNames != nil {
			nParams := len(builtinParamNames)
			paramNames := builtinParamNames
			provided := make([]bool, nParams)
			argExprs := make([]ast.Expr, nParams)
			positionalIndex := 0
			seenNamed := false

			for _, arg := range call.Args {
				if named, ok := arg.(*ast.NamedArg); ok {
					seenNamed = true
					// Find parameter index by name
					idx := -1
					for i, name := range paramNames {
						if name == named.Name {
							idx = i
							break
						}
					}
					if idx == -1 {
						c.addError(named.Pos(), "function %s has no parameter named %q", builtinName, named.Name)
						continue
					}
					if provided[idx] {
						c.addError(named.Pos(), "parameter %q specified multiple times", named.Name)
						continue
					}
					provided[idx] = true
					argExprs[idx] = named.Value
				} else {
					// Positional argument
					if seenNamed {
						c.addError(arg.Pos(), "positional arguments cannot follow named arguments")
						continue
					}
					if positionalIndex >= nParams {
						c.addError(arg.Pos(), "too many arguments in call to %s", builtinName)
						continue
					}
					provided[positionalIndex] = true
					argExprs[positionalIndex] = arg
					positionalIndex++
				}
			}

			// Check for missing required parameters (builtins don't have defaults)
			for i := 0; i < nParams; i++ {
				if !provided[i] {
					c.addError(call.Pos(), "missing argument for required parameter %q", paramNames[i])
				}
			}

			// Type-check provided arguments
			for i := 0; i < nParams; i++ {
				if provided[i] {
					argType := c.checkExpr(argExprs[i])
					paramType := fnType.ParamTypes[i]
					if !c.assignable(paramType, argType) {
						c.addError(argExprs[i].Pos(), "cannot use expression of type %s as argument %d (%q) of type %s",
							argType.String(), i+1, paramNames[i], paramType.String())
					}
				}
			}
			return fnType.Result
		}

		// Not a builtin - check if this is a method call (MemberExpr with receiver parameter)
		// For method calls (both instance methods and built-in methods), the receiver is implicitly the first argument
		var effectiveArgs []ast.Expr
		if member, ok := call.Callee.(*ast.MemberExpr); ok {
			// Check if this looks like a method call
			// (function type has one more parameter than call.Args)
			if len(fnType.ParamTypes) == len(call.Args)+1 {
				// This is likely a method call (instance method or built-in method) - prepend receiver
				effectiveArgs = append([]ast.Expr{member.X}, call.Args...)
			} else {
				effectiveArgs = call.Args
			}
		} else {
			effectiveArgs = call.Args
		}

		// Check for named arguments (not allowed for function values)
		for _, arg := range effectiveArgs {
			if _, ok := arg.(*ast.NamedArg); ok {
				c.addError(call.Pos(), "named arguments are only allowed when calling a function by name")
				return fnType.Result
			}
		}
		// Positional-only logic with effective arguments
		if len(effectiveArgs) != len(fnType.ParamTypes) {
			c.addError(call.Pos(), "function expects %d arguments, got %d",
				len(fnType.ParamTypes), len(effectiveArgs))
			return fnType.Result
		}
		for i, arg := range effectiveArgs {
			argType := c.checkExpr(arg)
			paramType := fnType.ParamTypes[i]
			if !c.assignable(paramType, argType) {
				c.addError(arg.Pos(), "cannot use expression of type %s as argument %d of type %s",
					argType.String(), i+1, paramType.String())
			}
		}
		return fnType.Result
	}

	// We have a function declaration - handle named arguments and defaults
	// For instance methods, the receiver is the first parameter
	// For static methods, there's no receiver parameter
	isInstanceMethod := fnDecl.Receiver != nil && fnDecl.Receiver.Kind == ast.ReceiverInstance

	// Build effective arguments list (prepend receiver for instance methods)
	var effectiveArgs []ast.Expr
	if isInstanceMethod {
		// For instance methods, prepend the receiver expression
		if member, ok := call.Callee.(*ast.MemberExpr); ok {
			effectiveArgs = append([]ast.Expr{member.X}, call.Args...)
		} else {
			c.addError(call.Pos(), "internal error: instance method call callee is not a MemberExpr")
			return fnType.Result
		}
	} else {
		effectiveArgs = call.Args
	}

	// Build parameter names (include receiver for instance methods)
	var paramNames []string
	if isInstanceMethod {
		paramNames = append([]string{fnDecl.Receiver.Name}, func() []string {
			names := make([]string, len(fnDecl.Params))
			for i, p := range fnDecl.Params {
				names[i] = p.Name
			}
			return names
		}()...)
	} else {
		paramNames = make([]string, len(fnDecl.Params))
		for i, param := range fnDecl.Params {
			paramNames[i] = param.Name
		}
	}

	nParams := len(paramNames) // Includes receiver for instance methods
	paramTypes := fnType.ParamTypes
	hasDefault := make([]bool, nParams)
	if isInstanceMethod {
		hasDefault[0] = false // Receiver has no default
		for i, param := range fnDecl.Params {
			hasDefault[i+1] = param.Default != nil
		}
	} else {
		for i, param := range fnDecl.Params {
			hasDefault[i] = param.Default != nil
		}
	}

	// Process call arguments
	provided := make([]bool, nParams)
	argExprs := make([]ast.Expr, nParams)
	positionalIndex := 0
	seenNamed := false

	for _, arg := range effectiveArgs {
		if named, ok := arg.(*ast.NamedArg); ok {
			seenNamed = true
			// Find parameter index by name
			idx := -1
			for i, name := range paramNames {
				if name == named.Name {
					idx = i
					break
				}
			}
			if idx == -1 {
				c.addError(named.Pos(), "function %s has no parameter named %q", fnDecl.Name, named.Name)
				continue
			}
			if provided[idx] {
				c.addError(named.Pos(), "parameter %q specified multiple times", named.Name)
				continue
			}
			provided[idx] = true
			argExprs[idx] = named.Value
		} else {
			// Positional argument
			if seenNamed {
				c.addError(arg.Pos(), "positional arguments cannot follow named arguments")
				continue
			}
			if positionalIndex >= nParams {
				c.addError(arg.Pos(), "too many arguments in call to %s", fnDecl.Name)
				continue
			}
			provided[positionalIndex] = true
			argExprs[positionalIndex] = arg
			positionalIndex++
		}
	}

	// Check for missing required parameters
	for i := 0; i < nParams; i++ {
		if !provided[i] {
			// For instance methods, i==0 is the receiver (always required, no default)
			if isInstanceMethod && i == 0 {
				c.addError(call.Pos(), "missing receiver argument for instance method call")
			} else if !hasDefault[i] {
				c.addError(call.Pos(), "missing argument for required parameter %q", paramNames[i])
			}
			// If hasDefault[i], the default will be used (no error)
		}
	}

	// Type-check provided arguments
	for i := 0; i < nParams; i++ {
		if provided[i] {
			argType := c.checkExpr(argExprs[i])
			paramType := paramTypes[i]
			if !c.assignable(paramType, argType) {
				c.addError(argExprs[i].Pos(), "cannot use expression of type %s as argument %d (%q) of type %s",
					argType.String(), i+1, paramNames[i], paramType.String())
			}
		}
	}

	return fnType.Result
}

func (c *Checker) checkIndex(idx *ast.IndexExpr) Type {
	xType := c.checkExpr(idx.X)

	// Support indexing for bytes (returns int)
	if Equal(xType, Bytes) {
		indexType := c.checkExpr(idx.Index)
		if !Equal(indexType, Int) {
			c.addError(idx.Index.Pos(), "index expression must be int, got %s", indexType.String())
		}
		return Int // bytes[index] returns int (byte value 0-255)
	}

	// Support indexing for dict (returns value type)
	if dictType, ok := xType.(*Dict); ok {
		indexType := c.checkExpr(idx.Index)
		if !Equal(indexType, String) {
			c.addError(idx.Index.Pos(), "dict index must be string, got %s", indexType.String())
		}
		if dictType.ValueType == nil {
			return Any
		}
		return dictType.ValueType
	}

	// Support indexing for lists
	listType, ok := xType.(*List)
	if !ok {
		c.addError(idx.Pos(), "indexing is only supported for list, bytes, or dict types, got %s", xType.String())
		return Invalid
	}

	indexType := c.checkExpr(idx.Index)
	if !Equal(indexType, Int) {
		c.addError(idx.Index.Pos(), "index expression must be int, got %s", indexType.String())
	}

	// If only one element type, return it
	if len(listType.ElementTypes) == 1 {
		return listType.ElementTypes[0]
	}
	// If multiple variants, return any (dynamic typing)
	return Any
}

func (c *Checker) checkUnary(u *ast.UnaryExpr) Type {
	xType := c.checkExpr(u.X)

	// Check if operand is a union (restrict unions from unary operations)
	_, xIsUnion := xType.(*Union)

	switch u.Op {
	case token.Bang:
		if xIsUnion {
			c.addError(u.Pos(), "operator ! does not support union types, got %s", xType.String())
			return Invalid
		}
		if !Equal(xType, Bool) {
			c.addError(u.Pos(), "operator ! expects bool, got %s", xType.String())
			return Invalid
		}
		return Bool
	case token.Minus:
		if xIsUnion {
			c.addError(u.Pos(), "unary - does not support union types, got %s", xType.String())
			return Invalid
		}
		if Equal(xType, Int) {
			return Int
		}
		if Equal(xType, Float) {
			return Float
		}
		c.addError(u.Pos(), "unary - expects int or float, got %s", xType.String())
		return Invalid
	default:
		c.addError(u.Pos(), "unsupported unary operator %s", u.Op)
		return Invalid
	}
}

func (c *Checker) checkBinary(b *ast.BinaryExpr) Type {
	left := c.checkExpr(b.Left)
	right := c.checkExpr(b.Right)

	// Special-case '+' to allow string concatenation.
	if b.Op == token.Plus {
		if Equal(left, String) && Equal(right, String) {
			return String
		}
		leftIsInt := Equal(left, Int)
		leftIsFloat := Equal(left, Float)
		rightIsInt := Equal(right, Int)
		rightIsFloat := Equal(right, Float)
		if (leftIsInt || leftIsFloat) && (rightIsInt || rightIsFloat) {
			if leftIsFloat || rightIsFloat {
				return Float
			}
			return Int
		}
		c.addError(b.Pos(), "operator '+' is not defined for types %s and %s",
			left.String(), right.String())
		return Invalid
	}

	// Check if either operand is a union (restrict unions from most operations)
	_, leftIsUnion := left.(*Union)
	_, rightIsUnion := right.(*Union)

	switch b.Op {
	case token.Minus, token.Star, token.Slash, token.Percent:
		if leftIsUnion || rightIsUnion {
			c.addError(b.Pos(), "operator %s does not support union types, got (%s, %s)",
				b.Op, left.String(), right.String())
			return Invalid
		}
		// Support int/int -> int, float/float -> float, int/float or float/int -> float
		leftIsInt := Equal(left, Int)
		leftIsFloat := Equal(left, Float)
		rightIsInt := Equal(right, Int)
		rightIsFloat := Equal(right, Float)

		if !(leftIsInt || leftIsFloat) || !(rightIsInt || rightIsFloat) {
			c.addError(b.Pos(), "operator %s expects numeric types (int or float), got (%s, %s)",
				b.Op, left.String(), right.String())
			return Invalid
		}
		// If either operand is float, result is float; otherwise int
		if leftIsFloat || rightIsFloat {
			return Float
		}
		return Int

	case token.Lt, token.LtEq, token.Gt, token.GtEq:
		if leftIsUnion || rightIsUnion {
			c.addError(b.Pos(), "operator %s does not support union types, got (%s, %s)",
				b.Op, left.String(), right.String())
			return Invalid
		}
		// Support comparisons between int and float
		leftIsInt := Equal(left, Int)
		leftIsFloat := Equal(left, Float)
		rightIsInt := Equal(right, Int)
		rightIsFloat := Equal(right, Float)

		if !(leftIsInt || leftIsFloat) || !(rightIsInt || rightIsFloat) {
			c.addError(b.Pos(), "operator %s expects numeric types (int or float), got (%s, %s)",
				b.Op, left.String(), right.String())
			return Invalid
		}
		return Bool

	case token.Eq, token.NotEq:
		// Equality is allowed between identical union types or with any
		if leftIsUnion || rightIsUnion {
			if !(Equal(left, right) || Equal(left, Any) || Equal(right, Any)) {
				c.addError(b.Pos(), "operator %s expects compatible types, got (%s, %s)",
					b.Op, left.String(), right.String())
				return Invalid
			}
			return Bool
		}
		// Allow T? == any? (for == none comparisons)
		leftOpt, leftIsOpt := left.(*Optional)
		rightOpt, rightIsOpt := right.(*Optional)
		if leftIsOpt && rightIsOpt {
			if Equal(leftOpt.Inner, Any) || Equal(rightOpt.Inner, Any) {
				return Bool
			}
		}
		// Simplified: types must match or one of them must be any
		if !(Equal(left, right) || Equal(left, Any) || Equal(right, Any)) {
			c.addError(b.Pos(), "operator %s expects compatible types, got (%s, %s)",
				b.Op, left.String(), right.String())
			return Invalid
		}
		return Bool

	case token.AndAnd, token.OrOr:
		if leftIsUnion || rightIsUnion {
			c.addError(b.Pos(), "logical operator %s does not support union types, got (%s, %s)",
				b.Op, left.String(), right.String())
			return Invalid
		}
		if !Equal(left, Bool) || !Equal(right, Bool) {
			c.addError(b.Pos(), "logical operator %s expects (bool, bool), got (%s, %s)",
				b.Op, left.String(), right.String())
			return Invalid
		}
		return Bool

	default:
		c.addError(b.Pos(), "unsupported binary operator %s", b.Op)
		return Invalid
	}
}

// ----- Assignability Rule -----

func (c *Checker) assignable(dst, src Type) bool {
	if IsInvalid(dst) || IsInvalid(src) {
		return true // Avoid cascading errors
	}

	// T := T
	if Equal(dst, src) {
		return true
	}

	// T? := T (promote to optional)
	if opt, ok := dst.(*Optional); ok {
		if srcOpt, ok2 := src.(*Optional); ok2 {
			return c.assignable(opt.Inner, srcOpt.Inner)
		}
		return c.assignable(opt.Inner, src)
	}

	// T := T? (unwrapping optional - explicit only, checked elsewhere)
	// This is handled by explicit unwrapping in the type checker

	// any := T (any as sink)
	if Equal(dst, Any) {
		return true
	}

	// T := any (any as source)
	if Equal(src, Any) {
		return true
	}

	// Union handling
	dstUnion, dstIsUnion := dst.(*Union)
	srcUnion, srcIsUnion := src.(*Union)

	if dstIsUnion && !srcIsUnion {
		// dst is union, src is not: src is assignable if it matches any variant
		for _, v := range dstUnion.Variants {
			if c.assignable(v, src) {
				return true
			}
		}
		return false
	}

	if !dstIsUnion && srcIsUnion {
		// src is union, dst is not: src is assignable if all variants are assignable to dst
		for _, v := range srcUnion.Variants {
			if !c.assignable(dst, v) {
				return false
			}
		}
		return true
	}

	if dstIsUnion && srcIsUnion {
		// Both are unions: every variant of src must be assignable to some variant of dst
		for _, sv := range srcUnion.Variants {
			assignableToSome := false
			for _, dv := range dstUnion.Variants {
				if c.assignable(dv, sv) {
					assignableToSome = true
					break
				}
			}
			if !assignableToSome {
				return false
			}
		}
		return true
	}

	// list<...> := list<...> (check "superset" of types)
	dstList, dstIsList := dst.(*List)
	srcList, srcIsList := src.(*List)
	if dstIsList && srcIsList {
		return c.listAssignable(dstList, srcList)
	}

	// dict<K, V> := dict<K2, V2>
	dstDict, dstIsDict := dst.(*Dict)
	srcDict, srcIsDict := src.(*Dict)
	if dstIsDict && srcIsDict {
		if !c.assignable(dstDict.keyType(), srcDict.keyType()) {
			return false
		}
		dstValue := dstDict.ValueType
		srcValue := srcDict.ValueType
		if dstValue == nil {
			dstValue = Any
		}
		if srcValue == nil {
			srcValue = Any
		}
		return c.assignable(dstValue, srcValue)
	}

	// Interface satisfaction: if dst is an interface, check if src satisfies it
	if dstInterface, ok := dst.(*Interface); ok {
		return c.satisfiesInterface(src, dstInterface)
	}

	return false
}

func (c *Checker) listAssignable(dst, src *List) bool {
	// Simplified: each element type of src must be assignable to dst
	for _, s := range src.ElementTypes {
		ok := false
		for _, d := range dst.ElementTypes {
			// Check if s is assignable to d (handles unions, any, etc.)
			if c.assignable(d, s) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

// addInterfaceSatisfactionError adds a detailed error message when a type does not satisfy an interface.
func (c *Checker) addInterfaceSatisfactionError(pos token.Position, typ Type, iface *Interface, context string) {
	var missingMethods []string
	var wrongSignatures []string

	for _, requiredMethod := range iface.Methods {
		method := c.findMethod(typ, requiredMethod.Name)
		if method == nil {
			missingMethods = append(missingMethods, fmt.Sprintf("%s(%s) | %s",
				requiredMethod.Name,
				formatParamTypes(requiredMethod.ParamTypes),
				requiredMethod.Return.String()))
			continue
		}

		// Check signature compatibility
		if !c.methodSignaturesMatch(requiredMethod, method, typ, iface) {
			wrongSignatures = append(wrongSignatures, fmt.Sprintf("%s: expected (%s) | %s, got (%s) | %s",
				requiredMethod.Name,
				formatParamTypes(requiredMethod.ParamTypes),
				requiredMethod.Return.String(),
				formatParamTypes(method.ParamTypes[1:]), // Skip receiver
				method.Result.String()))
		}
	}

	if len(missingMethods) > 0 {
		c.addError(pos, "type %s does not satisfy interface %s: missing methods: %s",
			typ.String(), iface.Name, strings.Join(missingMethods, ", "))
	}
	if len(wrongSignatures) > 0 {
		c.addError(pos, "type %s does not satisfy interface %s: incompatible method signatures: %s",
			typ.String(), iface.Name, strings.Join(wrongSignatures, ", "))
	}
	if len(missingMethods) == 0 && len(wrongSignatures) == 0 {
		// Fallback (shouldn't happen, but be safe)
		c.addError(pos, "type %s does not satisfy interface %s", typ.String(), iface.Name)
	}
}

// formatParamTypes formats a list of parameter types as a string.
func formatParamTypes(params []Type) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = p.String()
	}
	return strings.Join(parts, ", ")
}

// satisfiesInterface checks if a type satisfies an interface (structural typing).
// A type satisfies an interface if it has all required methods with matching signatures.
func (c *Checker) satisfiesInterface(typ Type, iface *Interface) bool {
	// Check each required method
	for _, requiredMethod := range iface.Methods {
		// Find the method on the type
		method := c.findMethod(typ, requiredMethod.Name)
		if method == nil {
			return false // Method not found
		}

		// Check method signature compatibility
		if !c.methodSignaturesMatch(requiredMethod, method, typ, iface) {
			return false // Signature mismatch
		}
	}

	return true
}

// findMethod finds a method on a type. Returns nil if not found.
// Only instance methods are considered (static methods do not satisfy interfaces).
func (c *Checker) findMethod(typ Type, methodName string) *Method {
	switch t := typ.(type) {
	case *Struct:
		// Check instance methods
		if t.InstanceMethods != nil {
			if method, ok := t.InstanceMethods[methodName]; ok {
				return method
			}
		}
		// Static methods do not satisfy interfaces
		return nil

	case *Basic:
		// Check built-in methods for basic types
		return c.findBuiltinMethod(t, methodName)

	case *List:
		// Check built-in methods for lists
		return c.findBuiltinMethodForList(t, methodName)
	case *Dict:
		// Check built-in methods for dicts
		return c.findBuiltinMethodForDict(t, methodName)

	default:
		return nil
	}
}

// findBuiltinMethod finds a built-in method on a basic type.
func (c *Checker) findBuiltinMethod(typ *Basic, methodName string) *Method {
	typeKind, found := builtins.TypeKindFromString(typ.Name)
	if !found {
		return nil
	}

	methodBuiltin := builtins.LookupMethod(typeKind, methodName)
	if methodBuiltin == nil {
		return nil
	}

	// Convert built-in method to Method type
	paramTypes := make([]Type, len(methodBuiltin.Meta.Params))
	for i, p := range methodBuiltin.Meta.Params {
		paramTypes[i] = c.typeRefToType(p)
	}
	resultType := c.typeRefToType(methodBuiltin.Meta.Result)

	return &Method{
		Name:       methodName,
		Receiver:   typ,
		ParamTypes: paramTypes,
		Result:     resultType,
		IsStatic:   false, // Built-in methods are instance methods
	}
}

// findBuiltinMethodForList finds a built-in method on a list type.
func (c *Checker) findBuiltinMethodForList(typ *List, methodName string) *Method {
	methodBuiltin := builtins.LookupMethod(builtins.TypeList, methodName)
	if methodBuiltin == nil {
		return nil
	}

	// Convert built-in method to Method type
	paramTypes := make([]Type, len(methodBuiltin.Meta.Params))
	for i, p := range methodBuiltin.Meta.Params {
		paramTypes[i] = c.typeRefToType(p)
	}
	resultType := c.typeRefToType(methodBuiltin.Meta.Result)

	return &Method{
		Name:       methodName,
		Receiver:   typ,
		ParamTypes: paramTypes,
		Result:     resultType,
		IsStatic:   false, // Built-in methods are instance methods
	}
}

// findBuiltinMethodForDict finds a built-in method on a dict type.
func (c *Checker) findBuiltinMethodForDict(typ *Dict, methodName string) *Method {
	methodBuiltin := builtins.LookupMethod(builtins.TypeDict, methodName)
	if methodBuiltin == nil {
		return nil
	}
	keyType := typ.keyType()
	valueType := typ.ValueType
	if valueType == nil {
		valueType = Any
	}
	var paramTypes []Type
	var resultType Type
	switch methodName {
	case "length":
		paramTypes = []Type{typ}
		resultType = Int
	case "keys":
		paramTypes = []Type{typ}
		resultType = &List{ElementTypes: []Type{keyType}}
	case "values":
		paramTypes = []Type{typ}
		resultType = &List{ElementTypes: []Type{valueType}}
	case "has":
		paramTypes = []Type{typ, keyType}
		resultType = Bool
	case "get":
		paramTypes = []Type{typ, keyType}
		resultType = &Optional{Inner: valueType}
	case "set":
		paramTypes = []Type{typ, keyType, valueType}
		resultType = Void
	case "remove":
		paramTypes = []Type{typ, keyType}
		resultType = Bool
	default:
		return nil
	}
	return &Method{
		Name:       methodName,
		Receiver:   typ,
		ParamTypes: paramTypes,
		Result:     resultType,
		IsStatic:   false,
	}
}

// methodSignaturesMatch checks if a method signature matches an interface requirement.
// It also checks visibility rules: public interfaces can only require public methods.
func (c *Checker) methodSignaturesMatch(required InterfaceMethod, actual *Method, typ Type, iface *Interface) bool {
	// Check parameter count (actual includes receiver as first param, required does not)
	if len(actual.ParamTypes) != len(required.ParamTypes)+1 {
		return false
	}

	// Check parameter types (skip receiver in actual)
	for i, reqParam := range required.ParamTypes {
		actualParam := actual.ParamTypes[i+1] // Skip receiver
		if !Equal(reqParam, actualParam) {
			return false
		}
	}

	// Check return type
	if !Equal(required.Return, actual.Result) {
		return false
	}

	// Check visibility: if interface is public, the method must be public
	if iface.IsPublic {
		// For struct methods, check if method is public
		if _, ok := typ.(*Struct); ok {
			if !actual.IsPublic {
				// Private method cannot satisfy public interface
				// Exception: if both are in the same module, allow it (internal use)
				if iface.DefiningModule != c.currentModule {
					return false
				}
			}
		}
		// Built-in methods are always considered "public" for interface satisfaction
	}

	return true
}

func (c *Checker) checkAwait(a *ast.AwaitExpr) Type {
	exprType := c.checkExpr(a.Expr)
	if IsInvalid(exprType) {
		return Invalid
	}
	fut, ok := exprType.(*Future)
	if !ok {
		c.addError(a.Pos(), "await expects Future<T>, got %s", exprType.String())
		return Invalid
	}
	return fut.Inner
}
