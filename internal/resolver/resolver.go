package resolver

import (
	"avenir/internal/ast"
)

// UpvalueInfo describes a captured variable (upvalue).
type UpvalueInfo struct {
	Name    string
	IsLocal bool // true: captures a local from immediately enclosing function
	Index   int  // slot index in enclosing function's locals OR "upvalue index" of enclosing function
}

// FunctionInfo contains metadata about a function's upvalues.
type FunctionInfo struct {
	Node     ast.Node // *ast.FunDecl or *ast.FuncLiteral
	Upvalues []UpvalueInfo
	Locals   []string // names of local variables (parameters + declared vars) in order
}

// Resolver analyzes the AST to identify captured variables (upvalues).
type Resolver struct {
	funcInfos    map[ast.Node]*FunctionInfo
	funcParents  map[ast.Node]ast.Node // function -> parent function (for walking up the chain)
	errors       []error
}

// NewResolver creates a new resolver.
func NewResolver() *Resolver {
	return &Resolver{
		funcInfos:   make(map[ast.Node]*FunctionInfo),
		funcParents: make(map[ast.Node]ast.Node),
	}
}

// Resolve analyzes the program and returns function metadata including upvalues.
func (r *Resolver) Resolve(prog *ast.Program) map[ast.Node]*FunctionInfo {
	// First pass: collect all functions and their local variables
	r.collectFunctions(prog)

	// Second pass: identify upvalues for each function
	r.identifyUpvalues(prog)

	return r.funcInfos
}

// collectFunctions walks the AST and collects function metadata.
func (r *Resolver) collectFunctions(prog *ast.Program) {
	// Collect top-level functions
	for _, fn := range prog.Funcs {
		r.collectFunction(fn, nil)
	}
}

// collectFunction collects metadata for a function and recursively processes nested functions.
func (r *Resolver) collectFunction(fn ast.Node, parent *FunctionInfo) *FunctionInfo {
	var info *FunctionInfo
	var params []string
	var body *ast.BlockStmt

	switch f := fn.(type) {
	case *ast.FunDecl:
		info = &FunctionInfo{Node: fn}
		for _, p := range f.Params {
			params = append(params, p.Name)
		}
		body = f.Body
	case *ast.FuncLiteral:
		info = &FunctionInfo{Node: fn}
		for _, p := range f.Params {
			params = append(params, p.Name)
		}
		body = f.Body
	default:
		return nil
	}

	info.Locals = append(info.Locals, params...)
	r.funcInfos[fn] = info

	// Track parent relationship
	if parent != nil {
		r.funcParents[fn] = parent.Node
	}

	// Walk body to find local variable declarations and nested functions
	r.collectLocalsAndNestedFunctions(body, info, parent)

	return info
}

// collectLocalsAndNestedFunctions walks a block to find local variables and nested functions.
func (r *Resolver) collectLocalsAndNestedFunctions(block *ast.BlockStmt, currentFunc *FunctionInfo, parentFunc *FunctionInfo) {
	if block == nil {
		return
	}

	for _, stmt := range block.Stmts {
		switch s := stmt.(type) {
		case *ast.VarDeclStmt:
			// Add to locals if not already present
			found := false
			for _, name := range currentFunc.Locals {
				if name == s.Name {
					found = true
					break
				}
			}
			if !found {
				currentFunc.Locals = append(currentFunc.Locals, s.Name)
			}

		case *ast.BlockStmt:
			// Recursively process nested blocks
			r.collectLocalsAndNestedFunctions(s, currentFunc, parentFunc)

		case *ast.IfStmt:
			if s.Then != nil {
				r.collectLocalsAndNestedFunctions(s.Then, currentFunc, parentFunc)
			}
			if s.Else != nil {
				if elseBlock, ok := s.Else.(*ast.BlockStmt); ok {
					r.collectLocalsAndNestedFunctions(elseBlock, currentFunc, parentFunc)
				}
			}

		case *ast.WhileStmt:
			if s.Body != nil {
				r.collectLocalsAndNestedFunctions(s.Body, currentFunc, parentFunc)
			}

		case *ast.ForStmt:
			if s.Body != nil {
				r.collectLocalsAndNestedFunctions(s.Body, currentFunc, parentFunc)
			}

		case *ast.ForEachStmt:
			// Loop variable is a local
			found := false
			for _, name := range currentFunc.Locals {
				if name == s.VarName {
					found = true
					break
				}
			}
			if !found {
				currentFunc.Locals = append(currentFunc.Locals, s.VarName)
			}
			if s.Body != nil {
				r.collectLocalsAndNestedFunctions(s.Body, currentFunc, parentFunc)
			}

		case *ast.TryStmt:
			if s.Body != nil {
				r.collectLocalsAndNestedFunctions(s.Body, currentFunc, parentFunc)
			}
			if s.CatchBody != nil {
				// Catch variable is a local
				found := false
				for _, name := range currentFunc.Locals {
					if name == s.CatchName {
						found = true
						break
					}
				}
				if !found {
					currentFunc.Locals = append(currentFunc.Locals, s.CatchName)
				}
				r.collectLocalsAndNestedFunctions(s.CatchBody, currentFunc, parentFunc)
			}
		}

		// Check for function literals in expressions
		r.findFunctionLiteralsInExpr(stmt, currentFunc, parentFunc)
	}
}

// findFunctionLiteralsInExpr recursively searches expressions for function literals.
func (r *Resolver) findFunctionLiteralsInExpr(node ast.Node, currentFunc *FunctionInfo, parentFunc *FunctionInfo) {
	switch n := node.(type) {
	case *ast.FuncLiteral:
		r.collectFunction(n, currentFunc)

	case *ast.ExprStmt:
		r.findFunctionLiteralsInExpr(n.Expression, currentFunc, parentFunc)

	case *ast.AssignStmt:
		r.findFunctionLiteralsInExpr(n.Value, currentFunc, parentFunc)

	case *ast.VarDeclStmt:
		if n.Value != nil {
			r.findFunctionLiteralsInExpr(n.Value, currentFunc, parentFunc)
		}

	case *ast.ReturnStmt:
		if n.Result != nil {
			r.findFunctionLiteralsInExpr(n.Result, currentFunc, parentFunc)
		}

	case *ast.IfStmt:
		r.findFunctionLiteralsInExpr(n.Cond, currentFunc, parentFunc)

	case *ast.WhileStmt:
		r.findFunctionLiteralsInExpr(n.Cond, currentFunc, parentFunc)

	case *ast.ForStmt:
		if n.Init != nil {
			r.findFunctionLiteralsInExpr(n.Init, currentFunc, parentFunc)
		}
		if n.Cond != nil {
			r.findFunctionLiteralsInExpr(n.Cond, currentFunc, parentFunc)
		}
		if n.Post != nil {
			r.findFunctionLiteralsInExpr(n.Post, currentFunc, parentFunc)
		}

	case *ast.BinaryExpr:
		r.findFunctionLiteralsInExpr(n.Left, currentFunc, parentFunc)
		r.findFunctionLiteralsInExpr(n.Right, currentFunc, parentFunc)

	case *ast.UnaryExpr:
		r.findFunctionLiteralsInExpr(n.X, currentFunc, parentFunc)

	case *ast.CallExpr:
		r.findFunctionLiteralsInExpr(n.Callee, currentFunc, parentFunc)
		for _, arg := range n.Args {
			if namedArg, ok := arg.(*ast.NamedArg); ok {
				r.findFunctionLiteralsInExpr(namedArg.Value, currentFunc, parentFunc)
			} else {
				r.findFunctionLiteralsInExpr(arg, currentFunc, parentFunc)
			}
		}

	case *ast.IndexExpr:
		r.findFunctionLiteralsInExpr(n.X, currentFunc, parentFunc)
		r.findFunctionLiteralsInExpr(n.Index, currentFunc, parentFunc)

	case *ast.ListLiteral:
		for _, el := range n.Elements {
			r.findFunctionLiteralsInExpr(el, currentFunc, parentFunc)
		}
	}
}

// identifyUpvalues identifies which variables are captured as upvalues for each function.
func (r *Resolver) identifyUpvalues(prog *ast.Program) {
	// Process top-level functions
	for _, fn := range prog.Funcs {
		r.identifyUpvaluesForFunction(fn, nil, nil)
	}
}

// identifyUpvaluesForFunction identifies upvalues for a function.
//
// Algorithm:
//   1. Collect all identifiers used in the function body.
//   2. For each identifier:
//      - If it's a local (parameter or declared variable), skip it.
//      - If it's in the parent function's locals, add as upvalue with IsLocal=true.
//      - If it's in the parent function's upvalues, add as upvalue with IsLocal=false.
//   3. Recursively process nested functions (so they can identify their upvalues).
//
// Parameters:
//   - fn: The function node to analyze
//   - currentFunc: The FunctionInfo for fn (can be nil, will be looked up)
//   - parentFunc: The FunctionInfo for the enclosing function (nil for top-level functions)
func (r *Resolver) identifyUpvaluesForFunction(fn ast.Node, currentFunc *FunctionInfo, parentFunc *FunctionInfo) {
	info := r.funcInfos[fn]
	if info == nil {
		return
	}

	var body *ast.BlockStmt
	switch f := fn.(type) {
	case *ast.FunDecl:
		body = f.Body
	case *ast.FuncLiteral:
		body = f.Body
	default:
		return
	}

	// Find all identifiers used in the function body
	usedVars := make(map[string]bool)
	r.collectUsedIdentifiers(body, usedVars)

	// For each used variable, determine if it's an upvalue
	for varName := range usedVars {
		// Check if it's a local (parameter or declared variable)
		isLocal := false
		for _, local := range info.Locals {
			if local == varName {
				isLocal = true
				break
			}
		}
		if isLocal {
			continue // Not an upvalue
		}

		// Check if it's in an enclosing function's scope by walking up the parent chain
		enclosingFunc := parentFunc
		depth := 0
		for enclosingFunc != nil {
			// Check enclosing function's locals
			for i, local := range enclosingFunc.Locals {
				if local == varName {
					if depth == 0 {
						// Direct parent has it as a local
						info.Upvalues = append(info.Upvalues, UpvalueInfo{
							Name:    varName,
							IsLocal: true,
							Index:   i,
						})
					} else {
						// It's in a grandparent or higher - we need to ensure all
						// intermediate functions capture it. We'll do this by
						// adding it to the immediate parent first, then propagating.
						// For now, add it as if it comes from parent's upvalues
						// (we'll fix this during propagation)
						if parentFunc != nil {
							// Check if parent already has it
							foundInParent := false
							for j, upv := range parentFunc.Upvalues {
								if upv.Name == varName {
									info.Upvalues = append(info.Upvalues, UpvalueInfo{
										Name:    varName,
										IsLocal: false,
										Index:   j,
									})
									foundInParent = true
									break
								}
							}
							if !foundInParent {
								// Add placeholder - will be fixed during propagation
								info.Upvalues = append(info.Upvalues, UpvalueInfo{
									Name:    varName,
									IsLocal: false,
									Index:   -1, // Placeholder
								})
							}
						}
					}
					goto found
				}
			}

			// Check enclosing function's upvalues
			for i, upv := range enclosingFunc.Upvalues {
				if upv.Name == varName {
					if depth == 0 {
						// Direct parent has it as an upvalue
						info.Upvalues = append(info.Upvalues, UpvalueInfo{
							Name:    varName,
							IsLocal: false,
							Index:   i,
						})
					} else {
						// It's in a grandparent's upvalues - add placeholder
						if parentFunc != nil {
							foundInParent := false
							for j, upv2 := range parentFunc.Upvalues {
								if upv2.Name == varName {
									info.Upvalues = append(info.Upvalues, UpvalueInfo{
										Name:    varName,
										IsLocal: false,
										Index:   j,
									})
									foundInParent = true
									break
								}
							}
							if !foundInParent {
								info.Upvalues = append(info.Upvalues, UpvalueInfo{
									Name:    varName,
									IsLocal: false,
									Index:   -1, // Placeholder
								})
							}
						}
					}
					goto found
				}
			}

			// Walk up to grandparent
			if parentNode, ok := r.funcParents[enclosingFunc.Node]; ok {
				enclosingFunc = r.funcInfos[parentNode]
				depth++
			} else {
				break
			}
		}
	found:
	}

	// Recursively process nested functions AFTER identifying upvalues for this function
	// This ensures nested functions can correctly identify their upvalues (including
	// variables from this function's scope).
	r.processNestedFunctions(body, info, parentFunc)

	// After processing nested functions, propagate upvalues: if a nested function
	// captures a variable from grandparent or higher, we need to also capture it
	// in this function to create the chain of upvalues.
	r.propagateUpvaluesFromNestedFunctions(body, info, parentFunc)
}

// propagateUpvaluesFromNestedFunctions ensures that if a nested function captures
// a variable from a grandparent or higher scope, the parent function also captures it.
func (r *Resolver) propagateUpvaluesFromNestedFunctions(block *ast.BlockStmt, currentFunc *FunctionInfo, parentFunc *FunctionInfo) {
	if block == nil || parentFunc == nil {
		return
	}

	// Find all nested function literals and check their upvalues
	for _, stmt := range block.Stmts {
		r.findNestedFunctionLiteralsAndPropagate(stmt, currentFunc, parentFunc)
	}
}

// findNestedFunctionLiteralsAndPropagate finds function literals and propagates their upvalues.
func (r *Resolver) findNestedFunctionLiteralsAndPropagate(node ast.Node, currentFunc *FunctionInfo, parentFunc *FunctionInfo) {
	switch n := node.(type) {
	case *ast.FuncLiteral:
		// Get the nested function's info
		nestedInfo := r.funcInfos[n]
		if nestedInfo != nil {
			// For each upvalue in the nested function, check if it comes from grandparent or higher
			for _, upv := range nestedInfo.Upvalues {
				// Check if this upvalue is already in currentFunc's upvalues
				found := false
				for _, currentUpv := range currentFunc.Upvalues {
					if currentUpv.Name == upv.Name {
						found = true
						break
					}
				}

				// If not found, check if it's in parentFunc's scope (grandparent or higher)
				if !found {
					// Check if it's in parent's locals (grandparent)
					for i, local := range parentFunc.Locals {
						if local == upv.Name {
							// It's in grandparent - add to currentFunc's upvalues
							currentFunc.Upvalues = append(currentFunc.Upvalues, UpvalueInfo{
								Name:    upv.Name,
								IsLocal: true,
								Index:   i,
							})
							// Update nested function's upvalue to reference currentFunc's upvalue
							for j := range nestedInfo.Upvalues {
								if nestedInfo.Upvalues[j].Name == upv.Name {
									nestedInfo.Upvalues[j] = UpvalueInfo{
										Name:    upv.Name,
										IsLocal: false,
										Index:   len(currentFunc.Upvalues) - 1,
									}
									break
								}
							}
							break
						}
					}

					// Check if it's in parent's upvalues (great-grandparent or higher)
					if !found {
						for i, parentUpv := range parentFunc.Upvalues {
							if parentUpv.Name == upv.Name {
								// It's in grandparent's upvalues - add to currentFunc's upvalues
								currentFunc.Upvalues = append(currentFunc.Upvalues, UpvalueInfo{
									Name:    upv.Name,
									IsLocal: false,
									Index:   i,
								})
								// Update nested function's upvalue to reference currentFunc's upvalue
								for j := range nestedInfo.Upvalues {
									if nestedInfo.Upvalues[j].Name == upv.Name {
										nestedInfo.Upvalues[j] = UpvalueInfo{
											Name:    upv.Name,
											IsLocal: false,
											Index:   len(currentFunc.Upvalues) - 1,
										}
										break
									}
								}
								break
							}
						}
					}
				}
			}
		}
		// Continue searching for nested function literals inside this one
		if n.Body != nil {
			r.findNestedFunctionLiteralsAndPropagate(n.Body, currentFunc, parentFunc)
		}

	case *ast.BlockStmt:
		for _, stmt := range n.Stmts {
			r.findNestedFunctionLiteralsAndPropagate(stmt, currentFunc, parentFunc)
		}

	case *ast.ExprStmt:
		r.findNestedFunctionLiteralsAndPropagate(n.Expression, currentFunc, parentFunc)

	case *ast.AssignStmt:
		r.findNestedFunctionLiteralsAndPropagate(n.Value, currentFunc, parentFunc)

	case *ast.VarDeclStmt:
		if n.Value != nil {
			r.findNestedFunctionLiteralsAndPropagate(n.Value, currentFunc, parentFunc)
		}

	case *ast.ReturnStmt:
		if n.Result != nil {
			r.findNestedFunctionLiteralsAndPropagate(n.Result, currentFunc, parentFunc)
		}

	case *ast.BinaryExpr:
		r.findNestedFunctionLiteralsAndPropagate(n.Left, currentFunc, parentFunc)
		r.findNestedFunctionLiteralsAndPropagate(n.Right, currentFunc, parentFunc)

	case *ast.UnaryExpr:
		r.findNestedFunctionLiteralsAndPropagate(n.X, currentFunc, parentFunc)

	case *ast.CallExpr:
		r.findNestedFunctionLiteralsAndPropagate(n.Callee, currentFunc, parentFunc)
		for _, arg := range n.Args {
			if namedArg, ok := arg.(*ast.NamedArg); ok {
				r.findNestedFunctionLiteralsAndPropagate(namedArg.Value, currentFunc, parentFunc)
			} else {
				r.findNestedFunctionLiteralsAndPropagate(arg, currentFunc, parentFunc)
			}
		}

	case *ast.IndexExpr:
		r.findNestedFunctionLiteralsAndPropagate(n.X, currentFunc, parentFunc)
		r.findNestedFunctionLiteralsAndPropagate(n.Index, currentFunc, parentFunc)

	case *ast.ListLiteral:
		for _, el := range n.Elements {
			r.findNestedFunctionLiteralsAndPropagate(el, currentFunc, parentFunc)
		}

	case *ast.IfStmt:
		if n.Then != nil {
			r.findNestedFunctionLiteralsAndPropagate(n.Then, currentFunc, parentFunc)
		}
		if n.Else != nil {
			r.findNestedFunctionLiteralsAndPropagate(n.Else, currentFunc, parentFunc)
		}

	case *ast.WhileStmt:
		if n.Body != nil {
			r.findNestedFunctionLiteralsAndPropagate(n.Body, currentFunc, parentFunc)
		}

	case *ast.ForStmt:
		if n.Body != nil {
			r.findNestedFunctionLiteralsAndPropagate(n.Body, currentFunc, parentFunc)
		}

	case *ast.ForEachStmt:
		if n.Body != nil {
			r.findNestedFunctionLiteralsAndPropagate(n.Body, currentFunc, parentFunc)
		}

	case *ast.TryStmt:
		if n.Body != nil {
			r.findNestedFunctionLiteralsAndPropagate(n.Body, currentFunc, parentFunc)
		}
		if n.CatchBody != nil {
			r.findNestedFunctionLiteralsAndPropagate(n.CatchBody, currentFunc, parentFunc)
		}
	}
}

// collectUsedIdentifiers collects all identifier names used in a node.
func (r *Resolver) collectUsedIdentifiers(node ast.Node, used map[string]bool) {
	switch n := node.(type) {
	case *ast.IdentExpr:
		used[n.Name] = true

	case *ast.BlockStmt:
		for _, stmt := range n.Stmts {
			r.collectUsedIdentifiers(stmt, used)
		}

	case *ast.VarDeclStmt:
		if n.Value != nil {
			r.collectUsedIdentifiers(n.Value, used)
		}

	case *ast.AssignStmt:
		r.collectUsedIdentifiers(n.Value, used)

	case *ast.ExprStmt:
		r.collectUsedIdentifiers(n.Expression, used)

	case *ast.ReturnStmt:
		if n.Result != nil {
			r.collectUsedIdentifiers(n.Result, used)
		}

	case *ast.IfStmt:
		r.collectUsedIdentifiers(n.Cond, used)
		if n.Then != nil {
			r.collectUsedIdentifiers(n.Then, used)
		}
		if n.Else != nil {
			r.collectUsedIdentifiers(n.Else, used)
		}

	case *ast.WhileStmt:
		r.collectUsedIdentifiers(n.Cond, used)
		if n.Body != nil {
			r.collectUsedIdentifiers(n.Body, used)
		}

	case *ast.ForStmt:
		if n.Init != nil {
			r.collectUsedIdentifiers(n.Init, used)
		}
		if n.Cond != nil {
			r.collectUsedIdentifiers(n.Cond, used)
		}
		if n.Post != nil {
			r.collectUsedIdentifiers(n.Post, used)
		}
		if n.Body != nil {
			r.collectUsedIdentifiers(n.Body, used)
		}

	case *ast.ForEachStmt:
		r.collectUsedIdentifiers(n.ListExpr, used)
		if n.Body != nil {
			r.collectUsedIdentifiers(n.Body, used)
		}

	case *ast.TryStmt:
		if n.Body != nil {
			r.collectUsedIdentifiers(n.Body, used)
		}
		if n.CatchBody != nil {
			r.collectUsedIdentifiers(n.CatchBody, used)
		}

	case *ast.ThrowStmt:
		r.collectUsedIdentifiers(n.Expr, used)

	case *ast.BinaryExpr:
		r.collectUsedIdentifiers(n.Left, used)
		r.collectUsedIdentifiers(n.Right, used)

	case *ast.UnaryExpr:
		r.collectUsedIdentifiers(n.X, used)

	case *ast.CallExpr:
		r.collectUsedIdentifiers(n.Callee, used)
		for _, arg := range n.Args {
			if namedArg, ok := arg.(*ast.NamedArg); ok {
				r.collectUsedIdentifiers(namedArg.Value, used)
			} else {
				r.collectUsedIdentifiers(arg, used)
			}
		}

	case *ast.IndexExpr:
		r.collectUsedIdentifiers(n.X, used)
		r.collectUsedIdentifiers(n.Index, used)

	case *ast.ListLiteral:
		for _, el := range n.Elements {
			r.collectUsedIdentifiers(el, used)
		}

	case *ast.FuncLiteral:
		// Don't process function literal body here - nested functions are processed
		// separately in identifyUpvaluesForFunction. We only collect identifiers
		// that are used in the function literal's body for the current function's
		// upvalue analysis, not for the nested function itself.
		// Note: This means function literals don't contribute their internal
		// identifiers to the parent's used set, which is correct.
	}
}

// processNestedFunctions processes nested functions within a block.
func (r *Resolver) processNestedFunctions(block *ast.BlockStmt, currentFunc *FunctionInfo, parentFunc *FunctionInfo) {
	if block == nil {
		return
	}

	for _, stmt := range block.Stmts {
		r.findAndProcessFunctionLiterals(stmt, currentFunc, parentFunc)
	}
}

// findAndProcessFunctionLiterals finds and processes function literals in a statement.
func (r *Resolver) findAndProcessFunctionLiterals(node ast.Node, currentFunc *FunctionInfo, parentFunc *FunctionInfo) {
	switch n := node.(type) {
	case *ast.FuncLiteral:
		r.identifyUpvaluesForFunction(n, r.funcInfos[n], currentFunc)

	case *ast.ExprStmt:
		r.findAndProcessFunctionLiterals(n.Expression, currentFunc, parentFunc)

	case *ast.AssignStmt:
		r.findAndProcessFunctionLiterals(n.Value, currentFunc, parentFunc)

	case *ast.VarDeclStmt:
		if n.Value != nil {
			r.findAndProcessFunctionLiterals(n.Value, currentFunc, parentFunc)
		}

	case *ast.ReturnStmt:
		if n.Result != nil {
			r.findAndProcessFunctionLiterals(n.Result, currentFunc, parentFunc)
		}

	case *ast.BinaryExpr:
		r.findAndProcessFunctionLiterals(n.Left, currentFunc, parentFunc)
		r.findAndProcessFunctionLiterals(n.Right, currentFunc, parentFunc)

	case *ast.UnaryExpr:
		r.findAndProcessFunctionLiterals(n.X, currentFunc, parentFunc)

	case *ast.CallExpr:
		r.findAndProcessFunctionLiterals(n.Callee, currentFunc, parentFunc)
		for _, arg := range n.Args {
			if namedArg, ok := arg.(*ast.NamedArg); ok {
				r.findAndProcessFunctionLiterals(namedArg.Value, currentFunc, parentFunc)
			} else {
				r.findAndProcessFunctionLiterals(arg, currentFunc, parentFunc)
			}
		}

	case *ast.IndexExpr:
		r.findAndProcessFunctionLiterals(n.X, currentFunc, parentFunc)
		r.findAndProcessFunctionLiterals(n.Index, currentFunc, parentFunc)

	case *ast.ListLiteral:
		for _, el := range n.Elements {
			r.findAndProcessFunctionLiterals(el, currentFunc, parentFunc)
		}

	case *ast.BlockStmt:
		for _, stmt := range n.Stmts {
			r.findAndProcessFunctionLiterals(stmt, currentFunc, parentFunc)
		}

	case *ast.IfStmt:
		if n.Then != nil {
			r.findAndProcessFunctionLiterals(n.Then, currentFunc, parentFunc)
		}
		if n.Else != nil {
			r.findAndProcessFunctionLiterals(n.Else, currentFunc, parentFunc)
		}

	case *ast.WhileStmt:
		if n.Body != nil {
			r.findAndProcessFunctionLiterals(n.Body, currentFunc, parentFunc)
		}

	case *ast.ForStmt:
		if n.Body != nil {
			r.findAndProcessFunctionLiterals(n.Body, currentFunc, parentFunc)
		}

	case *ast.ForEachStmt:
		if n.Body != nil {
			r.findAndProcessFunctionLiterals(n.Body, currentFunc, parentFunc)
		}

	case *ast.TryStmt:
		if n.Body != nil {
			r.findAndProcessFunctionLiterals(n.Body, currentFunc, parentFunc)
		}
		if n.CatchBody != nil {
			r.findAndProcessFunctionLiterals(n.CatchBody, currentFunc, parentFunc)
		}
	}
}
