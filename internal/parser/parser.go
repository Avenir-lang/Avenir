package parser

import (
	"fmt"
	"strconv"

	"avenir/internal/ast"
	"avenir/internal/lexer"
	"avenir/internal/token"
)

type Parser struct {
	l *lexer.Lexer

	cur  token.Token
	peek token.Token

	errors []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	// init cur/peek
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.cur = p.peek
	p.peek = p.l.NextToken()
}

func (p *Parser) errorf(pos token.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf("%d:%d: ", pos.Line, pos.Column) + fmt.Sprintf(format, args...)
	p.errors = append(p.errors, msg)
}

func (p *Parser) expect(kind token.Kind) token.Token {
	if p.cur.Kind != kind {
		p.errorf(p.cur.Pos, "expected %s, got %s (%q)", kind, p.cur.Kind, p.cur.Lexeme)
	}
	tok := p.cur
	p.nextToken()
	return tok
}

// ---------- Top-level ----------

func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{}

	// pckg ...
	if p.cur.Kind == token.Pckg {
		prog.Package = p.parsePackage()
	} else {
		p.errorf(p.cur.Pos, "program must start with 'pckg'")
	}

	// imports (zero or more)
	for p.cur.Kind == token.Import {
		imp := p.parseImportDecl()
		if imp != nil {
			prog.Imports = append(prog.Imports, imp)
		}
	}

	// functions (zero or more)
	for p.cur.Kind != token.EOF {
		if p.cur.Kind == token.Fun || p.cur.Kind == token.Pub {
			if p.cur.Kind == token.Pub && (p.peek.Kind == token.Struct || p.peek.Kind == token.Mut || p.peek.Kind == token.Interface) {
				// pub struct, pub mut struct, or pub interface declaration
				isPublic := true
				p.nextToken() // consume pub
				if p.cur.Kind == token.Interface {
					interfaceDecl := p.parseInterfaceDecl(isPublic)
					if interfaceDecl != nil {
						prog.Interfaces = append(prog.Interfaces, interfaceDecl)
					}
				} else {
					isMutable := false
					if p.cur.Kind == token.Mut {
						isMutable = true
						p.nextToken() // consume mut
					}
					structDecl := p.parseStructDecl(isPublic, isMutable)
					if structDecl != nil {
						prog.Structs = append(prog.Structs, structDecl)
					}
				}
			} else {
				fn := p.parseFunDecl()
				if fn != nil {
					prog.Funcs = append(prog.Funcs, fn)
				}
			}
		} else if p.cur.Kind == token.Mut && p.peek.Kind == token.Struct {
			// mut struct declaration
			p.nextToken() // consume mut
			structDecl := p.parseStructDecl(false, true)
			if structDecl != nil {
				prog.Structs = append(prog.Structs, structDecl)
			}
		} else if p.cur.Kind == token.Struct {
			structDecl := p.parseStructDecl(false, false)
			if structDecl != nil {
				prog.Structs = append(prog.Structs, structDecl)
			}
		} else if p.cur.Kind == token.Interface {
			interfaceDecl := p.parseInterfaceDecl(false)
			if interfaceDecl != nil {
				prog.Interfaces = append(prog.Interfaces, interfaceDecl)
			}
		} else {
			p.errorf(p.cur.Pos, "unexpected token at top level: %s", p.cur.Kind)
			p.nextToken()
		}
	}

	return prog
}

func (p *Parser) parseStructDecl(isPublic bool, isMutable bool) *ast.StructDecl {
	structTok := p.cur
	if structTok.Kind != token.Struct {
		p.errorf(structTok.Pos, "expected 'struct', got %s", structTok.Kind)
		return nil
	}
	p.nextToken()

	if p.cur.Kind != token.Ident {
		p.errorf(p.cur.Pos, "expected struct name after 'struct'")
		return nil
	}
	nameTok := p.cur
	p.nextToken()

	p.expect(token.LBrace)

	var fields []*ast.FieldDecl
	for p.cur.Kind != token.RBrace && p.cur.Kind != token.EOF {
		// Check for optional 'pub' and/or 'mut' before field name
		// Order: pub mut field, pub field, mut field, field
		isFieldPublic := false
		isFieldMutable := false

		if p.cur.Kind == token.Pub {
			isFieldPublic = true
			p.nextToken()
			if p.cur.Kind == token.Mut {
				isFieldMutable = true
				p.nextToken()
			}
		} else if p.cur.Kind == token.Mut {
			isFieldMutable = true
			p.nextToken()
		}

		if p.cur.Kind != token.Ident {
			p.errorf(p.cur.Pos, "expected field name")
			break
		}
		fieldNameTok := p.cur
		p.nextToken()

		p.expect(token.Pipe)
		fieldType := p.parseType()

		// Parse optional default value
		var defaultExpr ast.Expr
		if p.cur.Kind == token.Assign {
			p.nextToken() // consume '='
			defaultExpr = p.parseExpr()
		}

		fields = append(fields, &ast.FieldDecl{
			Name:        fieldNameTok.Lexeme,
			NamePos:     fieldNameTok.Pos,
			Type:        fieldType,
			IsPublic:    isFieldPublic,
			IsMutable:   isFieldMutable,
			DefaultExpr: defaultExpr,
		})

		// Optional semicolon (for consistency with other declarations)
		if p.cur.Kind == token.Semicolon {
			p.nextToken()
		}
	}

	p.expect(token.RBrace)

	return &ast.StructDecl{
		Name:      nameTok.Lexeme,
		NamePos:   nameTok.Pos,
		Fields:    fields,
		IsPublic:  isPublic,
		IsMutable: isMutable,
	}
}

func (p *Parser) parseInterfaceDecl(isPublic bool) *ast.InterfaceDecl {
	interfaceTok := p.cur
	if interfaceTok.Kind != token.Interface {
		p.errorf(interfaceTok.Pos, "expected 'interface', got %s", interfaceTok.Kind)
		return nil
	}
	p.nextToken()

	if p.cur.Kind != token.Ident {
		p.errorf(p.cur.Pos, "expected interface name after 'interface'")
		return nil
	}
	nameTok := p.cur
	p.nextToken()

	p.expect(token.LBrace)

	var methods []*ast.InterfaceMethod
	for p.cur.Kind != token.RBrace && p.cur.Kind != token.EOF {
		// Parse method signature: fun methodName(param1 | type1, param2 | type2) | returnType
		if p.cur.Kind != token.Fun {
			p.errorf(p.cur.Pos, "expected 'fun' for interface method")
			break
		}
		p.nextToken()

		if p.cur.Kind != token.Ident {
			p.errorf(p.cur.Pos, "expected method name")
			break
		}
		methodNameTok := p.cur
		p.nextToken()

		p.expect(token.LParen)

		var paramTypes []ast.TypeNode
		if p.cur.Kind != token.RParen {
			for {
				// Interface methods have parameter names and types: name | type
				// But we only need the types for the interface signature
				if p.cur.Kind != token.Ident {
					p.errorf(p.cur.Pos, "expected parameter name")
					break
				}
				// Skip parameter name
				p.nextToken()
				p.expect(token.Pipe)
				paramType := p.parseType()
				paramTypes = append(paramTypes, paramType)
				if p.cur.Kind == token.Comma {
					p.nextToken()
					continue
				}
				break
			}
		}
		p.expect(token.RParen)

		p.expect(token.Pipe)
		returnType := p.parseType()

		methods = append(methods, &ast.InterfaceMethod{
			Name:       methodNameTok.Lexeme,
			NamePos:    methodNameTok.Pos,
			ParamTypes: paramTypes,
			Return:     returnType,
		})

		// Optional semicolon
		if p.cur.Kind == token.Semicolon {
			p.nextToken()
		}
	}

	p.expect(token.RBrace)

	return &ast.InterfaceDecl{
		Name:     nameTok.Lexeme,
		NamePos:  nameTok.Pos,
		Methods:  methods,
		IsPublic: isPublic,
	}
}

func (p *Parser) parsePackage() *ast.PackageDecl {
	pckgTok := p.cur
	p.nextToken()

	if p.cur.Kind != token.Ident {
		p.errorf(p.cur.Pos, "expected package name after 'pckg'")
		return nil
	}

	// Parse dotted package name: ident ('.' ident)*
	var parts []string
	parts = append(parts, p.cur.Lexeme)
	p.nextToken()

	for p.cur.Kind == token.Dot {
		p.nextToken()
		if p.cur.Kind != token.Ident {
			p.errorf(p.cur.Pos, "expected identifier after '.' in package name")
			return nil
		}
		parts = append(parts, p.cur.Lexeme)
		p.nextToken()
	}

	// Join parts with dots
	name := ""
	for i, part := range parts {
		if i > 0 {
			name += "."
		}
		name += part
	}

	if p.cur.Kind != token.Semicolon {
		p.errorf(p.cur.Pos, "expected ';' after package name")
	} else {
		p.nextToken()
	}

	return &ast.PackageDecl{
		Name:    name,
		NamePos: pckgTok.Pos,
	}
}

func (p *Parser) parseImportDecl() *ast.ImportDecl {
	importTok := p.cur
	p.nextToken()

	// Parse path: ident ('.' ident)*
	var path []string
	if p.cur.Kind != token.Ident {
		p.errorf(p.cur.Pos, "expected identifier after 'import'")
		return nil
	}
	path = append(path, p.cur.Lexeme)
	p.nextToken()

	for p.cur.Kind == token.Dot {
		p.nextToken()
		if p.cur.Kind != token.Ident {
			p.errorf(p.cur.Pos, "expected identifier after '.' in import path")
			return nil
		}
		path = append(path, p.cur.Lexeme)
		p.nextToken()
	}

	// Optional 'as alias'
	var alias string
	if p.cur.Kind == token.Ident && p.cur.Lexeme == "as" {
		p.nextToken()
		if p.cur.Kind != token.Ident {
			p.errorf(p.cur.Pos, "expected identifier after 'as'")
			return nil
		}
		alias = p.cur.Lexeme
		p.nextToken()
	}

	// Require semicolon
	if p.cur.Kind != token.Semicolon {
		p.errorf(p.cur.Pos, "expected ';' after import")
	} else {
		p.nextToken()
	}

	return &ast.ImportDecl{
		ImportPos: importTok.Pos,
		Path:      path,
		Alias:     alias,
	}
}

func (p *Parser) parseFunDecl() *ast.FunDecl {
	// Optional 'pub' keyword
	isPublic := false
	if p.cur.Kind == token.Pub {
		isPublic = true
		p.nextToken()
	}

	// Must have 'fun' after 'pub' (or just 'fun' if no 'pub')
	if p.cur.Kind != token.Fun {
		p.errorf(p.cur.Pos, "expected 'fun' after 'pub'")
		return nil
	}
	funTok := p.cur
	p.nextToken()

	// Check for method receiver patterns:
	// 1. Instance method: (name | Type).methodName
	// 2. Static method: Type.methodName
	var receiver *ast.Receiver

	// First, check for static method: Type.methodName
	if p.cur.Kind == token.Ident && p.peek.Kind == token.Dot {
		// This could be a static method: Type.methodName
		// Save state for potential backtrack
		savedCur := p.cur
		savedPeek := p.peek

		typeNameTok := p.cur
		p.nextToken() // consume type name
		p.nextToken() // consume '.'

		// Check if next token is an identifier (method name)
		if p.cur.Kind == token.Ident {
			// Confirmed: Type.methodName (static method)
			// Parse the type name as a SimpleType
			receiver = &ast.Receiver{
				Kind:    ast.ReceiverStatic,
				Name:    "", // Static methods have no receiver variable
				NamePos: typeNameTok.Pos,
				Type:    &ast.SimpleType{Name: typeNameTok.Lexeme, NamePos: typeNameTok.Pos},
			}
			// Don't consume the method name yet - it will be parsed below
		} else {
			// Not a static method, restore state
			p.cur = savedCur
			p.peek = savedPeek
		}
	}

	// If not a static method, check for instance method: (name | Type).methodName
	if receiver == nil && p.cur.Kind == token.LParen && p.peek.Kind == token.Ident {
		// Save current state for potential backtrack
		savedCur := p.cur
		savedPeek := p.peek

		// Try to parse as method receiver
		p.nextToken() // consume '('
		if p.cur.Kind == token.Ident {
			recvNameTok := p.cur
			recvName := p.cur.Lexeme
			p.nextToken()
			if p.cur.Kind == token.Pipe {
				// This looks like (name | Type) - likely a receiver
				p.nextToken()
				recvType := p.parseType()
				if p.cur.Kind == token.RParen {
					p.nextToken()
					if p.cur.Kind == token.Dot {
						// Confirmed: (name | Type).methodName (instance method)
						receiver = &ast.Receiver{
							Kind:    ast.ReceiverInstance,
							Name:    recvName,
							NamePos: recvNameTok.Pos,
							Type:    recvType,
						}
						p.nextToken() // consume '.'
					} else {
						// Not a receiver, restore state
						p.cur = savedCur
						p.peek = savedPeek
					}
				} else {
					// Not a receiver, restore state
					p.cur = savedCur
					p.peek = savedPeek
				}
			} else {
				// Not a receiver, restore state
				p.cur = savedCur
				p.peek = savedPeek
			}
		} else {
			// Not a receiver, restore state
			p.cur = savedCur
			p.peek = savedPeek
		}
	}

	// Parse function/method name
	if p.cur.Kind != token.Ident {
		p.errorf(p.cur.Pos, "expected function or method name after 'fun'")
		return nil
	}
	nameTok := p.cur
	p.nextToken()

	p.expect(token.LParen)

	var params []*ast.Param
	if p.cur.Kind != token.RParen {
		for {
			if p.cur.Kind != token.Ident {
				p.errorf(p.cur.Pos, "expected parameter name")
				break
			}
			paramNameTok := p.cur
			p.nextToken()

			p.expect(token.Pipe)
			paramType := p.parseType()

			// Parse optional default value
			var defaultExpr ast.Expr
			if p.cur.Kind == token.Assign {
				p.nextToken() // consume '='
				defaultExpr = p.parseExpr()
			}

			params = append(params, &ast.Param{
				Name:    paramNameTok.Lexeme,
				NamePos: paramNameTok.Pos,
				Type:    paramType,
				Default: defaultExpr,
			})

			if p.cur.Kind == token.Comma {
				p.nextToken()
				continue
			}
			break
		}
	}

	p.expect(token.RParen)

	p.expect(token.Pipe)
	retType := p.parseType()

	body := p.parseBlock()

	return &ast.FunDecl{
		Name:     nameTok.Lexeme,
		NamePos:  funTok.Pos,
		Receiver: receiver,
		Params:   params,
		Return:   retType,
		Body:     body,
		IsPublic: isPublic,
	}
}

func (p *Parser) parseFuncLiteral() ast.Expr {
	funTok := p.cur
	p.nextToken()

	p.expect(token.LParen)

	var params []*ast.Param
	if p.cur.Kind != token.RParen {
		for {
			if p.cur.Kind != token.Ident {
				p.errorf(p.cur.Pos, "expected parameter name")
				break
			}
			paramNameTok := p.cur
			p.nextToken()

			p.expect(token.Pipe)
			paramType := p.parseType()

			// Parse optional default value
			var defaultExpr ast.Expr
			if p.cur.Kind == token.Assign {
				p.nextToken() // consume '='
				defaultExpr = p.parseExpr()
			}

			params = append(params, &ast.Param{
				Name:    paramNameTok.Lexeme,
				NamePos: paramNameTok.Pos,
				Type:    paramType,
				Default: defaultExpr,
			})

			if p.cur.Kind == token.Comma {
				p.nextToken()
				continue
			}
			break
		}
	}

	p.expect(token.RParen)

	p.expect(token.Pipe)
	res := p.parseType()

	body := p.parseBlock()

	return &ast.FuncLiteral{
		FunPos: funTok.Pos,
		Params: params,
		Return: res,
		Body:   body,
	}
}

// ---------- Types ----------

func (p *Parser) parseType() ast.TypeNode {
	switch p.cur.Kind {
	case token.Lt:
		// Union type: <T1|T2|T3>
		ltTok := p.cur
		p.nextToken()

		var variants []ast.TypeNode

		// Parse first variant
		firstVariant := p.parseType()
		variants = append(variants, firstVariant)

		// Parse additional variants separated by |
		// At least one | is required for a union (must have at least 2 variants)
		if p.cur.Kind != token.Pipe {
			p.errorf(p.cur.Pos, "expected '|' after first variant in union type")
			// Try to recover by expecting >
			if p.cur.Kind == token.Gt {
				p.nextToken()
			}
			unionType := &ast.UnionType{
				UnionPos: ltTok.Pos,
				Variants: variants,
			}
			if p.cur.Kind == token.Question {
				qPos := p.cur.Pos
				p.nextToken()
				return &ast.OptionalType{
					Inner:    unionType,
					QMarkPos: qPos,
				}
			}
			return unionType
		}

		// Parse remaining variants
		for p.cur.Kind == token.Pipe {
			p.nextToken() // consume |
			variant := p.parseType()
			variants = append(variants, variant)
		}

		// Expect closing >
		if p.cur.Kind != token.Gt {
			p.errorf(p.cur.Pos, "expected '>' at end of union type")
		} else {
			p.nextToken()
		}
		unionType := &ast.UnionType{
			UnionPos: ltTok.Pos,
			Variants: variants,
		}
		if p.cur.Kind == token.Question {
			qPos := p.cur.Pos
			p.nextToken()
			return &ast.OptionalType{
				Inner:    unionType,
				QMarkPos: qPos,
			}
		}
		return unionType

	case token.IntType, token.FloatType, token.StringType, token.BoolType, token.VoidType, token.AnyType, token.ErrorType, token.BytesType:
		tok := p.cur
		p.nextToken()
		simpleType := &ast.SimpleType{
			Name:    tok.Lexeme,
			NamePos: tok.Pos,
		}
		if p.cur.Kind == token.Question {
			qPos := p.cur.Pos
			p.nextToken()
			return &ast.OptionalType{
				Inner:    simpleType,
				QMarkPos: qPos,
			}
		}
		return simpleType

	case token.ListType:
		listTok := p.cur
		p.nextToken()
		if p.cur.Kind != token.Lt {
			p.errorf(p.cur.Pos, "expected '<' after 'list'")
		} else {
			p.nextToken()
		}

		var elems []ast.TypeNode
		if p.cur.Kind != token.Gt {
			for {
				t := p.parseType()
				elems = append(elems, t)
				if p.cur.Kind == token.Comma {
					p.nextToken()
					continue
				}
				break
			}
		}

		if p.cur.Kind != token.Gt {
			p.errorf(p.cur.Pos, "expected '>' at end of list type")
		} else {
			p.nextToken()
		}

		listType := &ast.ListType{
			ListPos:      listTok.Pos,
			ElementTypes: elems,
		}
		// Check for optional: list<T>?
		if p.cur.Kind == token.Question {
			qPos := p.cur.Pos
			p.nextToken()
			return &ast.OptionalType{
				Inner:    listType,
				QMarkPos: qPos,
			}
		}
		return listType
	case token.DictType:
		dictTok := p.cur
		p.nextToken()
		if p.cur.Kind != token.Lt {
			p.errorf(p.cur.Pos, "expected '<' after 'dict'")
		} else {
			p.nextToken()
		}
		valueType := p.parseType()
		if p.cur.Kind == token.Comma {
			p.errorf(p.cur.Pos, "dict expects a single value type")
			for p.cur.Kind == token.Comma {
				p.nextToken()
				_ = p.parseType()
			}
		}
		if p.cur.Kind != token.Gt {
			p.errorf(p.cur.Pos, "expected '>' at end of dict type")
		} else {
			p.nextToken()
		}
		dictType := &ast.DictType{
			DictPos:   dictTok.Pos,
			ValueType: valueType,
		}
		if p.cur.Kind == token.Question {
			qPos := p.cur.Pos
			p.nextToken()
			return &ast.OptionalType{
				Inner:    dictType,
				QMarkPos: qPos,
			}
		}
		return dictType
	case token.Fun:
		funTok := p.cur
		p.nextToken()

		p.expect(token.LParen)
		var paramTypes []ast.TypeNode
		if p.cur.Kind != token.RParen {
			for {
				pt := p.parseType()
				paramTypes = append(paramTypes, pt)
				if p.cur.Kind == token.Comma {
					p.nextToken()
					continue
				}
				break
			}
		}
		p.expect(token.RParen)
		p.expect(token.Pipe)
		res := p.parseType()

		funcType := &ast.FuncType{
			FunPos:     funTok.Pos,
			ParamTypes: paramTypes,
			Result:     res,
		}
		// Check for optional: fun(...) | T?
		if p.cur.Kind == token.Question {
			qPos := p.cur.Pos
			p.nextToken()
			return &ast.OptionalType{
				Inner:    funcType,
				QMarkPos: qPos,
			}
		}
		return funcType

	case token.Ident:
		typ := p.parseQualifiedType()
		if p.cur.Kind == token.Question {
			qPos := p.cur.Pos
			p.nextToken()
			return &ast.OptionalType{
				Inner:    typ,
				QMarkPos: qPos,
			}
		}
		return typ
	default:
		p.errorf(p.cur.Pos, "expected type, got %s", p.cur.Kind)
		tok := p.cur
		p.nextToken()
		return &ast.SimpleType{
			Name:    "error",
			NamePos: tok.Pos,
		}
	}
}

func (p *Parser) parseQualifiedType() ast.TypeNode {
	startTok := p.cur
	path := []string{startTok.Lexeme}
	p.nextToken()

	for p.cur.Kind == token.Dot {
		p.nextToken()
		if p.cur.Kind != token.Ident {
			p.errorf(p.cur.Pos, "expected identifier after '.' in qualified type")
			break
		}
		path = append(path, p.cur.Lexeme)
		p.nextToken()
	}

	if len(path) == 1 {
		return &ast.SimpleType{
			Name:    path[0],
			NamePos: startTok.Pos,
		}
	}
	return &ast.QualifiedType{
		Path:    path,
		PathPos: startTok.Pos,
	}
}

// ---------- Blocks & statements ----------

func (p *Parser) parseBlock() *ast.BlockStmt {
	lbrace := p.expect(token.LBrace)

	block := &ast.BlockStmt{
		LBrace: lbrace.Pos,
	}

	for p.cur.Kind != token.RBrace && p.cur.Kind != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Stmts = append(block.Stmts, stmt)
		} else {
			// infinite cycle block
			p.nextToken()
		}
	}

	if p.cur.Kind == token.RBrace {
		block.RBrace = p.cur.Pos
		p.nextToken()
	} else {
		p.errorf(p.cur.Pos, "expected '}' to close block")
	}

	return block
}

func (p *Parser) parseStatement() ast.Stmt {
	switch p.cur.Kind {
	case token.Var:
		return p.parseVarDeclStmt()
	case token.If:
		return p.parseIfStmt()
	case token.While:
		return p.parseWhileStmt()
	case token.For:
		return p.parseForStmt()
	case token.Try:
		return p.parseTryStmt()
	case token.Throw:
		return p.parseThrowStmt()
	case token.Return:
		return p.parseReturnStmt()
	case token.Break:
		return p.parseBreakStmt()
	case token.LBrace:
		return p.parseBlock()
	case token.Semicolon:
		p.nextToken()
		return nil
	default:
		// Try to parse struct field assignment (expr.field = value)
		// or regular assignment (ident = value) or expr-stmt
		// We need to peek ahead to see if it's a member expression followed by assignment
		if p.cur.Kind == token.Ident && p.peek.Kind == token.Dot {
			// Could be struct field assignment: ident.field = ...
			// Save current state to backtrack if needed
			savedCur := p.cur
			p.nextToken() // consume ident, cur is now dot, peek is now ident after dot
			p.nextToken() // consume dot, cur is now ident after dot, peek is now token after that
			if p.cur.Kind == token.Ident && p.peek.Kind == token.Assign {
				// This is struct field assignment: ident.field = ...
				structIdent := savedCur
				fieldName := p.cur
				p.nextToken() // consume field name
				assignPos := p.cur.Pos
				p.expect(token.Assign)
				value := p.parseExpr()
				p.expect(token.Semicolon)
				return &ast.StructFieldAssignStmt{
					Struct: &ast.IdentExpr{
						Name:    structIdent.Lexeme,
						NamePos: structIdent.Pos,
					},
					Field:     fieldName.Lexeme,
					FieldPos:  fieldName.Pos,
					Value:     value,
					AssignPos: assignPos,
				}
			}
			// Not a struct field assignment - we've consumed tokens, so we need to
			// reconstruct the expression. We consumed: ident, dot, ident
			// So we need to create a MemberExpr and continue parsing from there
			// Create a MemberExpr for what we've consumed so far
			expr := ast.Expr(&ast.MemberExpr{
				X: &ast.IdentExpr{
					Name:    savedCur.Lexeme,
					NamePos: savedCur.Pos,
				},
				Name:    p.cur.Lexeme,
				NamePos: p.cur.Pos,
			})
			p.nextToken() // consume the ident we're currently on
			// Continue parsing postfix operations (calls, indexing, etc.)
			for {
				switch p.cur.Kind {
				case token.LParen:
					// function call
					lparen := p.cur
					p.nextToken()
					var args []ast.Expr
					if p.cur.Kind != token.RParen {
						for {
							var arg ast.Expr
							if p.cur.Kind == token.Ident && p.peek.Kind == token.Assign {
								nameTok := p.cur
								p.nextToken()
								p.expect(token.Assign)
								valueExpr := p.parseExpr()
								arg = &ast.NamedArg{
									Name:    nameTok.Lexeme,
									NamePos: nameTok.Pos,
									Value:   valueExpr,
								}
							} else {
								arg = p.parseExpr()
							}
							args = append(args, arg)
							if p.cur.Kind == token.Comma {
								p.nextToken()
								continue
							}
							break
						}
					}
					rparen := p.expect(token.RParen)
					expr = &ast.CallExpr{
						Callee: expr,
						LParen: lparen.Pos,
						Args:   args,
						RParen: rparen.Pos,
					}
				case token.LBracket:
					// list indexing
					lbr := p.cur
					p.nextToken()
					indexExpr := p.parseExpr()
					rbr := p.expect(token.RBracket)
					expr = &ast.IndexExpr{
						X:        expr,
						LBracket: lbr.Pos,
						Index:    indexExpr,
						RBracket: rbr.Pos,
					}
				default:
					// Done with postfix operations
					p.expect(token.Semicolon)
					return &ast.ExprStmt{Expression: expr}
				}
			}
		}
		// assignment or expr-stmt
		if p.cur.Kind == token.Ident && p.peek.Kind == token.Assign {
			return p.parseAssignStmt()
		}
		expr := p.parseExpr()
		p.expect(token.Semicolon)
		return &ast.ExprStmt{Expression: expr}
	}
}

func (p *Parser) parseVarDeclStmt() ast.Stmt {
	varTok := p.cur
	p.nextToken()

	if p.cur.Kind != token.Ident {
		p.errorf(p.cur.Pos, "expected variable name after 'var'")
		return nil
	}
	nameTok := p.cur
	p.nextToken()

	p.expect(token.Pipe)
	typ := p.parseType()

	p.expect(token.Assign)
	value := p.parseExpr()
	p.expect(token.Semicolon)

	return &ast.VarDeclStmt{
		VarPos:  varTok.Pos,
		Name:    nameTok.Lexeme,
		NamePos: nameTok.Pos,
		Type:    typ,
		Value:   value,
	}
}

func (p *Parser) parseAssignStmt() ast.Stmt {
	nameTok := p.cur
	p.nextToken()
	p.expect(token.Assign)
	value := p.parseExpr()
	p.expect(token.Semicolon)

	return &ast.AssignStmt{
		Name:    nameTok.Lexeme,
		NamePos: nameTok.Pos,
		Value:   value,
	}
}

func (p *Parser) parseReturnStmt() ast.Stmt {
	retTok := p.cur
	p.nextToken()

	var result ast.Expr
	if p.cur.Kind != token.Semicolon {
		result = p.parseExpr()
	}
	p.expect(token.Semicolon)

	return &ast.ReturnStmt{
		ReturnPos: retTok.Pos,
		Result:    result,
	}
}

func (p *Parser) parseIfStmt() ast.Stmt {
	ifTok := p.cur
	p.nextToken()
	p.expect(token.LParen)

	// first part condition
	cond := p.parseExpr()

	// sugar: if (a > 0; a < 10; flag)
	for p.cur.Kind == token.Semicolon {
		p.nextToken()
		rhs := p.parseExpr()
		cond = &ast.BinaryExpr{
			OpPos: cond.Pos(),
			Op:    token.AndAnd,
			Left:  cond,
			Right: rhs,
		}
	}

	p.expect(token.RParen)

	thenBlock := p.parseBlock()

	var elseStmt ast.Stmt
	if p.cur.Kind == token.Else {
		p.nextToken()
		if p.cur.Kind == token.If {
			elseStmt = p.parseIfStmt()
		} else {
			elseStmt = p.parseBlock()
		}
	}

	return &ast.IfStmt{
		IfPos: ifTok.Pos,
		Cond:  cond,
		Then:  thenBlock,
		Else:  elseStmt,
	}
}

func (p *Parser) parseWhileStmt() ast.Stmt {
	whileTok := p.cur
	p.nextToken()
	p.expect(token.LParen)
	cond := p.parseExpr()
	p.expect(token.RParen)
	body := p.parseBlock()

	return &ast.WhileStmt{
		WhilePos: whileTok.Pos,
		Cond:     cond,
		Body:     body,
	}
}

func (p *Parser) parseForStmt() ast.Stmt {
	forTok := p.cur
	p.nextToken()
	p.expect(token.LParen)

	// Check if this is a foreach loop: `for (ident in expr)`
	if p.cur.Kind == token.Ident && p.peek.Kind == token.In {
		varNameTok := p.cur
		p.nextToken() // consume ident
		p.nextToken() // consume 'in'
		listExpr := p.parseExpr()
		p.expect(token.RParen)
		body := p.parseBlock()

		return &ast.ForEachStmt{
			ForPos:   forTok.Pos,
			VarName:  varNameTok.Lexeme,
			VarPos:   varNameTok.Pos,
			ListExpr: listExpr,
			Body:     body,
		}
	}

	// C-style for loop: `for (init; cond; post)`
	var init ast.Stmt
	var cond ast.Expr
	var post ast.Stmt

	// Parse init (optional)
	if p.cur.Kind != token.Semicolon {
		if p.cur.Kind == token.Var {
			init = p.parseVarDeclStmt()
		} else if p.cur.Kind == token.Ident && p.peek.Kind == token.Assign {
			init = p.parseAssignStmt()
		} else {
			// expression statement
			expr := p.parseExpr()
			p.expect(token.Semicolon)
			init = &ast.ExprStmt{Expression: expr}
		}
	} else {
		p.nextToken() // consume semicolon
	}

	// Parse cond (optional)
	if p.cur.Kind != token.Semicolon {
		cond = p.parseExpr()
	}
	if p.cur.Kind == token.Semicolon {
		p.nextToken()
	}

	// Parse post (optional)
	if p.cur.Kind != token.RParen {
		// post can be an assignment or expression statement
		// Note: post doesn't end with semicolon, it ends with RParen
		if p.cur.Kind == token.Ident && p.peek.Kind == token.Assign {
			// Parse assignment without consuming semicolon
			nameTok := p.cur
			p.nextToken()
			p.expect(token.Assign)
			value := p.parseExpr()
			post = &ast.AssignStmt{
				Name:    nameTok.Lexeme,
				NamePos: nameTok.Pos,
				Value:   value,
			}
		} else {
			// expression statement
			expr := p.parseExpr()
			post = &ast.ExprStmt{Expression: expr}
		}
	}
	p.expect(token.RParen)

	body := p.parseBlock()

	return &ast.ForStmt{
		ForPos: forTok.Pos,
		Init:   init,
		Cond:   cond,
		Post:   post,
		Body:   body,
	}
}

func (p *Parser) parseThrowStmt() ast.Stmt {
	throwTok := p.cur
	p.nextToken()

	// expression is required
	expr := p.parseExpr()
	p.expect(token.Semicolon)

	return &ast.ThrowStmt{
		ThrowPos: throwTok.Pos,
		Expr:     expr,
	}
}

func (p *Parser) parseBreakStmt() ast.Stmt {
	breakTok := p.cur
	p.nextToken()

	// break is always followed by a semicolon
	p.expect(token.Semicolon)

	return &ast.BreakStmt{
		BreakPos: breakTok.Pos,
	}
}

func (p *Parser) parseTryStmt() ast.Stmt {
	tryTok := p.cur
	p.nextToken()

	// expect block for try body
	body := p.parseBlock()

	ts := &ast.TryStmt{
		TryPos: tryTok.Pos,
		Body:   body,
	}

	if p.cur.Kind == token.Catch {
		p.nextToken()

		// catch (e | error) { ... }
		p.expect(token.LParen)

		if p.cur.Kind != token.Ident {
			p.errorf(p.cur.Pos, "expected identifier after 'catch('")
		}
		nameTok := p.cur
		p.nextToken()

		p.expect(token.Pipe)
		catchType := p.parseType()

		p.expect(token.RParen)

		catchBody := p.parseBlock()

		ts.CatchName = nameTok.Lexeme
		ts.CatchPos = nameTok.Pos
		ts.CatchType = catchType
		ts.CatchBody = catchBody
	}

	return ts
}

// ---------- Expressions (with priorities) ----------

func (p *Parser) parseExpr() ast.Expr {
	return p.parseOr()
}

func (p *Parser) parseOr() ast.Expr {
	left := p.parseAnd()
	for p.cur.Kind == token.OrOr {
		opTok := p.cur
		p.nextToken()
		right := p.parseAnd()
		left = &ast.BinaryExpr{
			OpPos: opTok.Pos,
			Op:    opTok.Kind,
			Left:  left,
			Right: right,
		}
	}
	return left
}

func (p *Parser) parseAnd() ast.Expr {
	left := p.parseEquality()
	for p.cur.Kind == token.AndAnd {
		opTok := p.cur
		p.nextToken()
		right := p.parseEquality()
		left = &ast.BinaryExpr{
			OpPos: opTok.Pos,
			Op:    opTok.Kind,
			Left:  left,
			Right: right,
		}
	}
	return left
}

func (p *Parser) parseEquality() ast.Expr {
	left := p.parseRelational()
	for p.cur.Kind == token.Eq || p.cur.Kind == token.NotEq {
		opTok := p.cur
		p.nextToken()
		right := p.parseRelational()
		left = &ast.BinaryExpr{
			OpPos: opTok.Pos,
			Op:    opTok.Kind,
			Left:  left,
			Right: right,
		}
	}
	return left
}

func (p *Parser) parseRelational() ast.Expr {
	left := p.parseAdditive()
	for p.cur.Kind == token.Lt || p.cur.Kind == token.LtEq ||
		p.cur.Kind == token.Gt || p.cur.Kind == token.GtEq {
		opTok := p.cur
		p.nextToken()
		right := p.parseAdditive()
		left = &ast.BinaryExpr{
			OpPos: opTok.Pos,
			Op:    opTok.Kind,
			Left:  left,
			Right: right,
		}
	}
	return left
}

func (p *Parser) parseAdditive() ast.Expr {
	left := p.parseMultiplicative()
	for p.cur.Kind == token.Plus || p.cur.Kind == token.Minus {
		opTok := p.cur
		p.nextToken()
		right := p.parseMultiplicative()
		left = &ast.BinaryExpr{
			OpPos: opTok.Pos,
			Op:    opTok.Kind,
			Left:  left,
			Right: right,
		}
	}
	return left
}

func (p *Parser) parseMultiplicative() ast.Expr {
	left := p.parseUnary()
	for p.cur.Kind == token.Star || p.cur.Kind == token.Slash || p.cur.Kind == token.Percent {
		opTok := p.cur
		p.nextToken()
		right := p.parseUnary()
		left = &ast.BinaryExpr{
			OpPos: opTok.Pos,
			Op:    opTok.Kind,
			Left:  left,
			Right: right,
		}
	}
	return left
}

func (p *Parser) parseUnary() ast.Expr {
	if p.cur.Kind == token.Bang || p.cur.Kind == token.Minus {
		opTok := p.cur
		p.nextToken()
		x := p.parseUnary()
		return &ast.UnaryExpr{
			OpPos: opTok.Pos,
			Op:    opTok.Kind,
			X:     x,
		}
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() ast.Expr {
	expr := p.parsePrimary()

	for {
		switch p.cur.Kind {
		case token.Dot:
			// Member access: expr.name
			p.nextToken()
			if p.cur.Kind != token.Ident {
				p.errorf(p.cur.Pos, "expected identifier after '.'")
				return expr
			}
			nameTok := p.cur
			p.nextToken()
			expr = &ast.MemberExpr{
				X:       expr,
				Name:    nameTok.Lexeme,
				NamePos: nameTok.Pos,
			}
		case token.LParen:
			// function call
			lparen := p.cur
			p.nextToken()
			var args []ast.Expr
			if p.cur.Kind != token.RParen {
				for {
					var arg ast.Expr
					// Check if this is a named argument: ident = expr
					if p.cur.Kind == token.Ident && p.peek.Kind == token.Assign {
						nameTok := p.cur
						p.nextToken() // consume ident
						p.expect(token.Assign)
						valueExpr := p.parseExpr()
						arg = &ast.NamedArg{
							Name:    nameTok.Lexeme,
							NamePos: nameTok.Pos,
							Value:   valueExpr,
						}
					} else {
						// Positional argument
						arg = p.parseExpr()
					}
					args = append(args, arg)
					if p.cur.Kind == token.Comma {
						p.nextToken()
						continue
					}
					break
				}
			}
			rparen := p.expect(token.RParen)
			expr = &ast.CallExpr{
				Callee: expr,
				LParen: lparen.Pos,
				Args:   args,
				RParen: rparen.Pos,
			}
		case token.LBracket:
			// list indexing
			lbr := p.cur
			p.nextToken()
			indexExpr := p.parseExpr()
			rbr := p.expect(token.RBracket)
			expr = &ast.IndexExpr{
				X:        expr,
				LBracket: lbr.Pos,
				Index:    indexExpr,
				RBracket: rbr.Pos,
			}
		default:
			return expr
		}
	}
}

func (p *Parser) parsePrimary() ast.Expr {
	switch p.cur.Kind {
	case token.Fun:
		return p.parseFuncLiteral()
	case token.Ident, token.ErrorType:
		// Allow error as identifier in expression contexts (for builtin function)
		// Could be a struct literal: TypeName{field = value, ...}
		// Peek ahead to see if it's followed by {
		if p.peek.Kind == token.LBrace {
			return p.parseStructLiteral()
		}
		// Otherwise it's just an identifier
		tok := p.cur
		p.nextToken()
		return &ast.IdentExpr{
			Name:    tok.Lexeme,
			NamePos: tok.Pos,
		}
	case token.Int:
		tok := p.cur
		p.nextToken()
		val, err := strconv.ParseInt(tok.Lexeme, 10, 64)
		if err != nil {
			p.errorf(tok.Pos, "invalid integer literal %q: %v", tok.Lexeme, err)
			val = 0
		}
		return &ast.IntLiteral{
			Value:  val,
			LitPos: tok.Pos,
			Raw:    tok.Lexeme,
		}
	case token.Float:
		tok := p.cur
		p.nextToken()
		val, err := strconv.ParseFloat(tok.Lexeme, 64)
		if err != nil {
			p.errorf(tok.Pos, "invalid float literal %q: %v", tok.Lexeme, err)
			val = 0
		}
		return &ast.FloatLiteral{
			Value:  val,
			LitPos: tok.Pos,
			Raw:    tok.Lexeme,
		}
	case token.String:
		tok := p.cur
		p.nextToken()
		return &ast.StringLiteral{
			Value:  tok.Lexeme,
			LitPos: tok.Pos,
		}
	case token.StringPart, token.InterpStart:
		return p.parseInterpolatedString()
	case token.Bytes:
		tok := p.cur
		p.nextToken()
		return &ast.BytesLiteral{
			Value:  []byte(tok.Lexeme),
			LitPos: tok.Pos,
		}
	case token.True, token.False:
		tok := p.cur
		p.nextToken()
		return &ast.BoolLiteral{
			Value:  tok.Kind == token.True,
			LitPos: tok.Pos,
		}
	case token.None:
		tok := p.cur
		p.nextToken()
		return &ast.NoneLiteral{
			LitPos: tok.Pos,
		}
	case token.Some:
		tok := p.cur
		p.nextToken()
		p.expect(token.LParen)
		val := p.parseExpr()
		p.expect(token.RParen)
		return &ast.SomeLiteral{
			SomePos: tok.Pos,
			Value:   val,
		}
	case token.LParen:
		p.nextToken()
		expr := p.parseExpr()
		p.expect(token.RParen)
		return expr
	case token.LBracket:
		// list literal: [expr, expr, ...]
		lbr := p.cur
		p.nextToken()
		var elems []ast.Expr
		if p.cur.Kind != token.RBracket {
			for {
				elem := p.parseExpr()
				elems = append(elems, elem)
				if p.cur.Kind == token.Comma {
					p.nextToken()
					continue
				}
				break
			}
		}
		rbr := p.expect(token.RBracket)
		return &ast.ListLiteral{
			LBracket: lbr.Pos,
			Elements: elems,
			RBracket: rbr.Pos,
		}
	case token.LBrace:
		return p.parseDictLiteral()
	default:
		tok := p.cur
		p.errorf(tok.Pos, "unexpected token in expression: %s", tok.Kind)
		p.nextToken()
		return &ast.IntLiteral{
			Value:  0,
			LitPos: tok.Pos,
			Raw:    "0",
		}
	}
}

func (p *Parser) parseStructLiteral() ast.Expr {
	typeNameTok := p.cur
	typeName := typeNameTok.Lexeme
	p.nextToken()

	lbrace := p.expect(token.LBrace)

	var fields []*ast.FieldInit
	if p.cur.Kind != token.RBrace {
		for {
			if p.cur.Kind != token.Ident {
				p.errorf(p.cur.Pos, "expected field name")
				break
			}
			fieldNameTok := p.cur
			p.nextToken()

			p.expect(token.Assign)
			value := p.parseExpr()

			fields = append(fields, &ast.FieldInit{
				Name:    fieldNameTok.Lexeme,
				NamePos: fieldNameTok.Pos,
				Value:   value,
			})

			if p.cur.Kind == token.Comma {
				p.nextToken()
				continue
			}
			break
		}
	}

	rbrace := p.expect(token.RBrace)

	return &ast.StructLiteral{
		TypeName:    typeName,
		TypeNamePos: typeNameTok.Pos,
		LBrace:      lbrace.Pos,
		Fields:      fields,
		RBrace:      rbrace.Pos,
	}
}

func (p *Parser) parseDictLiteral() ast.Expr {
	lbrace := p.cur
	p.nextToken()

	var entries []*ast.DictEntry
	if p.cur.Kind != token.RBrace {
		for {
			var keyTok token.Token
			switch p.cur.Kind {
			case token.Ident, token.String:
				keyTok = p.cur
				p.nextToken()
			default:
				p.errorf(p.cur.Pos, "expected dict key")
				return &ast.DictLiteral{LBrace: lbrace.Pos, Entries: entries, RBrace: lbrace.Pos}
			}

			p.expect(token.Colon)
			value := p.parseExpr()

			entries = append(entries, &ast.DictEntry{
				Key:    keyTok.Lexeme,
				KeyPos: keyTok.Pos,
				Value:  value,
			})

			if p.cur.Kind == token.Comma {
				p.nextToken()
				if p.cur.Kind == token.RBrace {
					break
				}
				continue
			}
			break
		}
	}

	rbrace := p.expect(token.RBrace)
	return &ast.DictLiteral{
		LBrace:  lbrace.Pos,
		Entries: entries,
		RBrace:  rbrace.Pos,
	}
}

func (p *Parser) parseInterpolatedString() ast.Expr {
	startPos := p.cur.Pos
	var parts []ast.StringPart

	for {
		switch p.cur.Kind {
		case token.StringPart:
			parts = append(parts, &ast.StringTextPart{
				Value:   p.cur.Lexeme,
				PartPos: p.cur.Pos,
			})
			p.nextToken()
		case token.InterpStart:
			p.nextToken()
			expr := p.parseExpr()
			parts = append(parts, &ast.StringExprPart{Expr: expr})
			if p.cur.Kind != token.InterpEnd {
				p.errorf(p.cur.Pos, "expected '}' to close interpolation")
			} else {
				p.nextToken()
			}
		case token.StringEnd:
			p.nextToken()
			return &ast.InterpolatedString{
				Parts:  parts,
				LitPos: startPos,
			}
		default:
			p.errorf(p.cur.Pos, "unexpected token in interpolated string: %s", p.cur.Kind)
			p.nextToken()
			return &ast.InterpolatedString{
				Parts:  parts,
				LitPos: startPos,
			}
		}
	}
}
