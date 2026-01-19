package ir

import (
	"fmt"

	"avenir/internal/ast"
	"avenir/internal/resolver"
	"avenir/internal/runtime/builtins"
	"avenir/internal/token"
	"avenir/internal/types"
)

// StructTypeInfo represents a struct type in the compiler.
type StructTypeInfo struct {
	Name   string
	Fields []types.Field
}

// Compiler compiles an AST program into an IR module.
type Compiler struct {
	prog *ast.Program
	mod  *Module

	funcIndex map[*ast.FunDecl]int // FunDecl -> index in module

	// For closures: track all functions (named and literals)
	funcLiteralIndex map[*ast.FuncLiteral]int // FuncLiteral -> index in module
	allFuncNodes     []ast.Node               // All functions in order (FunDecl + FuncLiteral)

	funcInfos map[ast.Node]*resolver.FunctionInfo // Resolver metadata

	bindings *types.Bindings // Binding info from type checker

	structTypes map[string]*StructTypeInfo // struct name -> struct type info
	structIndex map[string]int             // struct name -> index in structTypes array

	// Method mapping: (receiver type name, method name) -> function index
	methodIndex map[string]map[string]int // receiver type -> method name -> function index

	world *types.World // Store world for method lookup

	errors []error
}

// Compile compiles a single program. This is a convenience wrapper that
// builds a minimal world and calls CompileWorld.
func Compile(prog *ast.Program) (*Module, []error) {
	if prog == nil {
		return nil, []error{fmt.Errorf("nil program")}
	}

	// Build a minimal world for single-file compilation
	modName := "main"
	if prog.Package != nil {
		modName = prog.Package.Name
	}

	world := &types.World{
		Modules: make(map[string]*types.ModuleInfo),
		Entry:   modName,
	}
	world.Modules[modName] = &types.ModuleInfo{
		Name:  modName,
		Prog:  prog,
		Scope: nil, // Will be set by CheckWorldWithBindings
	}

	// Type-check with bindings
	bindings, typeErrs := types.CheckWorldWithBindings(world)
	if len(typeErrs) > 0 {
		return nil, typeErrs
	}

	// Use unified compilation pipeline
	entryModInfo := world.Modules[modName]
	return CompileWorld(world, entryModInfo, bindings)
}

// CompileWorld compiles all modules in a world into a single IR module.
// bindings must be the result of types.CheckWorldWithBindings(world).
func CompileWorld(world *types.World, entryMod *types.ModuleInfo, bindings *types.Bindings) (*Module, []error) {
	if world == nil || entryMod == nil {
		return nil, []error{fmt.Errorf("nil world or entry module")}
	}
	if bindings == nil {
		return nil, []error{fmt.Errorf("nil bindings (must call CheckWorldWithBindings first)")}
	}

	// Run resolver for all modules
	res := resolver.NewResolver()
	allFuncInfos := make(map[ast.Node]*resolver.FunctionInfo)
	for _, modInfo := range world.Modules {
		funcInfos := res.Resolve(modInfo.Prog)
		for k, v := range funcInfos {
			allFuncInfos[k] = v
		}
	}

	mod := &Module{MainIndex: -1}
	funcIndexByDecl := make(map[*ast.FunDecl]int)
	funcIndexByLiteral := make(map[*ast.FuncLiteral]int)
	allFuncNodes := []ast.Node{}

	// Collect all functions from all modules
	for modName, modInfo := range world.Modules {
		for _, fn := range modInfo.Prog.Funcs {
			idx := len(mod.Functions)
			info := allFuncInfos[fn]
			var upvalues []UpvalueInfo
			if info != nil {
				upvalues = make([]UpvalueInfo, len(info.Upvalues))
				for i, uv := range info.Upvalues {
					upvalues[i] = UpvalueInfo{
						IsLocal: uv.IsLocal,
						Index:   uv.Index,
					}
				}
			}
			// For instance methods, NumParams includes the receiver
			// For static methods, receiver is NOT a parameter
			numParams := len(fn.Params)
			if fn.Receiver != nil && fn.Receiver.Kind == ast.ReceiverInstance {
				numParams++ // Receiver is the first parameter for instance methods
			}
			irFn := &Function{
				Name:      fmt.Sprintf("%s.%s", modName, fn.Name),
				NumParams: numParams,
				Chunk:     Chunk{},
				Upvalues:  upvalues,
			}
			mod.Functions = append(mod.Functions, irFn)
			funcIndexByDecl[fn] = idx
			allFuncNodes = append(allFuncNodes, fn)
		}

		// Collect function literals from this module
		collectFuncLiteralsFromProg(modInfo.Prog, modName, mod, funcIndexByLiteral, &allFuncNodes, allFuncInfos)
	}

	// Find main in entry module using funcIndexByDecl
	entryModName := entryMod.Name
	for _, fn := range entryMod.Prog.Funcs {
		if fn.Name == "main" {
			if idx, ok := funcIndexByDecl[fn]; ok {
				mod.MainIndex = idx
				break
			}
		}
	}

	if mod.MainIndex < 0 {
		return nil, []error{fmt.Errorf("no 'main' function in entry module %q", entryModName)}
	}

	// Collect struct types from all modules
	// We'll look them up from the type checker's scope
	structTypes := make(map[string]*StructTypeInfo)
	structIndex := make(map[string]int)
	idx := 0
	for _, modInfo := range world.Modules {
		// Look up struct types from the module's scope
		if modInfo.Scope != nil {
			for _, st := range modInfo.Prog.Structs {
				// Look up the struct type symbol
				sym := modInfo.Scope.Lookup(st.Name)
				if sym != nil && sym.Kind == types.SymType {
					if structType, ok := sym.Type.(*types.Struct); ok {
						structTypes[st.Name] = &StructTypeInfo{
							Name:   st.Name,
							Fields: structType.Fields,
						}
						structIndex[st.Name] = idx
						idx++
					}
				}
			}
		}
	}

	// Populate methodIndex: map receiver type name -> method name -> function index
	methodIndex := make(map[string]map[string]int)
	for _, fnNode := range allFuncNodes {
		if fnDecl, ok := fnNode.(*ast.FunDecl); ok && fnDecl.Receiver != nil {
			// Extract receiver type name
			if simpleType, ok2 := fnDecl.Receiver.Type.(*ast.SimpleType); ok2 {
				receiverTypeName := simpleType.Name
				if _, exists := methodIndex[receiverTypeName]; !exists {
					methodIndex[receiverTypeName] = make(map[string]int)
				}
				if idx, ok3 := funcIndexByDecl[fnDecl]; ok3 {
					methodIndex[receiverTypeName][fnDecl.Name] = idx
				}
			}
		}
	}

	// Build struct type table aligned with struct indices
	if len(structIndex) > 0 {
		structTable := make([]StructTypeInfo, len(structIndex))
		for name, idx := range structIndex {
			info := structTypes[name]
			if info != nil {
				structTable[idx] = *info
			} else {
				structTable[idx] = StructTypeInfo{Name: name}
			}
		}
		mod.StructTypes = structTable
	}

	// Compile all functions
	c := &Compiler{
		prog:             nil, // Not used in multi-module mode
		mod:              mod,
		funcIndex:        funcIndexByDecl,
		funcLiteralIndex: funcIndexByLiteral,
		allFuncNodes:     allFuncNodes,
		funcInfos:        allFuncInfos,
		bindings:         bindings,
		structTypes:      structTypes,
		structIndex:      structIndex,
		methodIndex:      methodIndex,
		world:            world,
		errors:           []error{},
	}

	// Compile each function
	for _, fnNode := range allFuncNodes {
		var irFn *Function
		var fc *funcCompiler
		if fnDecl, ok := fnNode.(*ast.FunDecl); ok {
			idx := funcIndexByDecl[fnDecl]
			irFn = mod.Functions[idx]
			fc = newFuncCompiler(c, fnDecl, irFn, allFuncInfos[fnDecl])
		} else if funcLit, ok := fnNode.(*ast.FuncLiteral); ok {
			idx := funcIndexByLiteral[funcLit]
			irFn = mod.Functions[idx]
			fc = newFuncCompiler(c, funcLit, irFn, allFuncInfos[funcLit])
		} else {
			c.addError(fnNode.Pos(), "internal error: unknown function node type %T", fnNode)
			continue
		}
		fc.compile()
		irFn.Chunk.NumLocals = fc.nextLocal
		code := irFn.Chunk.Code
		if len(code) == 0 || code[len(code)-1].Op != OpReturn {
			irFn.Chunk.Emit(OpReturn, 0, 0)
		}
	}

	if len(c.errors) > 0 {
		return nil, c.errors
	}

	return mod, nil
}

// Helper functions for CompileWorld

func collectFuncLiteralsFromProg(prog *ast.Program, modName string, mod *Module, funcIndexByLiteral map[*ast.FuncLiteral]int, allFuncNodes *[]ast.Node, allFuncInfos map[ast.Node]*resolver.FunctionInfo) {
	for _, fn := range prog.Funcs {
		collectFuncLiteralsInNode(fn.Body, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
	}
}

func collectFuncLiteralsInNode(node ast.Node, modName string, mod *Module, funcIndexByLiteral map[*ast.FuncLiteral]int, allFuncNodes *[]ast.Node, allFuncInfos map[ast.Node]*resolver.FunctionInfo) {
	switch n := node.(type) {
	case *ast.FuncLiteral:
		idx := len(mod.Functions)
		info := allFuncInfos[n]
		var upvalues []UpvalueInfo
		if info != nil {
			upvalues = make([]UpvalueInfo, len(info.Upvalues))
			for i, uv := range info.Upvalues {
				upvalues[i] = UpvalueInfo{
					IsLocal: uv.IsLocal,
					Index:   uv.Index,
				}
			}
		}
		irFn := &Function{
			Name:      fmt.Sprintf("%s.<lambda_%d>", modName, idx),
			NumParams: len(n.Params),
			Chunk:     Chunk{},
			Upvalues:  upvalues,
		}
		mod.Functions = append(mod.Functions, irFn)
		funcIndexByLiteral[n] = idx
		*allFuncNodes = append(*allFuncNodes, n)
		if n.Body != nil {
			collectFuncLiteralsInNode(n.Body, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
	case *ast.BlockStmt:
		for _, stmt := range n.Stmts {
			collectFuncLiteralsInNode(stmt, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
	case *ast.VarDeclStmt:
		if n.Value != nil {
			collectFuncLiteralsInNode(n.Value, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
	case *ast.AssignStmt:
		collectFuncLiteralsInNode(n.Value, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
	case *ast.StructFieldAssignStmt:
		collectFuncLiteralsInNode(n.Struct, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		collectFuncLiteralsInNode(n.Value, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
	case *ast.ExprStmt:
		collectFuncLiteralsInNode(n.Expression, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
	case *ast.ReturnStmt:
		if n.Result != nil {
			collectFuncLiteralsInNode(n.Result, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
	case *ast.IfStmt:
		collectFuncLiteralsInNode(n.Cond, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		if n.Then != nil {
			collectFuncLiteralsInNode(n.Then, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
		if n.Else != nil {
			collectFuncLiteralsInNode(n.Else, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
	case *ast.WhileStmt:
		collectFuncLiteralsInNode(n.Cond, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		if n.Body != nil {
			collectFuncLiteralsInNode(n.Body, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
	case *ast.ForStmt:
		if n.Init != nil {
			collectFuncLiteralsInNode(n.Init, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
		if n.Cond != nil {
			collectFuncLiteralsInNode(n.Cond, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
		if n.Post != nil {
			collectFuncLiteralsInNode(n.Post, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
		if n.Body != nil {
			collectFuncLiteralsInNode(n.Body, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
	case *ast.ForEachStmt:
		collectFuncLiteralsInNode(n.ListExpr, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		if n.Body != nil {
			collectFuncLiteralsInNode(n.Body, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
	case *ast.TryStmt:
		if n.Body != nil {
			collectFuncLiteralsInNode(n.Body, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
		if n.CatchBody != nil {
			collectFuncLiteralsInNode(n.CatchBody, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
	case *ast.ThrowStmt:
		collectFuncLiteralsInNode(n.Expr, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
	case *ast.BinaryExpr:
		collectFuncLiteralsInNode(n.Left, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		collectFuncLiteralsInNode(n.Right, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
	case *ast.UnaryExpr:
		collectFuncLiteralsInNode(n.X, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
	case *ast.CallExpr:
		collectFuncLiteralsInNode(n.Callee, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		for _, arg := range n.Args {
			if namedArg, ok := arg.(*ast.NamedArg); ok {
				collectFuncLiteralsInNode(namedArg.Value, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
			} else {
				collectFuncLiteralsInNode(arg, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
			}
		}
	case *ast.IndexExpr:
		collectFuncLiteralsInNode(n.X, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		collectFuncLiteralsInNode(n.Index, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
	case *ast.ListLiteral:
		for _, el := range n.Elements {
			collectFuncLiteralsInNode(el, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
		}
	case *ast.MemberExpr:
		collectFuncLiteralsInNode(n.X, modName, mod, funcIndexByLiteral, allFuncNodes, allFuncInfos)
	}
}

func findFuncLiteralInProg(prog *ast.Program, target *ast.FuncLiteral) bool {
	for _, fn := range prog.Funcs {
		if findFuncLiteralInNode(fn.Body, target) {
			return true
		}
	}
	return false
}

func findFuncLiteralInNode(node ast.Node, target *ast.FuncLiteral) bool {
	switch n := node.(type) {
	case *ast.FuncLiteral:
		if n == target {
			return true
		}
		if n.Body != nil && findFuncLiteralInNode(n.Body, target) {
			return true
		}
	case *ast.BlockStmt:
		for _, stmt := range n.Stmts {
			if findFuncLiteralInNode(stmt, target) {
				return true
			}
		}
	case *ast.VarDeclStmt:
		if n.Value != nil && findFuncLiteralInNode(n.Value, target) {
			return true
		}
	case *ast.AssignStmt:
		if findFuncLiteralInNode(n.Value, target) {
			return true
		}
	case *ast.ExprStmt:
		if findFuncLiteralInNode(n.Expression, target) {
			return true
		}
	case *ast.ReturnStmt:
		if n.Result != nil && findFuncLiteralInNode(n.Result, target) {
			return true
		}
	case *ast.IfStmt:
		if findFuncLiteralInNode(n.Cond, target) || (n.Then != nil && findFuncLiteralInNode(n.Then, target)) || (n.Else != nil && findFuncLiteralInNode(n.Else, target)) {
			return true
		}
	case *ast.WhileStmt:
		if findFuncLiteralInNode(n.Cond, target) || (n.Body != nil && findFuncLiteralInNode(n.Body, target)) {
			return true
		}
	case *ast.ForStmt:
		if (n.Init != nil && findFuncLiteralInNode(n.Init, target)) || (n.Cond != nil && findFuncLiteralInNode(n.Cond, target)) || (n.Post != nil && findFuncLiteralInNode(n.Post, target)) || (n.Body != nil && findFuncLiteralInNode(n.Body, target)) {
			return true
		}
	case *ast.ForEachStmt:
		if findFuncLiteralInNode(n.ListExpr, target) || (n.Body != nil && findFuncLiteralInNode(n.Body, target)) {
			return true
		}
	case *ast.TryStmt:
		if (n.Body != nil && findFuncLiteralInNode(n.Body, target)) || (n.CatchBody != nil && findFuncLiteralInNode(n.CatchBody, target)) {
			return true
		}
	case *ast.ThrowStmt:
		if findFuncLiteralInNode(n.Expr, target) {
			return true
		}
	case *ast.BinaryExpr:
		if findFuncLiteralInNode(n.Left, target) || findFuncLiteralInNode(n.Right, target) {
			return true
		}
	case *ast.UnaryExpr:
		if findFuncLiteralInNode(n.X, target) {
			return true
		}
	case *ast.CallExpr:
		if findFuncLiteralInNode(n.Callee, target) {
			return true
		}
		for _, arg := range n.Args {
			if namedArg, ok := arg.(*ast.NamedArg); ok {
				if findFuncLiteralInNode(namedArg.Value, target) {
					return true
				}
			} else if findFuncLiteralInNode(arg, target) {
				return true
			}
		}
	case *ast.IndexExpr:
		if findFuncLiteralInNode(n.X, target) || findFuncLiteralInNode(n.Index, target) {
			return true
		}
	case *ast.ListLiteral:
		for _, el := range n.Elements {
			if findFuncLiteralInNode(el, target) {
				return true
			}
		}
	case *ast.MemberExpr:
		if findFuncLiteralInNode(n.X, target) {
			return true
		}
	}
	return false
}

func (c *Compiler) addError(pos token.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf("%d:%d: ", pos.Line, pos.Column) + fmt.Sprintf(format, args...)
	c.errors = append(c.errors, fmt.Errorf("%s", msg))
}

// ---------- Local scopes for local variables ----------

type localScope struct {
	parent *localScope
	slots  map[string]int // name -> slot index
}

func newLocalScope(parent *localScope) *localScope {
	return &localScope{
		parent: parent,
		slots:  make(map[string]int),
	}
}

func (s *localScope) lookup(name string) (int, bool) {
	for sc := s; sc != nil; sc = sc.parent {
		if idx, ok := sc.slots[name]; ok {
			return idx, true
		}
	}
	return 0, false
}

// ---------- funcCompiler ----------

type loopContext struct {
	breakJumps []int // indices of OpJump instructions that should jump to loop exit
}

type funcCompiler struct {
	c     *Compiler
	fnAst *ast.FunDecl
	fnLit *ast.FuncLiteral
	fn    *Function
	chunk *Chunk

	scope     *localScope
	nextLocal int
	parentFC  *funcCompiler // For nested functions: parent function compiler
	fnInfo    *resolver.FunctionInfo

	loopStack []loopContext // stack of active loops for break handling
}

// newFuncCompiler creates a new funcCompiler for a function AST node.
// fnNode can be either *ast.FunDecl or *ast.FuncLiteral.
func newFuncCompiler(c *Compiler, fnNode ast.Node, irFn *Function, fnInfo *resolver.FunctionInfo) *funcCompiler {
	fc := &funcCompiler{
		c:      c,
		fn:     irFn,
		chunk:  &irFn.Chunk,
		scope:  newLocalScope(nil),
		fnInfo: fnInfo,
	}

	var params []*ast.Param
	var receiver *ast.Receiver
	switch n := fnNode.(type) {
	case *ast.FunDecl:
		fc.fnAst = n
		params = n.Params
		receiver = n.Receiver
	case *ast.FuncLiteral:
		fc.fnLit = n
		params = n.Params
		receiver = nil
	}

	// For instance methods, receiver occupies slot 0, parameters start at slot 1
	// For static methods and regular functions, parameters start at slot 0
	slotOffset := 0
	if receiver != nil && receiver.Kind == ast.ReceiverInstance {
		// Only instance methods have a receiver variable
		fc.scope.slots[receiver.Name] = 0
		fc.nextLocal++
		slotOffset = 1
	}
	// Parameters occupy slots starting from slotOffset
	for i, p := range params {
		fc.scope.slots[p.Name] = slotOffset + i
		fc.nextLocal++
	}
	return fc
}

func (fc *funcCompiler) addError(node ast.Node, format string, args ...interface{}) {
	pos := node.Pos()
	fc.c.addError(pos, format, args...)
}

func (fc *funcCompiler) allocLocal(name string, node ast.Node) int {
	if fc.scope == nil {
		fc.addError(node, "internal error: no scope for local %q", name)
		return 0
	}
	if _, exists := fc.scope.slots[name]; exists {
		// Type checker would already catch this, but check anyway
		fc.addError(node, "redefinition of %q in the same block", name)
	}
	slot := fc.nextLocal
	fc.nextLocal++
	fc.scope.slots[name] = slot
	return slot
}

func (fc *funcCompiler) lookupLocal(name string) (int, bool) {
	if fc.scope == nil {
		return 0, false
	}
	return fc.scope.lookup(name)
}

func (fc *funcCompiler) resolveLocal(name string, node ast.Node) int {
	idx, ok := fc.lookupLocal(name)
	if !ok {
		// Check if it's an upvalue
		if upvalueIdx, ok := fc.lookupUpvalue(name); ok {
			return upvalueIdx // Return upvalue index (will be used with OpStoreUpvalue)
		}
		fc.addError(node, "unknown local %q", name)
		return 0
	}
	return idx
}

// lookupUpvalue checks if a name is an upvalue for the current function.
func (fc *funcCompiler) pushLoop() {
	fc.loopStack = append(fc.loopStack, loopContext{})
}

func (fc *funcCompiler) recordBreakJump(jumpIndex int) {
	if len(fc.loopStack) == 0 {
		// Should not happen if type checker did its job, but be defensive.
		return
	}
	top := len(fc.loopStack) - 1
	fc.loopStack[top].breakJumps = append(fc.loopStack[top].breakJumps, jumpIndex)
}

func (fc *funcCompiler) popLoop(afterLoopIP int) {
	if len(fc.loopStack) == 0 {
		return
	}
	top := len(fc.loopStack) - 1
	ctx := fc.loopStack[top]
	fc.loopStack = fc.loopStack[:top]

	// Patch all break jumps to the instruction after the loop
	for _, idx := range ctx.breakJumps {
		fc.chunk.Code[idx].A = afterLoopIP
	}
}

func (fc *funcCompiler) lookupUpvalue(name string) (int, bool) {
	// Get function info for current function
	var fnNode ast.Node
	if fc.fnAst != nil {
		fnNode = fc.fnAst
	} else if fc.fnLit != nil {
		fnNode = fc.fnLit
	} else {
		return 0, false
	}

	info := fc.c.funcInfos[fnNode]
	if info == nil {
		return 0, false
	}

	// Find the upvalue by name
	for i, uv := range info.Upvalues {
		if uv.Name == name {
			return i, true
		}
	}
	return 0, false
}

func (fc *funcCompiler) compile() {
	// Parameters (and receiver for methods) are already in their correct stack slots
	// when the function starts. The VM's callClosure sets up the frame so that:
	// - Arguments are at positions base, base+1, ..., base+numArgs-1
	// - The frame's Base points to the first argument
	// - Parameters are already accessible via OpLoadLocal with slot indices 0, 1, ...
	// So we don't need to store them - they're already in place.
	// We can proceed directly to compiling the function body.
	if fc.fnAst != nil {
		fc.compileBlock(fc.fnAst.Body)
	} else if fc.fnLit != nil {
		fc.compileBlock(fc.fnLit.Body)
	}
}

func (fc *funcCompiler) compileBlock(b *ast.BlockStmt) {
	// Create a new scope for the block
	prev := fc.scope
	fc.scope = newLocalScope(prev)
	for _, st := range b.Stmts {
		fc.compileStmt(st)
	}
	fc.scope = prev
}

func (fc *funcCompiler) compileStmt(s ast.Stmt) {
	switch st := s.(type) {
	case *ast.BlockStmt:
		fc.compileBlock(st)

	case *ast.VarDeclStmt:
		slot := fc.allocLocal(st.Name, st)
		fc.compileExpr(st.Value)
		// Store local: take value from top of stack, save to slot (no pop)
		fc.chunk.Emit(OpStoreLocal, slot, 0)
		// Value is not needed as an expression
		fc.chunk.Emit(OpPop, 0, 0)

	case *ast.AssignStmt:
		// Check if it's a local or upvalue
		if slot, ok := fc.lookupLocal(st.Name); ok {
			fc.compileExpr(st.Value)
			fc.chunk.Emit(OpStoreLocal, slot, 0)
			fc.chunk.Emit(OpPop, 0, 0)
		} else if upvalueIdx, ok := fc.lookupUpvalue(st.Name); ok {
			fc.compileExpr(st.Value)
			fc.chunk.Emit(OpStoreUpvalue, upvalueIdx, 0)
			fc.chunk.Emit(OpPop, 0, 0)
		} else {
			fc.addError(st, "unknown variable %q", st.Name)
		}

	case *ast.StructFieldAssignStmt:
		fc.compileStructFieldAssign(st)

	case *ast.ExprStmt:
		fc.compileExpr(st.Expression)
		fc.chunk.Emit(OpPop, 0, 0)

	case *ast.IfStmt:
		fc.compileIf(st)

	case *ast.WhileStmt:
		fc.compileWhile(st)

	case *ast.ForStmt:
		fc.compileFor(st)

	case *ast.ForEachStmt:
		fc.compileForEach(st)

	case *ast.ThrowStmt:
		fc.compileExpr(st.Expr)
		fc.chunk.Emit(OpThrow, 0, 0)

	case *ast.TryStmt:
		fc.compileTry(st)

	case *ast.ReturnStmt:
		if st.Result != nil {
			fc.compileExpr(st.Result)
			fc.chunk.Emit(OpReturn, 0, 1)
		} else {
			fc.chunk.Emit(OpReturn, 0, 0)
		}

	case *ast.BreakStmt:
		fc.compileBreak(st)

	default:
		// Other statements will be handled here
	}
}

func (fc *funcCompiler) compileBreak(s *ast.BreakStmt) {
	if len(fc.loopStack) == 0 {
		// Type checker should have already reported this, but keep a safety net.
		fc.addError(s, "'break' used outside of a loop")
		return
	}
	// Emit a jump with placeholder target; will be patched when we close the loop.
	jumpIndex := fc.chunk.Emit(OpJump, 0, 0)
	fc.recordBreakJump(jumpIndex)
}

func (fc *funcCompiler) compileWhile(s *ast.WhileStmt) {
	// Start a new loop context
	fc.pushLoop()

	// Label for loop start
	loopStart := len(fc.chunk.Code)

	// Compile condition
	fc.compileExpr(s.Cond)
	jumpIfFalseIdx := fc.chunk.Emit(OpJumpIfFalse, 0, 0)

	// Compile body
	fc.compileBlock(s.Body)

	// Jump back to loop start
	fc.chunk.Emit(OpJump, loopStart, 0)

	// After loop
	afterLoop := len(fc.chunk.Code)
	fc.chunk.Code[jumpIfFalseIdx].A = afterLoop

	// Patch all 'break' jumps in this loop to jump to afterLoop
	fc.popLoop(afterLoop)
}

func (fc *funcCompiler) compileFor(s *ast.ForStmt) {
	// Compile init (if present)
	if s.Init != nil {
		fc.compileStmt(s.Init)
	}

	// Start a new loop context
	fc.pushLoop()

	// Label for loop condition
	condStart := len(fc.chunk.Code)

	// Compile condition (if present)
	var jumpIfFalseIdx int = -1
	if s.Cond != nil {
		fc.compileExpr(s.Cond)
		jumpIfFalseIdx = fc.chunk.Emit(OpJumpIfFalse, 0, 0)
	}

	// Compile body
	fc.compileBlock(s.Body)

	// Compile post (if present)
	if s.Post != nil {
		fc.compileStmt(s.Post)
	}

	// Jump back to condition
	fc.chunk.Emit(OpJump, condStart, 0)

	// After loop
	afterLoop := len(fc.chunk.Code)
	if jumpIfFalseIdx >= 0 {
		fc.chunk.Code[jumpIfFalseIdx].A = afterLoop
	}

	// Patch 'break' targets in this loop
	fc.popLoop(afterLoop)
}

func (fc *funcCompiler) compileForEach(s *ast.ForEachStmt) {
	// Evaluate list expression once and store in a temporary local
	listSlot := fc.allocLocal("_foreach_list", s)
	fc.compileExpr(s.ListExpr)
	fc.chunk.Emit(OpStoreLocal, listSlot, 0)
	fc.chunk.Emit(OpPop, 0, 0)

	// Allocate index variable
	indexSlot := fc.allocLocal("_foreach_index", s)
	zeroIdx := fc.chunk.AddConstInt(0)
	fc.chunk.Emit(OpConst, zeroIdx, 0)
	fc.chunk.Emit(OpStoreLocal, indexSlot, 0)
	fc.chunk.Emit(OpPop, 0, 0)

	// Allocate loop variable
	varSlot := fc.allocLocal(s.VarName, s)

	// Start a new loop context
	fc.pushLoop()

	// Label for loop start
	loopStart := len(fc.chunk.Code)

	// Check if index < len(list) - if false (i.e., index >= len), exit loop
	// Load index first, then len (for OpLt: pops len, then index, compares index < len)
	fc.chunk.Emit(OpLoadLocal, indexSlot, 0)
	// Load list and call len builtin
	fc.chunk.Emit(OpLoadLocal, listSlot, 0)
	fc.chunk.Emit(OpCallBuiltin, int(builtins.Len), 1)
	// Compare: index < len (OpLt pops len then index, does index < len)
	fc.chunk.Emit(OpLt, 0, 0)
	// If index < len is false (i.e., index >= len), jump out
	jumpIfFalseIdx := fc.chunk.Emit(OpJumpIfFalse, 0, 0)

	// Load current element: list[index]
	fc.chunk.Emit(OpLoadLocal, listSlot, 0)
	fc.chunk.Emit(OpLoadLocal, indexSlot, 0)
	fc.chunk.Emit(OpIndex, 0, 0)
	fc.chunk.Emit(OpStoreLocal, varSlot, 0)
	fc.chunk.Emit(OpPop, 0, 0)

	// Compile body
	fc.compileBlock(s.Body)

	// Increment index
	oneIdx := fc.chunk.AddConstInt(1)
	fc.chunk.Emit(OpLoadLocal, indexSlot, 0)
	fc.chunk.Emit(OpConst, oneIdx, 0)
	fc.chunk.Emit(OpAdd, 0, 0)
	fc.chunk.Emit(OpStoreLocal, indexSlot, 0)
	fc.chunk.Emit(OpPop, 0, 0)

	// Jump back to loop start
	fc.chunk.Emit(OpJump, loopStart, 0)

	// After loop
	afterLoop := len(fc.chunk.Code)
	fc.chunk.Code[jumpIfFalseIdx].A = afterLoop

	// Patch 'break' jumps in this loop
	fc.popLoop(afterLoop)
}

func (fc *funcCompiler) compileTry(s *ast.TryStmt) {
	// We'll patch handler IP later
	beginIdx := fc.chunk.Emit(OpBeginTry, 0, 0)

	// compile try block
	fc.compileBlock(s.Body)

	fc.chunk.Emit(OpEndTry, 0, 0)
	jmpOverCatch := fc.chunk.Emit(OpJump, 0, 0)

	// handler starts here
	handlerIP := len(fc.chunk.Code)
	fc.chunk.Code[beginIdx].A = handlerIP

	if s.CatchBody != nil {
		// Exception value is on top of the stack.
		// Allocate local for catch variable and store it.
		slot := fc.allocLocal(s.CatchName, s)
		fc.chunk.Emit(OpStoreLocal, slot, 0)
		fc.chunk.Emit(OpPop, 0, 0)

		fc.compileBlock(s.CatchBody)
	}

	endIP := len(fc.chunk.Code)
	fc.chunk.Code[jmpOverCatch].A = endIP
}

func (fc *funcCompiler) compileIf(s *ast.IfStmt) {
	// cond
	fc.compileExpr(s.Cond)
	jumpIfFalseIdx := fc.chunk.Emit(OpJumpIfFalse, 0, 0)

	// then
	fc.compileBlock(s.Then)

	if s.Else != nil {
		// Jump over else block
		jumpEndIdx := fc.chunk.Emit(OpJump, 0, 0)
		elsePos := len(fc.chunk.Code)
		fc.chunk.Code[jumpIfFalseIdx].A = elsePos

		fc.compileStmt(s.Else)

		endPos := len(fc.chunk.Code)
		fc.chunk.Code[jumpEndIdx].A = endPos
	} else {
		afterThen := len(fc.chunk.Code)
		fc.chunk.Code[jumpIfFalseIdx].A = afterThen
	}
}

// ---------- Expressions ----------

func (fc *funcCompiler) compileExpr(e ast.Expr) {
	switch ex := e.(type) {
	case *ast.IntLiteral:
		idx := fc.chunk.AddConstInt(ex.Value)
		fc.chunk.Emit(OpConst, idx, 0)

	case *ast.FloatLiteral:
		idx := fc.chunk.AddConstFloat(ex.Value)
		fc.chunk.Emit(OpConst, idx, 0)

	case *ast.StringLiteral:
		idx := fc.chunk.AddConstString(ex.Value)
		fc.chunk.Emit(OpConst, idx, 0)

	case *ast.InterpolatedString:
		if len(ex.Parts) == 0 {
			idx := fc.chunk.AddConstString("")
			fc.chunk.Emit(OpConst, idx, 0)
			return
		}
		first := true
		for _, part := range ex.Parts {
			switch p := part.(type) {
			case *ast.StringTextPart:
				idx := fc.chunk.AddConstString(p.Value)
				fc.chunk.Emit(OpConst, idx, 0)
			case *ast.StringExprPart:
				fc.compileExpr(p.Expr)
				fc.chunk.Emit(OpStringify, 0, 0)
			default:
				fc.addError(ex, "unknown interpolated string part")
				return
			}
			if first {
				first = false
				continue
			}
			fc.chunk.Emit(OpConcatString, 0, 0)
		}

	case *ast.BytesLiteral:
		idx := fc.chunk.AddConstBytes(ex.Value)
		fc.chunk.Emit(OpConst, idx, 0)

	case *ast.BoolLiteral:
		idx := fc.chunk.AddConstBool(ex.Value)
		fc.chunk.Emit(OpConst, idx, 0)

	case *ast.NoneLiteral:
		noneIdx := fc.chunk.AddConstNone()
		fc.chunk.Emit(OpConst, noneIdx, 0)

	case *ast.SomeLiteral:
		fc.compileExpr(ex.Value)
		fc.chunk.Emit(OpMakeSome, 0, 0)

	case *ast.IdentExpr:
		// 1. locals
		if slot, ok := fc.lookupLocal(ex.Name); ok {
			fc.chunk.Emit(OpLoadLocal, slot, 0)
			return
		}
		// 2. upvalues
		if upvalueIdx, ok := fc.lookupUpvalue(ex.Name); ok {
			fc.chunk.Emit(OpLoadUpvalue, upvalueIdx, 0)
			return
		}
		// 3. function symbol bound by the type checker
		if fc.c.bindings != nil {
			if sym, ok := fc.c.bindings.Idents[ex]; ok && sym.Kind == types.SymFunc {
				if fnDecl, ok2 := sym.Node.(*ast.FunDecl); ok2 {
					if idx, ok3 := fc.c.funcIndex[fnDecl]; ok3 {
						fc.chunk.Emit(OpClosure, idx, 0)
						return
					}
				}
			}
		}
		// (builtins are handled specially in compileCall, not here)
		fc.addError(ex, "unknown identifier %q", ex.Name)

	case *ast.FuncLiteral:
		// Function literal: emit OpClosure
		//
		// Upvalue handling:
		//   - For IsLocal upvalues: We don't push anything. OpClosure will create
		//     an open upvalue pointing directly to the stack slot in the current frame.
		//   - For non-local upvalues (parent's upvalues): We push the value by loading
		//     the parent's upvalue. OpClosure will then try to share the parent's upvalue
		//     object (for nested closures sharing state).
		fnIndex, ok := fc.c.funcLiteralIndex[ex]
		if !ok {
			fc.addError(ex, "internal error: function literal not registered")
			return
		}
		info := fc.c.funcInfos[ex]
		numUpvalues := 0
		if info != nil {
			numUpvalues = len(info.Upvalues)
			// Push values for non-local upvalues (parent's upvalues)
			for _, uv := range info.Upvalues {
				if !uv.IsLocal {
					// Capture from parent function's upvalue - load and push the value
					parentUpvalueIdx, ok := fc.lookupUpvalue(uv.Name)
					if !ok {
						fc.addError(ex, "internal error: parent upvalue %q not found", uv.Name)
						// Push a dummy value to keep stack consistent
						zeroIdx := fc.chunk.AddConstInt(0)
						fc.chunk.Emit(OpConst, zeroIdx, 0)
						continue
					}
					// Load parent upvalue value to stack
					fc.chunk.Emit(OpLoadUpvalue, parentUpvalueIdx, 0)
				}
				// For IsLocal upvalues, we don't push - OpClosure will handle it directly
			}
		}
		fc.chunk.Emit(OpClosure, fnIndex, numUpvalues)

	case *ast.UnaryExpr:
		fc.compileExpr(ex.X)
		switch ex.Op {
		case token.Minus:
			fc.chunk.Emit(OpNegate, 0, 0)
		case token.Bang:
			// !x => x == false
			falseIdx := fc.chunk.AddConstBool(false)
			fc.chunk.Emit(OpConst, falseIdx, 0)
			fc.chunk.Emit(OpEq, 0, 0)
		default:
			fc.addError(ex, "unsupported unary op %s", ex.Op)
		}

	case *ast.BinaryExpr:
		if ex.Op == token.AndAnd || ex.Op == token.OrOr {
			fc.compileLogical(ex)
		} else {
			fc.compileBinary(ex)
		}

	case *ast.CallExpr:
		fc.compileCall(ex)

	case *ast.ListLiteral:
		fc.compileListLiteral(ex)

	case *ast.DictLiteral:
		fc.compileDictLiteral(ex)

	case *ast.StructLiteral:
		fc.compileStructLiteral(ex)

	case *ast.IndexExpr:
		// X[Index]
		fc.compileExpr(ex.X)
		fc.compileExpr(ex.Index)
		fc.chunk.Emit(OpIndex, 0, 0)

	case *ast.MemberExpr:
		// Use bindings to resolve member expressions
		if fc.c.bindings != nil {
			// Check if ex.X is a type identifier (static method call)
			if ident, ok := ex.X.(*ast.IdentExpr); ok {
				// Check if it's a type (static method call)
				if xSym, ok2 := fc.c.bindings.Idents[ident]; ok2 && xSym.Kind == types.SymType {
					if structType, ok3 := xSym.Type.(*types.Struct); ok3 {
						// Look up static method
						if methodMap, ok4 := fc.c.methodIndex[structType.Name]; ok4 {
							if methodIdx, ok5 := methodMap[ex.Name]; ok5 {
								// Verify it's actually a static method by checking the FunDecl
								if fc.c.world != nil {
									for _, modInfo := range fc.c.world.Modules {
										for _, fn := range modInfo.Prog.Funcs {
											if fn.Receiver != nil && fn.Name == ex.Name {
												if simpleType, ok6 := fn.Receiver.Type.(*ast.SimpleType); ok6 && simpleType.Name == structType.Name {
													if fn.Receiver.Kind == ast.ReceiverStatic {
														// This is a static method access
														fc.chunk.Emit(OpClosure, methodIdx, 0)
														return
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}

			// Check for instance methods (ex.X is a value, not a type)
			var xType types.Type
			if ident, ok := ex.X.(*ast.IdentExpr); ok {
				// ex.X is a variable - look it up in Idents
				if xSym, ok2 := fc.c.bindings.Idents[ident]; ok2 {
					// Skip if it's a type (already handled above)
					if xSym.Kind != types.SymType {
						xType = xSym.Type
					}
				}
			} else {
				// ex.X is an expression - look it up in ExprTypes
				if t, ok2 := fc.c.bindings.ExprTypes[ex.X]; ok2 {
					xType = t
				}
			}

			// If we found a struct type, check for instance methods
			if structType, ok := xType.(*types.Struct); ok {
				if methodMap, ok2 := fc.c.methodIndex[structType.Name]; ok2 {
					if methodIdx, ok3 := methodMap[ex.Name]; ok3 {
						// Verify it's actually an instance method by checking the FunDecl
						if fc.c.world != nil {
							for _, modInfo := range fc.c.world.Modules {
								for _, fn := range modInfo.Prog.Funcs {
									if fn.Receiver != nil && fn.Name == ex.Name {
										if simpleType, ok4 := fn.Receiver.Type.(*ast.SimpleType); ok4 && simpleType.Name == structType.Name {
											if fn.Receiver.Kind == ast.ReceiverInstance {
												// This is an instance method access
												fc.chunk.Emit(OpClosure, methodIdx, 0)
												return
											}
										}
									}
								}
							}
						}
					}
				}
			}

			// Check if it's a function (module access) or static method
			if sym, ok := fc.c.bindings.Members[ex]; ok && sym.Kind == types.SymFunc {
				// Check if it's a module function (has FunDecl node)
				if fnDecl, ok2 := sym.Node.(*ast.FunDecl); ok2 {
					if idx, ok3 := fc.c.funcIndex[fnDecl]; ok3 {
						fc.chunk.Emit(OpClosure, idx, 0)
						return
					}
				}
				// If it's a SymFunc but no FunDecl node, it might be a static method
				// Look up the method in methodIndex
				if ident, ok2 := ex.X.(*ast.IdentExpr); ok2 {
					// Try to get the type from bindings
					var structType *types.Struct
					if xSym, ok3 := fc.c.bindings.Idents[ident]; ok3 && xSym.Kind == types.SymType {
						if st, ok4 := xSym.Type.(*types.Struct); ok4 {
							structType = st
						}
					}
					// Also try global lookup if not found
					if structType == nil && fc.c.world != nil {
						for _, modInfo := range fc.c.world.Modules {
							if modInfo.Scope != nil {
								if typeSym := modInfo.Scope.Lookup(ident.Name); typeSym != nil && typeSym.Kind == types.SymType {
									if st, ok4 := typeSym.Type.(*types.Struct); ok4 {
										structType = st
										break
									}
								}
							}
						}
					}
					if structType != nil {
						if methodMap, ok3 := fc.c.methodIndex[structType.Name]; ok3 {
							if methodIdx, ok4 := methodMap[ex.Name]; ok4 {
								// Verify it's a static method
								if fc.c.world != nil {
									for _, modInfo := range fc.c.world.Modules {
										for _, fn := range modInfo.Prog.Funcs {
											if fn.Receiver != nil && fn.Name == ex.Name {
												if simpleType, ok5 := fn.Receiver.Type.(*ast.SimpleType); ok5 && simpleType.Name == structType.Name {
													if fn.Receiver.Kind == ast.ReceiverStatic {
														fc.chunk.Emit(OpClosure, methodIdx, 0)
														return
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}

			// Check if it's a struct field access
			if sym, ok := fc.c.bindings.Members[ex]; ok && sym.Kind == types.SymVar {
				// This is a struct field access
				fc.compileExpr(ex.X)
				// Get the type of X (the struct instance)
				xType, ok := fc.c.bindings.ExprTypes[ex.X]
				if !ok {
					fc.addError(ex, "cannot determine type of struct expression")
					return
				}
				structType, ok := xType.(*types.Struct)
				if !ok {
					fc.addError(ex, "member access on non-struct type")
					return
				}
				// Find field index
				fieldIdx := -1
				for i, f := range structType.Fields {
					if f.Name == ex.Name {
						fieldIdx = i
						break
					}
				}
				if fieldIdx < 0 {
					fc.addError(ex, "struct %s has no field %q", structType.Name, ex.Name)
					return
				}
				fc.chunk.Emit(OpLoadField, fieldIdx, 0)
				return
			}

			// Dict key access: compile as index with constant string key
			var dictType *types.Dict
			if ident, ok := ex.X.(*ast.IdentExpr); ok {
				if xSym, ok2 := fc.c.bindings.Idents[ident]; ok2 && xSym.Kind != types.SymType {
					if dt, ok3 := xSym.Type.(*types.Dict); ok3 {
						dictType = dt
					}
				}
			} else if t, ok2 := fc.c.bindings.ExprTypes[ex.X]; ok2 {
				if dt, ok3 := t.(*types.Dict); ok3 {
					dictType = dt
				}
			}
			if dictType != nil {
				fc.compileExpr(ex.X)
				keyIdx := fc.chunk.AddConstString(ex.Name)
				fc.chunk.Emit(OpConst, keyIdx, 0)
				fc.chunk.Emit(OpIndex, 0, 0)
				return
			}
		}
		fc.addError(ex, "cannot resolve member %q", ex.Name)

	default:
		fc.addError(e, "unsupported expression of type %T", e)
	}
}

func (fc *funcCompiler) compileListLiteral(lit *ast.ListLiteral) {
	// Compile elements in order: e1, e2, ..., en
	count := len(lit.Elements)
	for _, el := range lit.Elements {
		fc.compileExpr(el)
	}
	// OpMakeList pops n values from stack and creates a list
	fc.chunk.Emit(OpMakeList, count, 0)
}

func (fc *funcCompiler) compileDictLiteral(lit *ast.DictLiteral) {
	// Compile entries in order: key1, value1, key2, value2, ...
	count := len(lit.Entries)
	for _, entry := range lit.Entries {
		keyIdx := fc.chunk.AddConstString(entry.Key)
		fc.chunk.Emit(OpConst, keyIdx, 0)
		fc.compileExpr(entry.Value)
	}
	// OpMakeDict pops 2*count values and creates a dict
	fc.chunk.Emit(OpMakeDict, count, 0)
}

func (fc *funcCompiler) compileStructLiteral(lit *ast.StructLiteral) {
	// Get struct type info
	structInfo, ok := fc.c.structTypes[lit.TypeName]
	if !ok {
		fc.addError(lit, "unknown struct type %q", lit.TypeName)
		return
	}

	// Build a map of field name -> value expression
	fieldMap := make(map[string]ast.Expr)
	for _, fieldInit := range lit.Fields {
		fieldMap[fieldInit.Name] = fieldInit.Value
	}

	// Compile field values in the order they appear in the struct definition
	fieldCount := len(structInfo.Fields)
	for _, field := range structInfo.Fields {
		valueExpr, ok := fieldMap[field.Name]
		if !ok {
			// Field not provided - use default if available
			if field.DefaultExpr != nil {
				// Compile the default expression (it's a compile-time constant)
				fc.compileExpr(field.DefaultExpr)
			} else {
				// No default and not provided - this should have been caught by type checker
				// but we need to push something to avoid crashes
				fc.addError(lit, "missing required field %q in struct literal", field.Name)
				fc.chunk.Emit(OpConst, fc.chunk.AddConstInt(0), 0)
			}
		} else {
			// Field provided - compile the provided value
			fc.compileExpr(valueExpr)
		}
	}

	// Get struct type index
	structIdx, ok := fc.c.structIndex[lit.TypeName]
	if !ok {
		fc.addError(lit, "struct type %q not found in index", lit.TypeName)
		return
	}

	// OpMakeStruct pops fieldCount values and creates a struct
	fc.chunk.Emit(OpMakeStruct, structIdx, fieldCount)
}

func (fc *funcCompiler) compileStructFieldAssign(s *ast.StructFieldAssignStmt) {
	// For now, we only support assignment to struct variables (not complex expressions)
	// This is a reasonable restriction: p.x = 10 is allowed, but getPoint().x = 10 is not
	if ident, ok := s.Struct.(*ast.IdentExpr); ok {
		// Get struct type from bindings
		var structType *types.Struct
		if fc.c.bindings != nil {
			if sym, ok2 := fc.c.bindings.Idents[ident]; ok2 {
				if st, ok3 := sym.Type.(*types.Struct); ok3 {
					structType = st
				}
			}
		}

		if structType == nil {
			fc.addError(s, "cannot determine struct type for field assignment")
			return
		}

		// Get struct type info
		structInfo, ok := fc.c.structTypes[structType.Name]
		if !ok {
			fc.addError(s, "unknown struct type %q for field assignment", structType.Name)
			return
		}

		// Find field index
		fieldIdx := -1
		for i, field := range structInfo.Fields {
			if field.Name == s.Field {
				fieldIdx = i
				break
			}
		}
		if fieldIdx == -1 {
			fc.addError(s, "struct %q has no field %q", structInfo.Name, s.Field)
			return
		}

		// Load struct variable
		if slot, ok := fc.lookupLocal(ident.Name); ok {
			fc.chunk.Emit(OpLoadLocal, slot, 0)
		} else if upvalueIdx, ok := fc.lookupUpvalue(ident.Name); ok {
			fc.chunk.Emit(OpLoadUpvalue, upvalueIdx, 0)
		} else {
			fc.addError(s, "unknown variable %q", ident.Name)
			return
		}

		// Compile value expression
		fc.compileExpr(s.Value)

		// OpStoreField: pops [struct, value], pushes new struct
		fc.chunk.Emit(OpStoreField, fieldIdx, 0)

		// Store updated struct back to variable
		if slot, ok := fc.lookupLocal(ident.Name); ok {
			fc.chunk.Emit(OpStoreLocal, slot, 0)
		} else if upvalueIdx, ok := fc.lookupUpvalue(ident.Name); ok {
			fc.chunk.Emit(OpStoreUpvalue, upvalueIdx, 0)
		}

		// Pop the result (statement, not expression)
		fc.chunk.Emit(OpPop, 0, 0)
	} else {
		// Complex expression - not supported for assignment
		fc.addError(s, "can only assign to fields of struct variables, not complex expressions")
	}
}

func (fc *funcCompiler) compileBinary(b *ast.BinaryExpr) {
	fc.compileExpr(b.Left)
	fc.compileExpr(b.Right)

	switch b.Op {
	case token.Plus:
		if fc.c.bindings != nil {
			leftType := fc.c.bindings.ExprTypes[b.Left]
			rightType := fc.c.bindings.ExprTypes[b.Right]
			if types.Equal(leftType, types.String) && types.Equal(rightType, types.String) {
				fc.chunk.Emit(OpConcatString, 0, 0)
				return
			}
		}
		fc.chunk.Emit(OpAdd, 0, 0)
	case token.Minus:
		fc.chunk.Emit(OpSub, 0, 0)
	case token.Star:
		fc.chunk.Emit(OpMul, 0, 0)
	case token.Slash:
		fc.chunk.Emit(OpDiv, 0, 0)
	case token.Percent:
		fc.chunk.Emit(OpMod, 0, 0)

	case token.Lt:
		fc.chunk.Emit(OpLt, 0, 0)
	case token.LtEq:
		fc.chunk.Emit(OpLte, 0, 0)
	case token.Gt:
		fc.chunk.Emit(OpGt, 0, 0)
	case token.GtEq:
		fc.chunk.Emit(OpGte, 0, 0)

	case token.Eq:
		fc.chunk.Emit(OpEq, 0, 0)
	case token.NotEq:
		fc.chunk.Emit(OpNeq, 0, 0)

	default:
		fc.addError(b, "unsupported binary operator %s", b.Op)
	}
}

// compileLogical compiles logical operators (&& and ||) with short-circuit evaluation.
func (fc *funcCompiler) compileLogical(b *ast.BinaryExpr) {
	switch b.Op {
	case token.AndAnd:
		// Evaluate left operand
		fc.compileExpr(b.Left)
		// If false, skip right and return false
		jmpFalse := fc.chunk.Emit(OpJumpIfFalse, 0, 0)

		// Left is true, evaluate right
		fc.compileExpr(b.Right)
		jmpEnd := fc.chunk.Emit(OpJump, 0, 0)

		falsePos := len(fc.chunk.Code)
		fc.chunk.Code[jmpFalse].A = falsePos

		// Left was false: push false
		falseIdx := fc.chunk.AddConstBool(false)
		fc.chunk.Emit(OpConst, falseIdx, 0)

		endPos := len(fc.chunk.Code)
		fc.chunk.Code[jmpEnd].A = endPos

	case token.OrOr:
		// Evaluate left operand
		fc.compileExpr(b.Left)
		// If false, evaluate right; if true, return true immediately
		jmpToRight := fc.chunk.Emit(OpJumpIfFalse, 0, 0)

		trueIdx := fc.chunk.AddConstBool(true)
		fc.chunk.Emit(OpConst, trueIdx, 0)
		jmpEnd := fc.chunk.Emit(OpJump, 0, 0)

		rightPos := len(fc.chunk.Code)
		fc.chunk.Code[jmpToRight].A = rightPos

		fc.compileExpr(b.Right)

		endPos := len(fc.chunk.Code)
		fc.chunk.Code[jmpEnd].A = endPos

	default:
		fc.addError(b, "compileLogical: unsupported op %v", b.Op)
	}
}

// reorderCallArgs reorders and validates call arguments based on parameter names.
// It supports both positional and named arguments, and returns:
//   - A slice of expressions in parameter order (nil for missing parameters)
//   - A boolean slice indicating which parameters were provided
//
// Validation errors (unknown parameter names, duplicates, too many arguments) are reported.
// The caller is responsible for handling missing parameters (defaults or errors).
func (fc *funcCompiler) reorderCallArgs(call *ast.CallExpr, paramNames []string, funcName string) ([]ast.Expr, []bool) {
	nParams := len(paramNames)
	paramNameToIndex := make(map[string]int, nParams)
	for i, name := range paramNames {
		paramNameToIndex[name] = i
	}

	// Mark which params are provided, and store expressions
	provided := make([]bool, nParams)
	argExprs := make([]ast.Expr, nParams)
	positionalIndex := 0

	for _, arg := range call.Args {
		if named, ok := arg.(*ast.NamedArg); ok {
			idx, exists := paramNameToIndex[named.Name]
			if !exists {
				fc.addError(named, "function %s has no parameter named %q", funcName, named.Name)
				continue
			}
			if provided[idx] {
				fc.addError(named, "parameter %q specified multiple times", named.Name)
				continue
			}
			provided[idx] = true
			argExprs[idx] = named.Value
		} else {
			// Positional argument
			if positionalIndex >= nParams {
				fc.addError(arg, "too many arguments in call to %s", funcName)
				continue
			}
			provided[positionalIndex] = true
			argExprs[positionalIndex] = arg
			positionalIndex++
		}
	}

	// Build argument list in parameter order
	finalArgs := make([]ast.Expr, nParams)
	for i := 0; i < nParams; i++ {
		if provided[i] {
			finalArgs[i] = argExprs[i]
		}
		// Leave nil for missing parameters - caller handles defaults
	}

	return finalArgs, provided
}

func (fc *funcCompiler) compileCall(call *ast.CallExpr) {
	// Builtins by simple name
	if ident, ok := call.Callee.(*ast.IdentExpr); ok {
		if builtin := builtins.LookupByName(ident.Name); builtin != nil {
			// builtin call with named argument support
			reorderedArgs, provided := fc.reorderCallArgs(call, builtin.Meta.ParamNames, builtin.Meta.Name)

			// Builtins don't have defaults - check that all parameters are provided
			allProvided := true
			for i, prov := range provided {
				if !prov {
					fc.addError(call, "missing argument for required parameter %q", builtin.Meta.ParamNames[i])
					allProvided = false
				}
			}

			if !allProvided {
				// Still compile args to keep code consistent
				for _, arg := range call.Args {
					if named, ok := arg.(*ast.NamedArg); ok {
						fc.compileExpr(named.Value)
					} else {
						fc.compileExpr(arg)
					}
				}
				return
			}

			// Compile arguments in parameter order
			for _, arg := range reorderedArgs {
				fc.compileExpr(arg)
			}

			fc.chunk.Emit(OpCallBuiltin, int(builtin.Meta.ID), len(reorderedArgs))
			return
		}
	}

	// Resolve static functions (ident or member) via bindings
	var fnDecl *ast.FunDecl
	var fnIndex int
	var hasFn bool
	var isMethodCall bool
	var receiverExpr ast.Expr // For method calls

	if fc.c.bindings != nil {
		switch cal := call.Callee.(type) {
		case *ast.IdentExpr:
			if sym, ok := fc.c.bindings.Idents[cal]; ok && sym.Kind == types.SymFunc {
				if decl, ok2 := sym.Node.(*ast.FunDecl); ok2 {
					if idx, ok3 := fc.c.funcIndex[decl]; ok3 {
						fnDecl = decl
						fnIndex = idx
						hasFn = true
					}
				}
			}
		case *ast.MemberExpr:
			// Check if it's a built-in method (SymFunc but no Node)
			if sym, ok := fc.c.bindings.Members[cal]; ok && sym.Kind == types.SymFunc && sym.Node == nil {
				// This is a built-in method
				// Get the receiver type to determine the method
				var receiverType types.Type
				if ident, ok2 := cal.X.(*ast.IdentExpr); ok2 {
					if xSym, ok3 := fc.c.bindings.Idents[ident]; ok3 {
						receiverType = xSym.Type
					}
				} else {
					if t, ok2 := fc.c.bindings.ExprTypes[cal.X]; ok2 {
						receiverType = t
					}
				}

				// Convert receiver type to builtin TypeKind
				var typeKind builtins.TypeKind
				var found bool
				switch t := receiverType.(type) {
				case *types.Basic:
					typeKind, found = builtins.TypeKindFromString(t.Name)
				case *types.List:
					typeKind, found = builtins.TypeList, true
				case *types.Dict:
					typeKind, found = builtins.TypeDict, true
				}

				if found {
					if methodBuiltin := builtins.LookupMethod(typeKind, cal.Name); methodBuiltin != nil {
						// Compile built-in method call with named argument support
						// Build effective arguments: receiver + call.Args
						effectiveArgs := append([]ast.Expr{cal.X}, call.Args...)
						reorderedArgs, provided := fc.reorderCallArgs(&ast.CallExpr{Args: effectiveArgs}, methodBuiltin.Meta.ParamNames, methodBuiltin.Meta.Name)

						// Check that all required parameters are provided
						allProvided := true
						for i, prov := range provided {
							if !prov {
								fc.addError(call, "missing argument for required parameter %q", methodBuiltin.Meta.ParamNames[i])
								allProvided = false
							}
						}

						if !allProvided {
							// Still compile args to keep code consistent
							for _, arg := range effectiveArgs {
								fc.compileExpr(arg)
							}
							return
						}

						// Compile arguments in parameter order (receiver first)
						for _, arg := range reorderedArgs {
							if arg != nil {
								fc.compileExpr(arg)
							}
						}

						// Emit OpCallBuiltin with receiver + arguments
						// Arity for methods includes the receiver
						fc.chunk.Emit(OpCallBuiltin, int(methodBuiltin.Meta.ID), methodBuiltin.Meta.Arity)
						return
					}
				}
			}

			// Check if it's a module function
			if sym, ok := fc.c.bindings.Members[cal]; ok && sym.Kind == types.SymFunc {
				if decl, ok2 := sym.Node.(*ast.FunDecl); ok2 {
					if idx, ok3 := fc.c.funcIndex[decl]; ok3 {
						fnDecl = decl
						fnIndex = idx
						hasFn = true
					}
				}
			}
			// Check if it's a static method call (Type.method)
			if !hasFn {
				if ident, ok := cal.X.(*ast.IdentExpr); ok {
					// Check if cal.X is a type identifier
					if xSym, ok2 := fc.c.bindings.Idents[ident]; ok2 && xSym.Kind == types.SymType {
						if structType, ok3 := xSym.Type.(*types.Struct); ok3 {
							// Look up static method in method index
							if methodMap, ok4 := fc.c.methodIndex[structType.Name]; ok4 {
								if methodIdx, ok5 := methodMap[cal.Name]; ok5 {
									// Find the method's FunDecl to verify it's static
									if fc.c.world != nil {
										for _, modInfo := range fc.c.world.Modules {
											for _, fn := range modInfo.Prog.Funcs {
												if fn.Receiver != nil && fn.Name == cal.Name {
													if simpleType, ok6 := fn.Receiver.Type.(*ast.SimpleType); ok6 && simpleType.Name == structType.Name {
														if fn.Receiver.Kind == ast.ReceiverStatic {
															// This is a static method call
															fnDecl = fn
															fnIndex = methodIdx
															hasFn = true
															isMethodCall = false // Static methods don't have receiver
															break
														}
													}
												}
											}
											if hasFn {
												break
											}
										}
									}
								}
							}
						}
					}
				}
			}
			// Check if it's an instance method call (value.method)
			if !hasFn {
				// Get the type of the receiver (X in X.method)
				var xType types.Type
				if ident, ok := cal.X.(*ast.IdentExpr); ok {
					// cal.X is a variable - look it up in Idents
					if xSym, ok2 := fc.c.bindings.Idents[ident]; ok2 {
						// Skip if it's a type (already handled above)
						if xSym.Kind != types.SymType {
							xType = xSym.Type
						}
					}
				} else {
					// cal.X is an expression - look it up in ExprTypes
					if t, ok2 := fc.c.bindings.ExprTypes[cal.X]; ok2 {
						xType = t
					}
				}
				if xType != nil {
					if structType, ok2 := xType.(*types.Struct); ok2 {
						// Look up instance method in method index
						if methodMap, ok3 := fc.c.methodIndex[structType.Name]; ok3 {
							if methodIdx, ok4 := methodMap[cal.Name]; ok4 {
								// Find the method's FunDecl to verify it's instance
								if fc.c.world != nil {
									for _, modInfo := range fc.c.world.Modules {
										for _, fn := range modInfo.Prog.Funcs {
											if fn.Receiver != nil && fn.Name == cal.Name {
												// Check if receiver type matches by name
												if simpleType, ok5 := fn.Receiver.Type.(*ast.SimpleType); ok5 && simpleType.Name == structType.Name {
													if fn.Receiver.Kind == ast.ReceiverInstance {
														// This is an instance method call
														fnDecl = fn
														fnIndex = methodIdx
														hasFn = true
														isMethodCall = true
														receiverExpr = cal.X
														break
													}
												}
											}
										}
										if hasFn {
											break
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// If we know fnDecl, handle named/default args and emit OpCall
	if hasFn && fnDecl != nil {
		// For methods, prepend receiver to parameter names and arguments
		var paramNames []string
		var effectiveArgs []ast.Expr
		if isMethodCall {
			// Receiver is the first parameter
			paramNames = append([]string{fnDecl.Receiver.Name}, func() []string {
				names := make([]string, len(fnDecl.Params))
				for i, p := range fnDecl.Params {
					names[i] = p.Name
				}
				return names
			}()...)
			effectiveArgs = append([]ast.Expr{receiverExpr}, call.Args...)
		} else {
			paramNames = make([]string, len(fnDecl.Params))
			for i, param := range fnDecl.Params {
				paramNames[i] = param.Name
			}
			effectiveArgs = call.Args
		}

		// Reorder arguments using the shared helper
		reorderedArgs, provided := fc.reorderCallArgs(&ast.CallExpr{Args: effectiveArgs}, paramNames, fnDecl.Name)

		// Build final argument list with defaults
		nParams := len(paramNames) // Includes receiver for instance methods
		finalArgs := make([]ast.Expr, nParams)
		for i := 0; i < nParams; i++ {
			if provided[i] {
				// reorderedArgs[i] should never be nil if provided[i] is true
				// but check anyway for safety
				if reorderedArgs[i] != nil {
					finalArgs[i] = reorderedArgs[i]
				} else {
					// This shouldn't happen, but handle it
					fc.addError(call, "internal error: parameter %q marked as provided but expression is nil", paramNames[i])
					finalArgs[i] = &ast.IntLiteral{Value: 0, LitPos: call.Pos(), Raw: "0"} // Dummy
				}
			} else {
				// Missing parameter - use default if available
				// For instance methods, i==0 is the receiver (no default, should always be provided)
				if isMethodCall && i == 0 {
					fc.addError(call, "missing receiver argument for method call")
					finalArgs[i] = &ast.IntLiteral{Value: 0, LitPos: call.Pos(), Raw: "0"} // Dummy
				} else {
					paramIdx := i
					if isMethodCall {
						paramIdx = i - 1 // Adjust for receiver
					}
					if paramIdx >= 0 && paramIdx < len(fnDecl.Params) && fnDecl.Params[paramIdx].Default != nil {
						finalArgs[i] = fnDecl.Params[paramIdx].Default
					} else {
						fc.addError(call, "missing argument for required parameter %q", paramNames[i])
						finalArgs[i] = &ast.IntLiteral{Value: 0, LitPos: call.Pos(), Raw: "0"} // Dummy
					}
				}
			}
		}

		// Compile arguments in parameter order (receiver first for methods)
		for _, arg := range finalArgs {
			fc.compileExpr(arg)
		}

		// Direct call by index with the number of parameters (including receiver for methods)
		fc.chunk.Emit(OpCall, fnIndex, nParams)
		return
	}

	// Fallback: call-by-value (OpCallValue) and ban named args
	for _, arg := range call.Args {
		if _, ok := arg.(*ast.NamedArg); ok {
			fc.addError(call, "named arguments are only allowed when calling a function by name")
		}
		fc.compileExpr(arg)
	}

	fc.compileExpr(call.Callee)
	fc.chunk.Emit(OpCallValue, len(call.Args), 0)
}
