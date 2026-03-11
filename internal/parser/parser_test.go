package parser_test

import (
	"testing"

	"avenir/internal/ast"
	"avenir/internal/lexer"
	"avenir/internal/parser"
)

func TestParseSimpleProgram(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var result | string = hello_or_bye(10);
    print(result);
}

fun hello_or_bye(a | int) | string {
    if (a > 10) {
        return "Hello";
    }
    if (a > 0; a < 10) {
        return "Bye";
    }
    return "...";
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if prog.Package == nil || prog.Package.Name != "main" {
		t.Fatalf("expected package 'main', got %#v", prog.Package)
	}

	if len(prog.Funcs) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(prog.Funcs))
	}

	if prog.Funcs[0].Name != "main" {
		t.Errorf("expected first function 'main', got %q", prog.Funcs[0].Name)
	}
	if prog.Funcs[1].Name != "hello_or_bye" {
		t.Errorf("expected second function 'hello_or_bye', got %q", prog.Funcs[1].Name)
	}
}

func TestParseWhileLoop(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var i | int = 0;
    while (i < 10) {
        i = i + 1;
    }
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}
}

func TestParseForLoop(t *testing.T) {
	input := `pckg main;

fun main() | void {
    for (var i | int = 0; i < 10; i = i + 1) {
        print(i);
    }
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}
}

func TestParseForEachLoop(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var xs | list<int> = [1, 2, 3];
    for (item in xs) {
        print(item);
    }
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}
}

func TestParseForLoopInfinite(t *testing.T) {
	input := `pckg main;

fun main() | void {
    for (;;) {
        print("infinite");
    }
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}
}

func TestParseDefaultParameter(t *testing.T) {
	input := `pckg main;

fun sum(a | int, b | int = 0) | int {
    return a + b;
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}

	fn := prog.Funcs[0]
	if len(fn.Params) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(fn.Params))
	}

	if fn.Params[0].Default != nil {
		t.Errorf("first parameter should not have default")
	}
	if fn.Params[1].Default == nil {
		t.Errorf("second parameter should have default")
	}
}

func TestParseNamedArguments(t *testing.T) {
	input := `pckg main;

fun sum(a | int, b | int = 0) | int {
    return a + b;
}

fun main() | void {
    sum(a=1, b=2);
    sum(1, b=2);
    sum(b=2, a=1);
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(prog.Funcs))
	}
}

func TestParseTryCatch(t *testing.T) {
	input := `pckg main;

fun main() | void {
    try {
        throw error("boom");
    } catch (e | error) {
        print(errorMessage(e));
    }
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}

	// Check that the try statement is in the body
	mainBody := prog.Funcs[0].Body
	if len(mainBody.Stmts) != 1 {
		t.Fatalf("expected 1 statement in main, got %d", len(mainBody.Stmts))
	}

	tryStmt, ok := mainBody.Stmts[0].(*ast.TryStmt)
	if !ok {
		t.Fatalf("expected TryStmt, got %T", mainBody.Stmts[0])
	}

	if tryStmt.CatchName != "e" {
		t.Errorf("expected catch name 'e', got %q", tryStmt.CatchName)
	}

	if tryStmt.CatchBody == nil {
		t.Errorf("expected catch body, got nil")
	}
}

func TestParseThrow(t *testing.T) {
	input := `pckg main;

fun main() | void {
    throw error("fail");
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}
}

func TestParseUnionTypeInVar(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var x | <int|string|bool> = 10;
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}
}

func TestParseUnionTypeInReturn(t *testing.T) {
	input := `pckg main;

fun f() | <string|bool> {
    return "ok";
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}
}

func TestParseUnionTypeInListElement(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var xs | list< <int|string> > = [1, "a"];
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}
}

func TestParseQualifiedTypes(t *testing.T) {
	input := `pckg main;

import std.net;

struct Box {
    sock | net.Socket
}

fun main(sock | net.Socket) | net.Server {
    var s | std.net.Socket = sock;
    return net.listen("0.0.0.0", 8080);
}
`
	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}
	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}
}

func TestParseInterpolatedString(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var x | int = 10;
    var s | string = "x=${x}, sum=${x + 2}";
}
`
	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}
}

func TestParseSingleQuotedStrings(t *testing.T) {
	input := `pckg main;

struct Config {
    name | string = 'Alice'
}

fun greet(msg | string) | void {
    print(msg);
}

fun main() | void {
    var a | string = 'hello';
    greet("hi");
    var b | string = 'a' + "b";
}
`
	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(prog.Structs))
	}
	field := prog.Structs[0].Fields[0]
	lit, ok := field.DefaultExpr.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expected string literal default, got %T", field.DefaultExpr)
	}
	if lit.Value != "Alice" {
		t.Fatalf("expected default %q, got %q", "Alice", lit.Value)
	}
}

func TestParseDictLiteral(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var user | dict<any> = {
        name: "Alex",
        "age": 30,
        meta: { active: true }
    };
}
`
	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}
	varDecl := prog.Funcs[0].Body.Stmts[0].(*ast.VarDeclStmt)
	dictLit, ok := varDecl.Value.(*ast.DictLiteral)
	if !ok {
		t.Fatalf("expected DictLiteral, got %T", varDecl.Value)
	}
	if len(dictLit.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(dictLit.Entries))
	}
	if dictLit.Entries[0].Key != "name" {
		t.Fatalf("expected key %q, got %q", "name", dictLit.Entries[0].Key)
	}
	if dictLit.Entries[1].Key != "age" {
		t.Fatalf("expected key %q, got %q", "age", dictLit.Entries[1].Key)
	}
	if dictLit.Entries[2].Key != "meta" {
		t.Fatalf("expected key %q, got %q", "meta", dictLit.Entries[2].Key)
	}
}

func TestParseImport(t *testing.T) {
	input := `pckg main;

import std.io;
import std.collections as coll;

fun main() | void {
    io.println("hi");
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(prog.Imports))
	}

	imp1 := prog.Imports[0]
	if len(imp1.Path) != 2 || imp1.Path[0] != "std" || imp1.Path[1] != "io" {
		t.Errorf("expected import path ['std', 'io'], got %v", imp1.Path)
	}
	if imp1.Alias != "" {
		t.Errorf("expected empty alias, got %q", imp1.Alias)
	}

	imp2 := prog.Imports[1]
	if len(imp2.Path) != 2 || imp2.Path[0] != "std" || imp2.Path[1] != "collections" {
		t.Errorf("expected import path ['std', 'collections'], got %v", imp2.Path)
	}
	if imp2.Alias != "coll" {
		t.Errorf("expected alias 'coll', got %q", imp2.Alias)
	}
}

func TestParsePubFun(t *testing.T) {
	input := `pckg main;

pub fun publicFunc() | void {
    print("public");
}

fun privateFunc() | void {
    print("private");
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(prog.Funcs))
	}

	if !prog.Funcs[0].IsPublic {
		t.Errorf("expected first function to be public")
	}
	if prog.Funcs[1].IsPublic {
		t.Errorf("expected second function to be private")
	}
}

func TestParseMemberExpr(t *testing.T) {
	input := `pckg main;

import std.io;

fun main() | void {
    io.println("test");
    var x | int = 10;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	// Check that member expression is parsed correctly
	mainFn := prog.Funcs[0]
	if mainFn.Name != "main" {
		t.Fatalf("expected function 'main', got %q", mainFn.Name)
	}
	// The member expression should be in the function body
	// This is a basic test - more detailed AST inspection could be added
}

func TestParseStructDecl(t *testing.T) {
	input := `pckg main;

struct Point {
    x | int
    y | int
}

fun main() | void {
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(prog.Structs))
	}

	st := prog.Structs[0]
	if st.Name != "Point" {
		t.Fatalf("expected struct name 'Point', got %q", st.Name)
	}

	if len(st.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(st.Fields))
	}

	if st.Fields[0].Name != "x" {
		t.Fatalf("expected first field 'x', got %q", st.Fields[0].Name)
	}
	if st.Fields[1].Name != "y" {
		t.Fatalf("expected second field 'y', got %q", st.Fields[1].Name)
	}
}

func TestParseStructLiteral(t *testing.T) {
	input := `pckg main;

struct Point {
    x | int
    y | int
}

fun main() | void {
    var p | Point = Point{x = 10, y = 20};
}
`

	l := lexer.New(input)
	p := parser.New(l)

	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	// Check that struct literal was parsed
	mainFunc := prog.Funcs[0]
	if mainFunc.Name != "main" {
		t.Fatalf("expected function 'main'")
	}

	// The struct literal should be in the variable declaration
	varDecl := mainFunc.Body.Stmts[0].(*ast.VarDeclStmt)
	structLit, ok := varDecl.Value.(*ast.StructLiteral)
	if !ok {
		t.Fatalf("expected StructLiteral, got %T", varDecl.Value)
	}

	if structLit.TypeName != "Point" {
		t.Fatalf("expected struct type 'Point', got %q", structLit.TypeName)
	}

	if len(structLit.Fields) != 2 {
		t.Fatalf("expected 2 field initializations, got %d", len(structLit.Fields))
	}

	if structLit.Fields[0].Name != "x" {
		t.Fatalf("expected first field 'x', got %q", structLit.Fields[0].Name)
	}
	if structLit.Fields[1].Name != "y" {
		t.Fatalf("expected second field 'y', got %q", structLit.Fields[1].Name)
	}
}

func TestParseAsyncFunction(t *testing.T) {
	input := `pckg main;

async fun fetch() | int {
    return 10;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}

	fn := prog.Funcs[0]
	if fn.Name != "fetch" {
		t.Fatalf("expected function name 'fetch', got %q", fn.Name)
	}
	if !fn.IsAsync {
		t.Fatalf("expected function to be async")
	}
}

func TestParseAsyncFunctionNotAsync(t *testing.T) {
	input := `pckg main;

fun fetch() | int {
    return 10;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if !prog.Funcs[0].IsAsync == false {
		t.Fatalf("expected function to NOT be async")
	}
}

func TestParsePubAsyncFunction(t *testing.T) {
	input := `pckg main;

pub async fun fetch() | int {
    return 10;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}

	fn := prog.Funcs[0]
	if !fn.IsPublic {
		t.Fatalf("expected function to be public")
	}
	if !fn.IsAsync {
		t.Fatalf("expected function to be async")
	}
}

func TestParseAwaitExpression(t *testing.T) {
	input := `pckg main;

async fun fetch() | int {
    return 10;
}

async fun main() | int {
    var x | int = await fetch();
    return x;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(prog.Funcs))
	}

	mainFn := prog.Funcs[1]
	if mainFn.Name != "main" {
		t.Fatalf("expected function name 'main', got %q", mainFn.Name)
	}

	varDecl, ok := mainFn.Body.Stmts[0].(*ast.VarDeclStmt)
	if !ok {
		t.Fatalf("expected VarDeclStmt, got %T", mainFn.Body.Stmts[0])
	}

	awaitExpr, ok := varDecl.Value.(*ast.AwaitExpr)
	if !ok {
		t.Fatalf("expected AwaitExpr, got %T", varDecl.Value)
	}

	callExpr, ok := awaitExpr.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr inside await, got %T", awaitExpr.Expr)
	}

	ident, ok := callExpr.Callee.(*ast.IdentExpr)
	if !ok {
		t.Fatalf("expected IdentExpr as callee, got %T", callExpr.Callee)
	}
	if ident.Name != "fetch" {
		t.Fatalf("expected callee 'fetch', got %q", ident.Name)
	}
}

func TestParseOptionalChaining(t *testing.T) {
	input := `pckg main;

fun main() | void {
	var user | any = none;
	user?.name;
	user?.name();
}
`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	body := prog.Funcs[0].Body
	firstExprStmt, ok := body.Stmts[1].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected first optional chain expression statement, got %T", body.Stmts[1])
	}
	if _, ok := firstExprStmt.Expression.(*ast.OptionalMemberExpr); !ok {
		t.Fatalf("expected OptionalMemberExpr, got %T", firstExprStmt.Expression)
	}

	secondExprStmt, ok := body.Stmts[2].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected second optional chain expression statement, got %T", body.Stmts[2])
	}
	if _, ok := secondExprStmt.Expression.(*ast.OptionalCallExpr); !ok {
		t.Fatalf("expected OptionalCallExpr, got %T", secondExprStmt.Expression)
	}
}

func TestParseSwitchContinueAndDefer(t *testing.T) {
	input := `pckg main;

fun main() | void {
	for (var i | int = 0; i < 3; i = i + 1) {
		switch i {
			case 0:
				continue;
			default:
				defer print(i);
		}
	}
}
`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	forStmt, ok := prog.Funcs[0].Body.Stmts[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", prog.Funcs[0].Body.Stmts[0])
	}

	switchStmt, ok := forStmt.Body.Stmts[0].(*ast.SwitchStmt)
	if !ok {
		t.Fatalf("expected SwitchStmt, got %T", forStmt.Body.Stmts[0])
	}

	if len(switchStmt.Cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(switchStmt.Cases))
	}
	if len(switchStmt.Default) != 1 {
		t.Fatalf("expected default clause with 1 statement, got %d", len(switchStmt.Default))
	}

	if _, ok := switchStmt.Cases[0].Body[0].(*ast.ContinueStmt); !ok {
		t.Fatalf("expected ContinueStmt in case body, got %T", switchStmt.Cases[0].Body[0])
	}

	if _, ok := switchStmt.Default[0].(*ast.DeferStmt); !ok {
		t.Fatalf("expected DeferStmt in default body, got %T", switchStmt.Default[0])
	}
}

func TestParseDecorator(t *testing.T) {
	input := `pckg main;

fun log(f | fun(int, int) | int) | fun(int, int) | int {
	return f;
}

@log
fun add(a | int, b | int) | int {
	return a + b;
}

fun main() | int {
	return add(1, 2);
}
`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	if len(prog.Funcs) != 3 {
		t.Fatalf("expected 3 functions, got %d", len(prog.Funcs))
	}

	addFn := prog.Funcs[1]
	if addFn.Name != "add" {
		t.Fatalf("expected function 'add', got %q", addFn.Name)
	}
	if len(addFn.Decorators) != 1 {
		t.Fatalf("expected 1 decorator, got %d", len(addFn.Decorators))
	}
	ident, ok := addFn.Decorators[0].Expr.(*ast.IdentExpr)
	if !ok {
		t.Fatalf("expected decorator expr to be *ast.IdentExpr, got %T", addFn.Decorators[0].Expr)
	}
	if ident.Name != "log" {
		t.Fatalf("expected decorator 'log', got %q", ident.Name)
	}
}

func TestParseDecoratorWithArgs(t *testing.T) {
	input := `pckg main;

@cache(60)
fun compute(x | int) | int {
	return x * x;
}

fun cache(ttl | int) | fun(fun(int) | int) | fun(int) | int {
	return fun(f | fun(int) | int) | fun(int) | int { return f; };
}

fun main() | int {
	return compute(5);
}
`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	computeFn := prog.Funcs[0]
	if computeFn.Name != "compute" {
		t.Fatalf("expected function 'compute', got %q", computeFn.Name)
	}
	if len(computeFn.Decorators) != 1 {
		t.Fatalf("expected 1 decorator, got %d", len(computeFn.Decorators))
	}
	dec := computeFn.Decorators[0]
	callExpr, ok := dec.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected decorator expr to be *ast.CallExpr, got %T", dec.Expr)
	}
	callee, ok := callExpr.Callee.(*ast.IdentExpr)
	if !ok {
		t.Fatalf("expected callee to be *ast.IdentExpr, got %T", callExpr.Callee)
	}
	if callee.Name != "cache" {
		t.Fatalf("expected decorator callee 'cache', got %q", callee.Name)
	}
	if len(callExpr.Args) != 1 {
		t.Fatalf("expected 1 decorator arg, got %d", len(callExpr.Args))
	}
}

func TestParseVariadicTypeParam(t *testing.T) {
	input := `pckg main;

fun wrap<R, ...Args>(f | fun(Args...) | R) | fun(Args...) | R {
	return f;
}

fun main() | void {
}
`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	wrapFn := prog.Funcs[0]
	if wrapFn.Name != "wrap" {
		t.Fatalf("expected function 'wrap', got %q", wrapFn.Name)
	}
	if len(wrapFn.TypeParams) != 2 {
		t.Fatalf("expected 2 type params, got %d", len(wrapFn.TypeParams))
	}
	if wrapFn.TypeParams[0].Name != "R" || wrapFn.TypeParams[0].IsVariadic {
		t.Fatalf("expected first type param 'R' non-variadic, got %q variadic=%v", wrapFn.TypeParams[0].Name, wrapFn.TypeParams[0].IsVariadic)
	}
	if wrapFn.TypeParams[1].Name != "Args" || !wrapFn.TypeParams[1].IsVariadic {
		t.Fatalf("expected second type param 'Args' variadic, got %q variadic=%v", wrapFn.TypeParams[1].Name, wrapFn.TypeParams[1].IsVariadic)
	}

	paramType := wrapFn.Params[0].Type
	funcType, ok := paramType.(*ast.FuncType)
	if !ok {
		t.Fatalf("expected FuncType for param, got %T", paramType)
	}
	if len(funcType.ParamTypes) != 1 {
		t.Fatalf("expected 1 param type in FuncType, got %d", len(funcType.ParamTypes))
	}
	expansion, ok := funcType.ParamTypes[0].(*ast.TypePackExpansion)
	if !ok {
		t.Fatalf("expected TypePackExpansion, got %T", funcType.ParamTypes[0])
	}
	if expansion.Name != "Args" {
		t.Fatalf("expected expansion name 'Args', got %q", expansion.Name)
	}
}
