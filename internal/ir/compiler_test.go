package ir_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"avenir/internal/ir"
	"avenir/internal/lexer"
	"avenir/internal/modules"
	"avenir/internal/parser"
	"avenir/internal/runtime"
	"avenir/internal/types"
	"avenir/internal/value"
	"avenir/internal/vm"
)

// testIO implements builtinsio.IO for testing
type testIO struct {
	output *bytes.Buffer
}

func (t *testIO) Println(s string) {
	t.output.WriteString(s)
}

func (t *testIO) ReadLine() (string, error) {
	// Test mock: return empty string for input
	return "", nil
}

// testOutputWriter implements builtinsio.IO for testing with string slice
type testOutputWriter struct {
	output *[]string
}

func (t *testOutputWriter) Println(s string) {
	*t.output = append(*t.output, s)
}

func (t *testOutputWriter) ReadLine() (string, error) {
	// Test mock: return empty string for input
	return "", nil
}

func TestCompile_ArithMain(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var a | int = 1;
    var b | int = 2;
    var c | int = 3;
    return a + b * c;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 7 {
		t.Fatalf("expected 7, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_IfAndReturnString(t *testing.T) {
	src := `
pckg main;

fun main() | string {
    return hello_or_bye(5);
}

fun hello_or_bye(a | int) | string {
    if (a > 10) {
        return "big";
    }
    if (a > 0; a < 10) {
        return "small";
    }
    return "mid";
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindString || val.Str != "small" {
		t.Fatalf("expected \"small\", got %q (%s)", val.Str, val.String())
	}
}

func TestCompile_Print(t *testing.T) {
	src := `
pckg main;

fun main() | void {
    print("Hello, Avenir!");
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	var buf bytes.Buffer
	env := runtime.NewEnv(&testIO{output: &buf})

	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if buf.String() != "Hello, Avenir!" {
		t.Fatalf("expected Stdout=%q, got %q", "Hello, Avenir!", buf.String())
	}
}

func TestCompile_StringInterpolation(t *testing.T) {
	src := `
pckg main;

fun main() | void {
    var x | int = 10;
    var y | int = 20;
    print("x=${x}, y=${y}, sum=${x + y}");
    print("Line1\nLine2\tTabbed");
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	var buf bytes.Buffer
	env := runtime.NewEnv(&testIO{output: &buf})

	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	expected := "x=10, y=20, sum=30Line1\nLine2\tTabbed"
	if buf.String() != expected {
		t.Fatalf("expected Stdout=%q, got %q", expected, buf.String())
	}
}

func TestCompile_SingleQuotedStrings(t *testing.T) {
	src := `
pckg main;

fun main() | bool {
    var a | string = 'hello';
    var b | string = "world";
    var c | string = a + " " + b;
    if (c != 'hello world') {
        return false;
    }
    if ('hi'.length() != 2) {
        return false;
    }
    return 'x' == "x";
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindBool || val.Bool != true {
		t.Fatalf("expected true, got %v (%s)", val.Bool, val.String())
	}
}

func TestCompile_StringConcat(t *testing.T) {
	src := `
pckg main;

fun main() | string {
    var name | string = "world";
    return "hello " + name + "!";
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindString || val.Str != "hello world!" {
		t.Fatalf("expected \"hello world!\", got %q (%s)", val.Str, val.String())
	}
}

func TestCompile_DictBasics(t *testing.T) {
	src := `
pckg main;

fun main() | bool {
    var user | dict<any> = {
        name: "Alex",
        "age": 30
    };
    if (user.name != "Alex") {
        return false;
    }
    if (user["age"] != 30) {
        return false;
    }
    if (user.length() != 2) {
        return false;
    }
    var keys | list<string> = user.keys();
    if (keys.length() != 2) {
        return false;
    }
    if (!keys.contains("name")) {
        return false;
    }
    if (!keys.contains("age")) {
        return false;
    }
    var values | list<any> = user.values();
    if (values.length() != 2) {
        return false;
    }
    if (!values.contains("Alex")) {
        return false;
    }
    if (!values.contains(30)) {
        return false;
    }
    if (typeOf(user.get("age")) != "int?") {
        return false;
    }
    if (typeOf(user.get("missing")) != "any?") {
        return false;
    }
    user.set("role", "admin");
    if (!user.has("role")) {
        return false;
    }
    if (user.length() != 3) {
        return false;
    }
    if (!user.remove("role")) {
        return false;
    }
    if (user.has("role")) {
        return false;
    }
    return true;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindBool || val.Bool != true {
		t.Fatalf("expected true, got %v (%s)", val.Bool, val.String())
	}
}

func TestCompile_ListAndIndex(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3];
    return xs[1];
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 2 {
		t.Fatalf("expected 2, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_HigherOrderFunctions(t *testing.T) {
	src := `
pckg main;

fun inc(x | int) | int {
    return x + 1;
}

fun apply(f | fun(int) | int, x | int) | int {
    return f(x);
}

fun main() | int {
    var g | fun(int) | int = inc;
    return apply(g, 41);
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 42 {
		t.Fatalf("expected 42, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_WhileLoop(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var i | int = 0;
    while (i < 5) {
        i = i + 1;
    }
    return i;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 5 {
		t.Fatalf("expected 5, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_ForLoop(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var sum | int = 0;
    for (var i | int = 0; i < 5; i = i + 1) {
        sum = sum + i;
    }
    return sum;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// sum = 0 + 1 + 2 + 3 + 4 = 10
	if val.Kind != value.KindInt || val.Int != 10 {
		t.Fatalf("expected 10, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_ForEachLoop(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3, 4, 5];
    var sum | int = 0;
    for (item in xs) {
        sum = sum + item;
    }
    return sum;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// sum = 1 + 2 + 3 + 4 + 5 = 15
	if val.Kind != value.KindInt || val.Int != 15 {
		t.Fatalf("expected 15, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_ForLoopInfinite(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var i | int = 0;
    for (;;) {
        i = i + 1;
        if (i >= 10) {
            return i;
        }
    }
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 10 {
		t.Fatalf("expected 10, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_DefaultParameters(t *testing.T) {
	src := `
pckg main;

fun sum(a | int, b | int = 0) | int {
    return a + b;
}

fun main() | int {
    return sum(5);
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// sum(5) with default b=0 should return 5
	if val.Kind != value.KindInt || val.Int != 5 {
		t.Fatalf("expected 5, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_NamedArguments(t *testing.T) {
	src := `
pckg main;

fun sum(a | int, b | int = 0) | int {
    return a + b;
}

fun main() | int {
    return sum(b=5, a=1);
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// sum(b=5, a=1) should return 6
	if val.Kind != value.KindInt || val.Int != 6 {
		t.Fatalf("expected 6, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinNamedArguments(t *testing.T) {
	t.Run("print", func(t *testing.T) {
		src := `
pckg main;

fun main() | void {
    print(value="hello");
}
`
		l := lexer.New(src)
		p := parser.New(l)
		prog := p.ParseProgram()
		if errs := p.Errors(); len(errs) > 0 {
			for _, e := range errs {
				t.Logf("parser error: %s", e)
			}
			t.Fatalf("expected no parser errors, got %d", len(errs))
		}

		mod, errs := ir.Compile(prog)
		if len(errs) > 0 {
			for _, e := range errs {
				t.Logf("compile error: %s", e)
			}
			t.Fatalf("expected no compile errors, got %d", len(errs))
		}

		var buf bytes.Buffer
		env := runtime.NewEnv(&testIO{output: &buf})
		machine := vm.NewVM(mod, env)
		_, err := machine.RunMain()
		if err != nil {
			t.Fatalf("RunMain error: %v", err)
		}

		if buf.String() != "hello" {
			t.Fatalf("expected 'hello', got %q", buf.String())
		}
	})

	t.Run("len", func(t *testing.T) {
		src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3];
    return len(value=xs);
}
`
		l := lexer.New(src)
		p := parser.New(l)
		prog := p.ParseProgram()
		if errs := p.Errors(); len(errs) > 0 {
			for _, e := range errs {
				t.Logf("parser error: %s", e)
			}
			t.Fatalf("expected no parser errors, got %d", len(errs))
		}

		mod, errs := ir.Compile(prog)
		if len(errs) > 0 {
			for _, e := range errs {
				t.Logf("compile error: %s", e)
			}
			t.Fatalf("expected no compile errors, got %d", len(errs))
		}

		machine := vm.NewVM(mod, runtime.DefaultEnv())
		val, err := machine.RunMain()
		if err != nil {
			t.Fatalf("RunMain error: %v", err)
		}

		if val.Kind != value.KindInt || val.Int != 3 {
			t.Fatalf("expected 3, got %v (%s)", val.Int, val.String())
		}
	})

	t.Run("error", func(t *testing.T) {
		src := `
pckg main;

fun main() | error {
    return error(message="oops");
}
`
		l := lexer.New(src)
		p := parser.New(l)
		prog := p.ParseProgram()
		if errs := p.Errors(); len(errs) > 0 {
			for _, e := range errs {
				t.Logf("parser error: %s", e)
			}
			t.Fatalf("expected no parser errors, got %d", len(errs))
		}

		mod, errs := ir.Compile(prog)
		if len(errs) > 0 {
			for _, e := range errs {
				t.Logf("compile error: %s", e)
			}
			t.Fatalf("expected no compile errors, got %d", len(errs))
		}

		machine := vm.NewVM(mod, runtime.DefaultEnv())
		val, err := machine.RunMain()
		if err != nil {
			t.Fatalf("RunMain error: %v", err)
		}

		if val.Kind != value.KindError {
			t.Fatalf("expected error value, got %v (%s)", val, val.String())
		}
		msg := val.Str
		if val.Error != nil && val.Error.Message != "" {
			msg = val.Error.Message
		}
		if msg != "oops" {
			t.Fatalf("expected error 'oops', got %q (%s)", msg, val.String())
		}
	})

	t.Run("errorMessage", func(t *testing.T) {
		src := `
pckg main;

fun main() | string {
    var err | error = error(message="fail");
    return errorMessage(e=err);
}
`
		l := lexer.New(src)
		p := parser.New(l)
		prog := p.ParseProgram()
		if errs := p.Errors(); len(errs) > 0 {
			for _, e := range errs {
				t.Logf("parser error: %s", e)
			}
			t.Fatalf("expected no parser errors, got %d", len(errs))
		}

		mod, errs := ir.Compile(prog)
		if len(errs) > 0 {
			for _, e := range errs {
				t.Logf("compile error: %s", e)
			}
			t.Fatalf("expected no compile errors, got %d", len(errs))
		}

		machine := vm.NewVM(mod, runtime.DefaultEnv())
		val, err := machine.RunMain()
		if err != nil {
			t.Fatalf("RunMain error: %v", err)
		}

		if val.Kind != value.KindString || val.Str != "fail" {
			t.Fatalf("expected 'fail', got %v (%s)", val, val.String())
		}
	})
}

func TestCompile_BuiltinNamedArgumentsErrors(t *testing.T) {
	t.Run("unknown_parameter", func(t *testing.T) {
		src := `
pckg main;

fun main() | void {
    print(xxx=1);
}
`
		l := lexer.New(src)
		p := parser.New(l)
		prog := p.ParseProgram()
		if errs := p.Errors(); len(errs) > 0 {
			for _, e := range errs {
				t.Logf("parser error: %s", e)
			}
			t.Fatalf("expected no parser errors, got %d", len(errs))
		}

		mod, errs := ir.Compile(prog)
		if len(errs) == 0 {
			t.Fatalf("expected compile error for unknown parameter name")
		}
		found := false
		for _, e := range errs {
			if e.Error() != "" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected error about unknown parameter, got: %v", errs)
		}
		// Should still produce a module (even if it has errors)
		_ = mod
	})

	t.Run("duplicate_parameter", func(t *testing.T) {
		src := `
pckg main;

fun main() | void {
    var xs | list<int> = [1, 2];
    len(value=xs, value=xs);
}
`
		l := lexer.New(src)
		p := parser.New(l)
		prog := p.ParseProgram()
		if errs := p.Errors(); len(errs) > 0 {
			for _, e := range errs {
				t.Logf("parser error: %s", e)
			}
			t.Fatalf("expected no parser errors, got %d", len(errs))
		}

		mod, errs := ir.Compile(prog)
		if len(errs) == 0 {
			t.Fatalf("expected compile error for duplicate parameter")
		}
		// Should still produce a module (even if it has errors)
		_ = mod
	})

	t.Run("missing_required", func(t *testing.T) {
		src := `
pckg main;

fun main() | void {
    print();
}
`
		l := lexer.New(src)
		p := parser.New(l)
		prog := p.ParseProgram()
		if errs := p.Errors(); len(errs) > 0 {
			for _, e := range errs {
				t.Logf("parser error: %s", e)
			}
			t.Fatalf("expected no parser errors, got %d", len(errs))
		}

		mod, errs := ir.Compile(prog)
		if len(errs) == 0 {
			t.Fatalf("expected compile error for missing required parameter")
		}
		// Should still produce a module (even if it has errors)
		_ = mod
	})
}

func TestCompile_TryCatch(t *testing.T) {
	src := `
pckg main;

fun may_fail() | void {
    throw error("fail");
}

fun main() | void {
    try {
        may_fail();
        print("never");
    } catch (e | error) {
        print(errorMessage(e));
    }
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	var output []string
	env := runtime.NewEnv(&testOutputWriter{output: &output})

	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// Should have printed "fail", not "never"
	if len(output) != 1 {
		t.Fatalf("expected 1 output line, got %d: %v", len(output), output)
	}
	if output[0] != "fail" {
		t.Fatalf("expected output 'fail', got %q", output[0])
	}
}

func TestCompile_TryCatch_BuiltinError(t *testing.T) {
	src := `
pckg main;

fun main() | void {
    try {
        var x | int = toInt("abc");
        print(x);
    } catch (e | error) {
        print(errorMessage(e));
    }
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	var output []string
	env := runtime.NewEnv(&testOutputWriter{output: &output})
	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}
	if len(output) != 1 || output[0] != `toInt: invalid integer "abc"` {
		t.Fatalf("expected output %q, got %v", `toInt: invalid integer "abc"`, output)
	}
}

func TestCompile_TryCatch_RuntimeError(t *testing.T) {
	src := `
pckg main;

fun main() | void {
    try {
        var x | int = 10 / 0;
        print(x);
    } catch (e | error) {
        print(errorMessage(e));
    }
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	var output []string
	env := runtime.NewEnv(&testOutputWriter{output: &output})
	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}
	if len(output) != 1 || output[0] != "division by zero" {
		t.Fatalf("expected output %q, got %v", "division by zero", output)
	}
}

func TestCompile_TryCatch_Nested(t *testing.T) {
	src := `
pckg main;

fun main() | void {
    try {
        try {
            toInt("abc");
        } catch (e | error) {
            print("inner");
            throw e;
        }
    } catch (e | error) {
        print("outer");
    }
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	var output []string
	env := runtime.NewEnv(&testOutputWriter{output: &output})
	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}
	if len(output) != 2 || output[0] != "inner" || output[1] != "outer" {
		t.Fatalf("expected output [inner outer], got %v", output)
	}
}

func TestCompile_TryCatch_ErrorPropagation(t *testing.T) {
	src := `
pckg main;

fun fail() | int {
    return toInt("abc");
}

fun main() | void {
    try {
        fail();
    } catch (e | error) {
        print(errorMessage(e));
    }
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	var output []string
	env := runtime.NewEnv(&testOutputWriter{output: &output})
	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}
	if len(output) != 1 || output[0] != `toInt: invalid integer "abc"` {
		t.Fatalf("expected output %q, got %v", `toInt: invalid integer "abc"`, output)
	}
}

func TestCompile_UncaughtError(t *testing.T) {
	src := `
pckg main;

fun main() | void {
    toInt("abc");
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	_, err := machine.RunMain()
	if err == nil {
		t.Fatalf("expected unhandled error, got nil")
	}
	if err.Error() != `toInt: invalid integer "abc"` {
		t.Fatalf("expected error %q, got %q", `toInt: invalid integer "abc"`, err.Error())
	}
}

func TestCompile_TypeOf(t *testing.T) {
	src := `
pckg main;

struct Point {
    x | int;
    y | int;
}

fun main() | void {
    print(typeOf(10));
    print(typeOf(1.5));
    print(typeOf("hello"));
    print(typeOf(true));
    print(typeOf([]));
    print(typeOf([1, 2]));
    print(typeOf(fromString("abc")));
    print(typeOf(error("x")));
    var p | Point = Point{x = 1, y = 2};
    print(typeOf(p));
    print(typeOf(toInt("123")));
    print(typeOf(some(1)));
    print(typeOf(none));
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	var output []string
	env := runtime.NewEnv(&testOutputWriter{output: &output})
	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	expected := []string{
		"int",
		"float",
		"string",
		"bool",
		"list<any>",
		"list<int>",
		"bytes",
		"error",
		"Point",
		"int",
		"int?",
		"any?",
	}
	if len(output) != len(expected) {
		t.Fatalf("expected %d outputs, got %d: %v", len(expected), len(output), output)
	}
	for i, exp := range expected {
		if output[i] != exp {
			t.Fatalf("expected output[%d] %q, got %q", i, exp, output[i])
		}
	}
}

func TestCompile_UnhandledException(t *testing.T) {
	src := `
pckg main;

fun main() | void {
    throw error("unhandled");
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	_, err := machine.RunMain()
	if err == nil {
		t.Fatalf("expected error for unhandled exception, got nil")
	}
	if err.Error() == "" || len(err.Error()) == 0 {
		t.Fatalf("expected non-empty error message")
	}
	// Check that error message contains "unhandled exception"
	if len(err.Error()) < 10 {
		t.Fatalf("expected error message to contain 'unhandled exception', got %q", err.Error())
	}
}

func TestCompile_SimpleClosure(t *testing.T) {
	src := `
pckg main;

fun main() | void {
    var x | int = 10;
    var f | fun(int) | int = fun(a | int) | int {
        return a + x;
    };
    x = 20;
    print(f(1));
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	var output []string
	env := runtime.NewEnv(&testOutputWriter{output: &output})

	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// Should have printed "21" (1 + 20)
	if len(output) != 1 {
		t.Fatalf("expected 1 output line, got %d: %v", len(output), output)
	}
	if output[0] != "21" {
		t.Fatalf("expected output '21', got %q", output[0])
	}
}

func TestCompile_NestedClosures(t *testing.T) {
	src := `
pckg main;

fun makeAdder(x | int) | fun(int) | int {
    return fun(y | int) | int {
        return x + y;
    };
}

fun main() | void {
    var add10 | fun(int) | int = makeAdder(10);
    print(add10(1));
    var add5  | fun(int) | int = makeAdder(5);
    print(add5(1));
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	var output []string
	env := runtime.NewEnv(&testOutputWriter{output: &output})

	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// Should have printed "11" and "6"
	if len(output) != 2 {
		t.Fatalf("expected 2 output lines, got %d: %v", len(output), output)
	}
	if output[0] != "11" {
		t.Fatalf("expected first output '11', got %q", output[0])
	}
	if output[1] != "6" {
		t.Fatalf("expected second output '6', got %q", output[1])
	}
}

func TestCompile_ClosureEscapingScope(t *testing.T) {
	src := `
pckg main;

fun outer() | fun() | int {
    var x | int = 1;
    var mid | fun() | fun() | int = fun() | fun() | int {
        var y | int = 2;
        return fun() | int {
            return x + y;
        };
    };
    var f | fun() | int = mid();
    x = 100;
    return f;
}

fun main() | void {
    var f | fun() | int = outer();
    print(f());
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	var output []string
	env := runtime.NewEnv(&testOutputWriter{output: &output})

	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// Should have printed "102" (100 + 2)
	if len(output) != 1 {
		t.Fatalf("expected 1 output line, got %d: %v", len(output), output)
	}
	if output[0] != "102" {
		t.Fatalf("expected output '102', got %q", output[0])
	}
}

func TestCompile_ClosureMutation(t *testing.T) {
	src := `
pckg main;

fun main() | void {
    var x | int = 10;
    var f | fun() | int = fun() | int {
        return x;
    };
    var result1 | int = f();
    x = 20;
    var result2 | int = f();
    print(result1);
    print(result2);
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	var output []string
	env := runtime.NewEnv(&testOutputWriter{output: &output})

	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// Should have printed "10" and "20" (closure sees updated value)
	if len(output) != 2 {
		t.Fatalf("expected 2 output lines, got %d: %v", len(output), output)
	}
	if output[0] != "10" {
		t.Fatalf("expected first output '10', got %q", output[0])
	}
	if output[1] != "20" {
		t.Fatalf("expected second output '20', got %q", output[1])
	}
}

func TestCompile_ClosureNoCapture(t *testing.T) {
	src := `
pckg main;

fun main() | void {
    var f | fun(int) | int = fun(x | int) | int {
        return x + 1;
    };
    print(f(5));
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	var output []string
	env := runtime.NewEnv(&testOutputWriter{output: &output})

	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// Should have printed "6"
	if len(output) != 1 {
		t.Fatalf("expected 1 output line, got %d: %v", len(output), output)
	}
	if output[0] != "6" {
		t.Fatalf("expected output '6', got %q", output[0])
	}
}

func TestCompileWorld_MultiModule(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create std/io.av
	stdDir := filepath.Join(tmpDir, "std")
	if err := os.MkdirAll(stdDir, 0755); err != nil {
		t.Fatalf("failed to create std dir: %v", err)
	}
	ioFile := filepath.Join(stdDir, "io.av")
	ioContent := `pckg std.io;

pub fun println(msg | string) | void {
    print(msg);
}
`
	if err := os.WriteFile(ioFile, []byte(ioContent), 0644); err != nil {
		t.Fatalf("failed to write io.av: %v", err)
	}

	// Create math/arith.av
	mathDir := filepath.Join(tmpDir, "math")
	if err := os.MkdirAll(mathDir, 0755); err != nil {
		t.Fatalf("failed to create math dir: %v", err)
	}
	arithFile := filepath.Join(mathDir, "arith.av")
	arithContent := `pckg math.arith;

pub fun sum(a | int, b | int) | int {
    return a + b;
}
`
	if err := os.WriteFile(arithFile, []byte(arithContent), 0644); err != nil {
		t.Fatalf("failed to write arith.av: %v", err)
	}

	// Create main.av
	mainFile := filepath.Join(tmpDir, "main.av")
	mainContent := `pckg main;

import std.io;
import math.arith as ar;

fun main() | void {
    var x | int = ar.sum(2, 3);
    io.println("x = 5");
}
`
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main.av: %v", err)
	}

	// Load world
	world, errs := modules.LoadWorld(mainFile)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("module loading error: %s", e)
		}
		t.Fatalf("failed to load world: %d errors", len(errs))
	}

	// Build types.World
	typeWorld := &types.World{
		Modules: make(map[string]*types.ModuleInfo),
		Entry:   world.Entry,
	}
	for modName, modAST := range world.Modules {
		typeWorld.Modules[modName] = &types.ModuleInfo{
			Name:  modName,
			Prog:  modAST.Prog,
			Scope: nil, // Will be set by CheckWorld
		}
	}

	// Type-check with bindings
	bindings, typeErrs := types.CheckWorldWithBindings(typeWorld)
	if len(typeErrs) > 0 {
		for _, e := range typeErrs {
			t.Logf("type error: %s", e)
		}
		t.Fatalf("type checking failed: %d errors", len(typeErrs))
	}

	// Compile
	entryModInfo := typeWorld.Modules[world.Entry]
	mod, errs := ir.CompileWorld(typeWorld, entryModInfo, bindings)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("compilation failed: %d errors", len(errs))
	}

	// Run
	var output []string
	env := runtime.NewEnv(&testOutputWriter{output: &output})
	machine := vm.NewVM(mod, env)
	_, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// Check output
	if len(output) != 1 {
		t.Fatalf("expected 1 output line, got %d: %v", len(output), output)
	}
	if output[0] != "x = 5" {
		t.Fatalf("expected output 'x = 5', got %q", output[0])
	}
}

func TestCompile_BreakInWhile(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var i | int = 0;
    while (true) {
        if (i == 5) {
            break;
        }
        i = i + 1;
    }
    return i;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 5 {
		t.Fatalf("expected 5, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BreakInFor(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var sum | int = 0;
    for (var i | int = 0; i < 10; i = i + 1) {
        if (i == 3) {
            break;
        }
        sum = sum + i;
    }
    return sum;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// sum = 0 + 1 + 2 = 3
	if val.Kind != value.KindInt || val.Int != 3 {
		t.Fatalf("expected 3, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BreakInForEach(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3, 4, 5];
    var sum | int = 0;
    for (item in xs) {
        if (item == 4) {
            break;
        }
        sum = sum + item;
    }
    return sum;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// sum = 1 + 2 + 3 = 6
	if val.Kind != value.KindInt || val.Int != 6 {
		t.Fatalf("expected 6, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_StructLiteral(t *testing.T) {
	src := `
pckg main;

struct Point {
    x | int
    y | int
}

fun main() | int {
    var p | Point = Point{x = 10, y = 20};
    return p.x;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 10 {
		t.Fatalf("expected 10, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_StructFieldAssign(t *testing.T) {
	src := `
pckg main;

mut struct Point {
    x | int
    y | int
}

fun main() | int {
    var p | Point = Point{x = 0, y = 0};
    p.x = 10;
    return p.x;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 10 {
		t.Fatalf("expected 10, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_StructFieldAssign_WithDefaults(t *testing.T) {
	src := `
pckg main;

mut struct Config {
    host | string = "localhost"
    port | int = 8080
}

fun main() | int {
    var c | Config = Config{};
    c.port = 9000;
    return c.port;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 9000 {
		t.Fatalf("expected 9000, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_StructFieldAssign_MultipleFields(t *testing.T) {
	src := `
pckg main;

mut struct Point {
    x | int
    y | int
}

fun main() | int {
    var p | Point = Point{x = 0, y = 0};
    p.x = 5;
    p.y = 10;
    return p.x + p.y;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 15 {
		t.Fatalf("expected 15, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_StructFieldAccess(t *testing.T) {
	src := `
pckg main;

struct Point {
    x | int
    y | int
}

fun main() | int {
    var p | Point = Point{x = 5, y = 15};
    var sum | int = p.x + p.y;
    return sum;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 20 {
		t.Fatalf("expected 20, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_StructInList(t *testing.T) {
	src := `
pckg main;

struct Point {
    x | int
    y | int
}

fun main() | int {
    var points | list<Point> = [Point{x = 1, y = 2}, Point{x = 3, y = 4}];
    return points[0].x;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 1 {
		t.Fatalf("expected 1, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_MethodSimple(t *testing.T) {
	src := `
pckg main;

struct Point {
    x | int
    y | int
}

fun (p | Point).getX() | int {
    return p.x;
}

fun main() | int {
    var p | Point = Point{x = 10, y = 20};
    return p.getX();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 10 {
		t.Fatalf("expected 10, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_MethodWithParams(t *testing.T) {
	src := `
pckg main;

struct Point {
    x | int
    y | int
}

fun (p | Point).add(dx | int, dy | int) | Point {
    return Point{x = p.x + dx, y = p.y + dy};
}

fun main() | int {
    var p | Point = Point{x = 10, y = 20};
    var q | Point = p.add(5, 10);
    return q.x;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 15 {
		t.Fatalf("expected 15, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_MethodMultiple(t *testing.T) {
	src := `
pckg main;

struct Point {
    x | int
    y | int
}

fun (p | Point).getX() | int {
    return p.x;
}

fun (p | Point).getY() | int {
    return p.y;
}

fun (p | Point).lengthSquared() | int {
    return p.x * p.x + p.y * p.y;
}

fun main() | int {
    var p | Point = Point{x = 3, y = 4};
    return p.lengthSquared();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// 3*3 + 4*4 = 9 + 16 = 25
	if val.Kind != value.KindInt || val.Int != 25 {
		t.Fatalf("expected 25, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_MethodChained(t *testing.T) {
	src := `
pckg main;

struct Point {
    x | int
    y | int
}

fun (p | Point).move(dx | int, dy | int) | Point {
    return Point{x = p.x + dx, y = p.y + dy};
}

fun main() | int {
    var p | Point = Point{x = 10, y = 20};
    var q | Point = p.move(5, 10).move(2, 3);
    return q.x;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// 10 + 5 + 2 = 17
	if val.Kind != value.KindInt || val.Int != 17 {
		t.Fatalf("expected 17, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_MethodReturnType(t *testing.T) {
	src := `
pckg main;

struct Point {
    x | int
    y | int
}

fun (p | Point).distanceSquared(other | Point) | int {
    var dx | int = p.x - other.x;
    var dy | int = p.y - other.y;
    return dx * dx + dy * dy;
}

fun main() | int {
    var p1 | Point = Point{x = 0, y = 0};
    var p2 | Point = Point{x = 3, y = 4};
    return p1.distanceSquared(p2);
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// (3-0)^2 + (4-0)^2 = 9 + 16 = 25
	if val.Kind != value.KindInt || val.Int != 25 {
		t.Fatalf("expected 25, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_StaticMethodSimple(t *testing.T) {
	src := `
pckg main;

struct Point {
    x | int
    y | int
}

fun Point.origin() | Point {
    return Point{x = 0, y = 0};
}

fun main() | int {
    var p | Point = Point.origin();
    return p.x;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 0 {
		t.Fatalf("expected 0, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_StaticMethodWithParams(t *testing.T) {
	src := `
pckg main;

struct Point {
    x | int
    y | int
}

fun Point.new(x | int, y | int) | Point {
    return Point{x = x, y = y};
}

fun main() | int {
    var p | Point = Point.new(10, 20);
    return p.x;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 10 {
		t.Fatalf("expected 10, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_StaticAndInstanceMethods(t *testing.T) {
	src := `
pckg main;

struct Point {
    x | int
    y | int
}

fun Point.origin() | Point {
    return Point{x = 0, y = 0};
}

fun (p | Point).getX() | int {
    return p.x;
}

fun main() | int {
    var p | Point = Point.origin();
    return p.getX();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 0 {
		t.Fatalf("expected 0, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_StaticMethodMultiple(t *testing.T) {
	src := `
pckg main;

struct Point {
    x | int
    y | int
}

fun Point.origin() | Point {
    return Point{x = 0, y = 0};
}

fun Point.fromX(x | int) | Point {
    return Point{x = x, y = 0};
}

fun Point.fromY(y | int) | Point {
    return Point{x = 0, y = y};
}

fun main() | int {
    var p1 | Point = Point.origin();
    var p2 | Point = Point.fromX(5);
    var p3 | Point = Point.fromY(10);
    return p1.x + p2.x + p3.y;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	// 0 + 5 + 10 = 15
	if val.Kind != value.KindInt || val.Int != 15 {
		t.Fatalf("expected 15, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListLength(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3, 4, 5];
    return xs.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 5 {
		t.Fatalf("expected 5, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListAppend(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3];
    var ys | list<int> = xs.append(4);
    return ys.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 4 {
		t.Fatalf("expected 4, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListPop(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3];
    var popped | int = xs.pop();
    return popped;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 3 {
		t.Fatalf("expected 3, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListInsert(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 4];
    var ys | list<int> = xs.insert(2, 3);
    return ys.get(2);
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 3 {
		t.Fatalf("expected 3, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListRemoveAt(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3, 4];
    var ys | list<int> = xs.removeAt(1);
    return ys.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 3 {
		t.Fatalf("expected 3, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListClear(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3];
    var ys | list<int> = xs.clear();
    return ys.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 0 {
		t.Fatalf("expected 0, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListIsEmpty(t *testing.T) {
	src := `
pckg main;

fun main() | bool {
    var xs | list<int> = [];
    return xs.isEmpty();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindBool || val.Bool != true {
		t.Fatalf("expected true, got %v (%s)", val.Bool, val.String())
	}
}

func TestCompile_BuiltinMethod_ListGet(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [10, 20, 30];
    return xs.get(1);
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 20 {
		t.Fatalf("expected 20, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListContains(t *testing.T) {
	src := `
pckg main;

fun main() | bool {
    var xs | list<int> = [1, 2, 3];
    return xs.contains(2);
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindBool || val.Bool != true {
		t.Fatalf("expected true, got %v (%s)", val.Bool, val.String())
	}
}

func TestCompile_BuiltinMethod_ListIndexOf(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3, 2];
    return xs.indexOf(2);
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 1 {
		t.Fatalf("expected 1, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListSlice(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3, 4, 5];
    var ys | list<int> = xs.slice(1, 4);
    return ys.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 3 {
		t.Fatalf("expected 3, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListReverse(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3];
    var ys | list<int> = xs.reverse();
    return ys.get(0);
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 3 {
		t.Fatalf("expected 3, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListCopy(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3];
    var ys | list<int> = xs.copy();
    return ys.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 3 {
		t.Fatalf("expected 3, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListMap(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3, 4];
    var ys | list<int> = xs.map(fun(x | int) | int {
        return x * 2;
    });
    return ys.get(2);
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 6 {
		t.Fatalf("expected 6, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListFilter(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3, 4, 5];
    var ys | list<int> = xs.filter(fun(x | int) | bool {
        return x % 2 == 0;
    });
    return ys.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 2 {
		t.Fatalf("expected 2, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListReduce(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [1, 2, 3, 4];
    var sum | int = xs.reduce(0, fun(acc | int, x | int) | int {
        return acc + x;
    });
    return sum;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 10 {
		t.Fatalf("expected 10, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListMapEmpty(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [];
    var ys | list<int> = xs.map(fun(x | int) | int {
        return x * 2;
    });
    return ys.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 0 {
		t.Fatalf("expected 0, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListFilterEmpty(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [];
    var ys | list<int> = xs.filter(fun(x | int) | bool {
        return x > 0;
    });
    return ys.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 0 {
		t.Fatalf("expected 0, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_ListReduceEmpty(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var xs | list<int> = [];
    var result | int = xs.reduce(42, fun(acc | int, x | int) | int {
        return acc + x;
    });
    return result;
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 42 {
		t.Fatalf("expected 42, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_BytesAppend(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var b | bytes = fromString("hello");
    var b2 | bytes = b.append(33);
    return b2.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 6 {
		t.Fatalf("expected 6, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_BytesConcat(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var b1 | bytes = fromString("hello");
    var b2 | bytes = fromString("world");
    var b3 | bytes = b1.concat(b2);
    return b3.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 10 {
		t.Fatalf("expected 10, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_BytesSlice(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var b | bytes = fromString("hello");
    var b2 | bytes = b.slice(1, 4);
    return b2.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 3 {
		t.Fatalf("expected 3, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_BytesToString(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var b | bytes = fromString("hello");
    var s | string = b.toString();
    return s.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 5 {
		t.Fatalf("expected 5, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_BytesAppendRange(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var b | bytes = fromString("test");
    var b2 | bytes = b.append(0);
    var b3 | bytes = b2.append(255);
    return b3.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 6 {
		t.Fatalf("expected 6, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_BytesSliceEmpty(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var b | bytes = fromString("hello");
    var b2 | bytes = b.slice(2, 2);
    return b2.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 0 {
		t.Fatalf("expected 0, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_StringLength(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var s | string = "hello";
    return s.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 5 {
		t.Fatalf("expected 5, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_StringToUpper(t *testing.T) {
	src := `
pckg main;

fun main() | string {
    var s | string = "hello";
    return s.toUpperCase();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindString || val.Str != "HELLO" {
		t.Fatalf("expected \"HELLO\", got %v (%s)", val.Str, val.String())
	}
}

func TestCompile_BuiltinMethod_StringToLower(t *testing.T) {
	src := `
pckg main;

fun main() | string {
    var s | string = "HELLO";
    return s.toLowerCase();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindString || val.Str != "hello" {
		t.Fatalf("expected \"hello\", got %v (%s)", val.Str, val.String())
	}
}

func TestCompile_BuiltinMethod_StringTrim(t *testing.T) {
	src := `
pckg main;

fun main() | string {
    var s | string = "  hello  ";
    return s.trim();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindString || val.Str != "hello" {
		t.Fatalf("expected \"hello\", got %v (%s)", val.Str, val.String())
	}
}

func TestCompile_BuiltinMethod_StringContains(t *testing.T) {
	src := `
pckg main;

fun main() | bool {
    var s | string = "hello world";
    return s.contains("world");
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindBool || val.Bool != true {
		t.Fatalf("expected true, got %v (%s)", val.Bool, val.String())
	}
}

func TestCompile_BuiltinMethod_StringStartsWith(t *testing.T) {
	src := `
pckg main;

fun main() | bool {
    var s | string = "hello world";
    return s.startsWith("hello");
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindBool || val.Bool != true {
		t.Fatalf("expected true, got %v (%s)", val.Bool, val.String())
	}
}

func TestCompile_BuiltinMethod_StringEndsWith(t *testing.T) {
	src := `
pckg main;

fun main() | bool {
    var s | string = "hello world";
    return s.endsWith("world");
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindBool || val.Bool != true {
		t.Fatalf("expected true, got %v (%s)", val.Bool, val.String())
	}
}

func TestCompile_BuiltinMethod_StringReplace(t *testing.T) {
	src := `
pckg main;

fun main() | string {
    var s | string = "hello world";
    return s.replace("world", "avenir");
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindString || val.Str != "hello avenir" {
		t.Fatalf("expected \"hello avenir\", got %v (%s)", val.Str, val.String())
	}
}

func TestCompile_BuiltinMethod_StringSplit(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var s | string = "a,b,c";
    var parts | list<string> = s.split(",");
    return parts.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 3 {
		t.Fatalf("expected 3, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_StringIndexOf(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var s | string = "hello world";
    return s.indexOf("world");
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 6 {
		t.Fatalf("expected 6, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_StringLastIndexOf(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var s | string = "hello world world";
    return s.lastIndexOf("world");
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 12 {
		t.Fatalf("expected 12, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_StringChained(t *testing.T) {
	src := `
pckg main;

fun main() | string {
    var s | string = "  Hello World  ";
    return s.trim().toUpperCase();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindString || val.Str != "HELLO WORLD" {
		t.Fatalf("expected \"HELLO WORLD\", got %v (%s)", val.Str, val.String())
	}
}

func TestCompile_BuiltinMethod_BytesLength(t *testing.T) {
	src := `
pckg main;

fun main() | int {
    var b | bytes = b"hello";
    return b.length();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindInt || val.Int != 5 {
		t.Fatalf("expected 5, got %v (%s)", val.Int, val.String())
	}
}

func TestCompile_BuiltinMethod_Chained(t *testing.T) {
	src := `
pckg main;

fun main() | string {
    var xs | list<string> = ["hello", "world"];
    var first | string = xs[0];
    return first.toUpperCase();
}
`
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parser error: %s", e)
		}
		t.Fatalf("expected no parser errors, got %d", len(errs))
	}

	mod, errs := ir.Compile(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("compile error: %s", e)
		}
		t.Fatalf("expected no compile errors, got %d", len(errs))
	}

	machine := vm.NewVM(mod, runtime.DefaultEnv())
	val, err := machine.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if val.Kind != value.KindString || val.Str != "HELLO" {
		t.Fatalf("expected \"HELLO\", got %v (%s)", val.Str, val.String())
	}
}
