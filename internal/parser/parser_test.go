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
