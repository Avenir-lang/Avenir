package types_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"avenir/internal/lexer"
	"avenir/internal/modules"
	"avenir/internal/parser"
	_ "avenir/internal/runtime" // Ensure builtin packages are loaded for registration
	"avenir/internal/types"
)

func TestCheckProgram_Valid(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var a | int = 10;
    var b | string = "Test";
    var c | bool = true;
    var d | list<any> = [10, "jdkdbd", false];
    var e | list<string, int> = ["foo", 42];

    var result | string = hello_or_bye(a);
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckProgram_InvalidAssignment(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var a | int = 10;
    a = "oops";
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error, got none")
	}
}

func TestCheckProgram_WhileLoop(t *testing.T) {
	input := `
pckg main;

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
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckProgram_WhileLoopInvalidCondition(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var i | int = 0;
    while (i) {
        i = i + 1;
    }
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for non-bool condition, got none")
	}
}

func TestCheckProgram_StringConcatValid(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var a | string = "hello ";
    var b | string = "world";
    var c | string = a + b;
    var d | string = "a" + "b";
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckProgram_StringConcatInvalid(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var a | string = "a" + 1;
    var b | string = 1 + "a";
    var c | string = "a" + true;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type errors for invalid string concatenation, got none")
	}
	expected := "operator '+' is not defined for types"
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), expected) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error %q, got: %v", expected, errs)
	}
}

func TestCheckProgram_DictValid(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var user | dict<string> = {
        name: "Alex",
        role: "admin"
    };
    var name | string = user.name;
    var hasRole | bool = user.has("role");
    var role | string? = user.get("role");
    var keys | list<string> = user.keys();
    var values | list<string> = user.values();
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckProgram_DictInvalidValueTypes(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var d | dict<int> = {
        a: 1,
        b: "oops"
    };
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error, got none")
	}
}

func TestCheckProgram_DictMissingKey(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var x | int = { a: 1 }.missing;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for missing dict key, got none")
	}
}

func TestCheckProgram_ForLoop(t *testing.T) {
	input := `
pckg main;

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
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckProgram_ForEachLoop(t *testing.T) {
	input := `
pckg main;

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
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckProgram_ForEachLoopInvalidType(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var x | int = 10;
    for (item in x) {
        print(item);
    }
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for non-list in foreach, got none")
	}
}

func TestCheckProgram_UnionTypeValid(t *testing.T) {
	input := `
pckg main;

fun f() | <string|bool> {
    if (true) {
        return "ok";
    }
    return false;
}

fun g() | void {
    var x | <int|string> = 10;
    var y | <int|string> = "hello";
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckProgram_UnionTypeInvalidReturn(t *testing.T) {
	input := `
pckg main;

fun f() | <string|bool> {
    return 1;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for int not in union, got none")
	}
}

func TestCheckProgram_UnionTypeInvalidAssignment(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var x | <int|string> = true;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for bool not in union, got none")
	}
}

func TestCheckProgram_UnionTypeWithAny(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var x | any = 10;
    var y | <int|string> = x;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors (any is assignable), got %d", len(errs))
	}
}

func TestCheckProgram_UnionTypeInList(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var xs | list< <int|string> > = [1, "a", 2];
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckProgram_UnionTypeNoOperations(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var x | <int|string> = 10;
    var y | <int|string> = "hello";
    var z | int = x + y;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for operation on union types, got none")
	}
}

func TestCheckProgram_DefaultParameters(t *testing.T) {
	input := `
pckg main;

fun sum(a | int, b | int = 0) | int {
    return a + b;
}

fun main() | void {
    sum(1);
    sum(1, 2);
    sum(a=1);
    sum(b=2, a=1);
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckProgram_DefaultParametersMissingRequired(t *testing.T) {
	input := `
pckg main;

fun sum(a | int, b | int = 0) | int {
    return a + b;
}

fun main() | void {
    sum();
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for missing required parameter, got none")
	}
}

func TestCheckProgram_DefaultParametersInvalidOrder(t *testing.T) {
	input := `
pckg main;

fun f(a | int = 0, b | int) | int {
    return a + b;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for invalid parameter order, got none")
	}
}

func TestCheckProgram_NamedArgumentsPositionalAfterNamed(t *testing.T) {
	input := `
pckg main;

fun sum(a | int, b | int = 0) | int {
    return a + b;
}

fun main() | void {
    sum(a=1, 2);
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for positional after named, got none")
	}
}

func TestCheckProgram_NamedArgumentsUnknownParameter(t *testing.T) {
	input := `
pckg main;

fun sum(a | int, b | int = 0) | int {
    return a + b;
}

fun main() | void {
    sum(x=1);
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for unknown named parameter, got none")
	}
}

func TestCheckProgram_TryCatchValid(t *testing.T) {
	input := `
pckg main;

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
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckProgram_ThrowInvalidType(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    throw "not error";
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for throw with non-error type, got none")
	}
}

func TestCheckProgram_CatchInvalidType(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    try {
        throw error("boom");
    } catch (e | string) {
        print(e);
    }
}
`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for catch with non-error type, got none")
	}
}

func TestCheckQualifiedType_UnknownModule(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var s | net.Socket = net.connect("example.com", 80);
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for missing module import, got none")
	}
}

func TestCheckQualifiedType_PrivateType(t *testing.T) {
	tmpDir := t.TempDir()

	stdDir := filepath.Join(tmpDir, "std")
	if err := os.MkdirAll(stdDir, 0755); err != nil {
		t.Fatalf("failed to create std dir: %v", err)
	}
	netFile := filepath.Join(stdDir, "net.av")
	netContent := `pckg std.net;

struct net {}

struct Hidden {
    x | int;
}

pub struct Socket {
    handle | any;
}
`
	if err := os.WriteFile(netFile, []byte(netContent), 0644); err != nil {
		t.Fatalf("failed to write net.av: %v", err)
	}

	mainFile := filepath.Join(tmpDir, "main.av")
	mainContent := `pckg main;

import std.net;

fun use_hidden(h | net.Hidden) | void {
    return;
}
`
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main.av: %v", err)
	}

	world, errs := modules.LoadWorld(mainFile)
	if len(errs) > 0 {
		t.Fatalf("failed to load world: %v", errs)
	}

	typeWorld := &types.World{
		Modules: make(map[string]*types.ModuleInfo),
		Entry:   world.Entry,
	}
	for modName, modAST := range world.Modules {
		typeWorld.Modules[modName] = &types.ModuleInfo{
			Name:  modName,
			Prog:  modAST.Prog,
			Scope: nil,
		}
	}

	_, typeErrs := types.CheckWorldWithBindings(typeWorld)
	if len(typeErrs) == 0 {
		t.Fatalf("expected type error for private type, got none")
	}
}

func TestCheckQualifiedType_UnknownType(t *testing.T) {
	tmpDir := t.TempDir()

	stdDir := filepath.Join(tmpDir, "std")
	if err := os.MkdirAll(stdDir, 0755); err != nil {
		t.Fatalf("failed to create std dir: %v", err)
	}
	netFile := filepath.Join(stdDir, "net.av")
	netContent := `pckg std.net;

struct net {}

pub struct Socket {
    handle | any;
}
`
	if err := os.WriteFile(netFile, []byte(netContent), 0644); err != nil {
		t.Fatalf("failed to write net.av: %v", err)
	}

	mainFile := filepath.Join(tmpDir, "main.av")
	mainContent := `pckg main;

import std.net;

fun use_missing(h | net.Missing) | void {
    return;
}
`
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main.av: %v", err)
	}

	world, errs := modules.LoadWorld(mainFile)
	if len(errs) > 0 {
		t.Fatalf("failed to load world: %v", errs)
	}

	typeWorld := &types.World{
		Modules: make(map[string]*types.ModuleInfo),
		Entry:   world.Entry,
	}
	for modName, modAST := range world.Modules {
		typeWorld.Modules[modName] = &types.ModuleInfo{
			Name:  modName,
			Prog:  modAST.Prog,
			Scope: nil,
		}
	}

	_, typeErrs := types.CheckWorldWithBindings(typeWorld)
	if len(typeErrs) == 0 {
		t.Fatalf("expected type error for unknown type, got none")
	}
}

func TestCheckProgram_FuncLiteralValid(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var f | fun(int) | int = fun(x | int) | int {
        return x + 1;
    };
    var result | int = f(5);
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckProgram_FuncLiteralCapture(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var x | int = 10;
    var f | fun(int) | int = fun(y | int) | int {
        return x + y;
    };
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckProgram_FuncLiteralTypeMismatch(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var f | fun(int) | string = fun(x | int) | int {
        return x;
    };
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for function literal type mismatch, got none")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "cannot assign expression") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected type mismatch error, got: %v", errs)
	}
}

func TestCheckInterpolatedString_InvalidExpr(t *testing.T) {
	input := `
pckg main;

fun main() | void {
    var s | string = "value=${unknown}";
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("unexpected parser errors: %v", errs)
	}

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for invalid interpolation expression, got none")
	}
}

func TestCheckStructDecl(t *testing.T) {
	input := `
pckg main;

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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckStructFieldAccess(t *testing.T) {
	input := `
pckg main;

struct Point {
    x | int
    y | int
}

fun main() | int {
    var p | Point = Point{x = 5, y = 15};
    return p.x;
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckStructMissingField(t *testing.T) {
	input := `
pckg main;

struct Point {
    x | int
    y | int
}

fun main() | void {
    var p | Point = Point{x = 10};
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

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for missing field 'y', got none")
	}

	hasMissingFieldError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "missing required field") {
			hasMissingFieldError = true
			break
		}
	}
	if !hasMissingFieldError {
		t.Fatalf("expected error about missing field, got: %v", errs)
	}
}

func TestCheckStructUnknownField(t *testing.T) {
	input := `
pckg main;

struct Point {
    x | int
    y | int
}

fun main() | void {
    var p | Point = Point{x = 10, y = 20, z = 30};
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

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for unknown field 'z', got none")
	}

	hasUnknownFieldError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "unknown field") {
			hasUnknownFieldError = true
			break
		}
	}
	if !hasUnknownFieldError {
		t.Fatalf("expected error about unknown field, got: %v", errs)
	}
}

func TestCheckStructFieldAccessError(t *testing.T) {
	input := `
pckg main;

struct Point {
    x | int
    y | int
}

fun main() | void {
    var p | Point = Point{x = 10, y = 20};
    var z | int = p.z;
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

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for accessing unknown field 'z', got none")
	}

	hasFieldError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "has no field") {
			hasFieldError = true
			break
		}
	}
	if !hasFieldError {
		t.Fatalf("expected error about unknown field access, got: %v", errs)
	}
}

func TestCheckVisibility_PublicFieldInPrivateStruct(t *testing.T) {
	input := `
pckg main;

struct Point {
    pub x | int
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

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for public field in private struct, got none")
	}

	hasError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "public field") && strings.Contains(e.Error(), "private struct") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatalf("expected error about public field in private struct, got: %v", errs)
	}
}

func TestCheckStructDefaults_AllDefaults(t *testing.T) {
	input := `
pckg main;

struct Config {
    host | string = "localhost"
    port | int = 8080
    secure | bool = false
}

fun main() | void {
    var c | Config = Config{};
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckStructDefaults_PartialInitialization(t *testing.T) {
	input := `
pckg main;

struct Config {
    host | string = "localhost"
    port | int = 8080
    secure | bool = false
}

fun main() | void {
    var c | Config = Config{ port = 9000 };
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckStructDefaults_MissingRequiredField(t *testing.T) {
	input := `
pckg main;

struct Config {
    host | string = "localhost"
    port | int
    secure | bool = false
}

fun main() | void {
    var c | Config = Config{};
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

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for missing required field 'port', got none")
	}

	hasError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "missing required field") && strings.Contains(e.Error(), "port") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatalf("expected error about missing required field 'port', got: %v", errs)
	}
}

func TestCheckStructDefaults_InvalidDefaultType(t *testing.T) {
	input := `
pckg main;

struct Config {
    port | int = "invalid"
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

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for invalid default type, got none")
	}

	hasError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "default value") && strings.Contains(e.Error(), "type") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatalf("expected error about invalid default type, got: %v", errs)
	}
}

func TestCheckStructDefaults_NonConstantDefault(t *testing.T) {
	input := `
pckg main;

struct Config {
    port | int = x
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

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for non-constant default, got none")
	}

	hasError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "compile-time constant") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatalf("expected error about compile-time constant, got: %v", errs)
	}
}

func TestCheckStructDefaults_VisibilityAndDefaults(t *testing.T) {
	input := `
pckg main;

pub struct Config {
    pub host | string = "localhost"
    port | int = 8080
}

fun main() | void {
    var c | Config = Config{ host = "example.com" };
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckVisibility_PrivateFieldAccessInMethod(t *testing.T) {
	input := `
pckg main;

struct Point {
    x | int
    y | int
}

fun (self | Point).sum() | int {
    return self.x + self.y;
}

fun main() | int {
    var p | Point = Point{x = 1, y = 2};
    return p.sum();
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckVisibility_PublicStruct(t *testing.T) {
	input := `
pckg main;

pub struct Point {
    x | int
    y | int
}

fun main() | void {
    var p | Point = Point{x = 1, y = 2};
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckVisibility_PublicFieldInPublicStruct(t *testing.T) {
	input := `
pckg main;

pub struct Point {
    pub x | int
    y | int
}

fun main() | void {
    var p | Point = Point{x = 1, y = 2};
    var n | int = p.x;
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckStructFieldAssign_PublicField(t *testing.T) {
	input := `
pckg main;

pub mut struct Point {
    pub x | int
    pub y | int
}

fun main() | void {
    var p | Point = Point{x = 0, y = 0};
    p.x = 10;
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckStructFieldAssign_PrivateField(t *testing.T) {
	input := `
pckg main;

mut struct Point {
    x | int
    y | int
}

fun main() | void {
    var p | Point = Point{x = 0, y = 0};
    p.x = 10;
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors (private fields accessible in same module), got %d", len(errs))
	}
}

func TestCheckStructFieldAssign_WrongType(t *testing.T) {
	input := `
pckg main;

mut struct Point {
    x | int
    y | int
}

fun main() | void {
    var p | Point = Point{x = 0, y = 0};
    p.x = "invalid";
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

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for wrong type assignment, got none")
	}

	hasError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "cannot assign") && strings.Contains(e.Error(), "type") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatalf("expected error about type mismatch, got: %v", errs)
	}
}

func TestCheckStructFieldAssign_NonExistentField(t *testing.T) {
	input := `
pckg main;

struct Point {
    x | int
    y | int
}

fun main() | void {
    var p | Point = Point{x = 0, y = 0};
    p.z = 10;
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

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for non-existent field, got none")
	}

	hasError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "has no field") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatalf("expected error about non-existent field, got: %v", errs)
	}
}

func TestCheckStructFieldAssign_WithDefaults(t *testing.T) {
	input := `
pckg main;

mut struct Config {
    host | string = "localhost"
    port | int = 8080
}

fun main() | void {
    var c | Config = Config{};
    c.port = 9000;
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckStructMutability_ImmutableStructAssignment(t *testing.T) {
	input := `
pckg main;

struct Point {
    x | int
    y | int
}

fun main() | void {
    var p | Point = Point{x = 0, y = 0};
    p.x = 10;
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

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for immutable field assignment, got none")
	}

	hasError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "cannot assign to immutable field") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatalf("expected error about immutable field, got: %v", errs)
	}
}

func TestCheckStructMutability_FieldLevelOverride_Mutable(t *testing.T) {
	input := `
pckg main;

struct Point {
    mut x | int
    y | int
}

fun main() | void {
    var p | Point = Point{x = 0, y = 0};
    p.x = 10;
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors (mutable field in immutable struct), got %d", len(errs))
	}
}

func TestCheckStructMutability_MutableStruct(t *testing.T) {
	input := `
pckg main;

mut struct Point {
    x | int
    y | int
}

fun main() | void {
    var p | Point = Point{x = 0, y = 0};
    p.x = 10;
    p.y = 20;
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

// ----- Interface Tests -----

func TestCheckInterface_ValidImplementation(t *testing.T) {
	input := `
pckg main;

interface Stringer {
    fun toString() | string
}

struct Point {
    x | int
    y | int
}

fun (p | Point).toString() | string {
    return "Point";
}

fun main() | void {
    var s | Stringer = Point{x = 1, y = 2};
    var str | string = s.toString();
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}

func TestCheckInterface_MissingMethod(t *testing.T) {
	input := `
pckg main;

interface Stringer {
    fun toString() | string
}

struct Point {
    x | int
    y | int
}

fun main() | void {
    var s | Stringer = Point{x = 1, y = 2};
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

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for missing method, got none")
	}

	hasError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "does not satisfy interface") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatalf("expected error about interface satisfaction, got: %v", errs)
	}
}

func TestCheckInterface_WrongSignature(t *testing.T) {
	input := `
pckg main;

interface Stringer {
    fun toString() | string
}

struct Point {
    x | int
    y | int
}

fun (p | Point).toString() | int {
    return 42;
}

fun main() | void {
    var s | Stringer = Point{x = 1, y = 2};
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

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error for wrong signature, got none")
	}

	hasError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "does not satisfy interface") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatalf("expected error about interface satisfaction, got: %v", errs)
	}
}

func TestCheckInterface_BuiltinTypeSatisfaction(t *testing.T) {
	input := `
pckg main;

interface Length {
    fun length() | int
}

fun main() | void {
    var l | Length = "hello";
    var len | int = l.length();
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors (string satisfies Length), got %d", len(errs))
	}
}

func TestCheckInterface_StaticMethodDoesNotSatisfy(t *testing.T) {
	input := `
pckg main;

interface Stringer {
    fun toString() | string
}

struct Point {
    x | int
    y | int
}

fun Point.toString() | string {
    return "Point";
}

fun main() | void {
    var s | Stringer = Point{x = 1, y = 2};
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

	errs := types.CheckProgram(prog)
	if len(errs) == 0 {
		t.Fatalf("expected type error (static method does not satisfy interface), got none")
	}

	hasError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "does not satisfy interface") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatalf("expected error about interface satisfaction, got: %v", errs)
	}
}

func TestCheckInterface_MultipleMethods(t *testing.T) {
	input := `
pckg main;

interface Writer {
    fun write(data | string) | void;
    fun flush() | void;
}

struct Buffer {
    data | string
}

fun (b | Buffer).write(data | string) | void {
    // ...
}

fun (b | Buffer).flush() | void {
    // ...
}

fun main() | void {
    var w | Writer = Buffer{data = ""};
    w.write("test");
    w.flush();
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

	errs := types.CheckProgram(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("expected no type errors, got %d", len(errs))
	}
}
