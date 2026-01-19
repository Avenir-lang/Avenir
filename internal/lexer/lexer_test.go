package lexer_test

import (
	"testing"

	"avenir/internal/lexer"
	"avenir/internal/token"
)

func TestNextToken_BasicProgram(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var a | int = 10;
    var b | string = "Test";
    var c | bool = true;
}
`

	tests := []struct {
		kind token.Kind
		lit  string
	}{
		{token.Pckg, "pckg"},
		{token.Ident, "main"},
		{token.Semicolon, ";"},

		{token.Fun, "fun"},
		{token.Ident, "main"},
		{token.LParen, "("},
		{token.RParen, ")"},
		{token.Pipe, "|"},
		{token.VoidType, "void"},
		{token.LBrace, "{"},

		{token.Var, "var"},
		{token.Ident, "a"},
		{token.Pipe, "|"},
		{token.IntType, "int"},
		{token.Assign, "="},
		{token.Int, "10"},
		{token.Semicolon, ";"},

		{token.Var, "var"},
		{token.Ident, "b"},
		{token.Pipe, "|"},
		{token.StringType, "string"},
		{token.Assign, "="},
		{token.String, "Test"},
		{token.Semicolon, ";"},

		{token.Var, "var"},
		{token.Ident, "c"},
		{token.Pipe, "|"},
		{token.BoolType, "bool"},
		{token.Assign, "="},
		{token.True, "true"},
		{token.Semicolon, ";"},

		{token.RBrace, "}"},
		{token.EOF, ""},
	}

	l := lexer.New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Kind != tt.kind {
			t.Fatalf("tests[%d] - kind wrong. expected=%s, got=%s (lexeme=%q, pos=%+v)",
				i, tt.kind, tok.Kind, tok.Lexeme, tok.Pos)
		}

		if tok.Lexeme != tt.lit {
			t.Fatalf("tests[%d] - lexeme wrong. expected=%q, got=%q",
				i, tt.lit, tok.Lexeme)
		}
	}
}

func TestStringEscapes(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var s | string = "Line1\nLine2\tTabbed\"";
    var u | string = "\u0041\x42\0";
}
`
	l := lexer.New(input)

	var lastString token.Token
	for {
		tok := l.NextToken()
		if tok.Kind == token.String {
			lastString = tok
		}
		if tok.Kind == token.EOF {
			break
		}
	}

	if len(l.Errors()) > 0 {
		t.Fatalf("unexpected lexer errors: %v", l.Errors())
	}
	if lastString.Lexeme != "AB\x00" {
		t.Fatalf("expected decoded string %q, got %q", "AB\x00", lastString.Lexeme)
	}
}

func TestStringInterpolationTokens(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var s | string = "a${b}c";
}
`
	tests := []struct {
		kind token.Kind
		lit  string
	}{
		{token.Pckg, "pckg"},
		{token.Ident, "main"},
		{token.Semicolon, ";"},
		{token.Fun, "fun"},
		{token.Ident, "main"},
		{token.LParen, "("},
		{token.RParen, ")"},
		{token.Pipe, "|"},
		{token.VoidType, "void"},
		{token.LBrace, "{"},
		{token.Var, "var"},
		{token.Ident, "s"},
		{token.Pipe, "|"},
		{token.StringType, "string"},
		{token.Assign, "="},
		{token.StringPart, "a"},
		{token.InterpStart, "${"},
		{token.Ident, "b"},
		{token.InterpEnd, "}"},
		{token.StringPart, "c"},
		{token.StringEnd, ""},
		{token.Semicolon, ";"},
		{token.RBrace, "}"},
		{token.EOF, ""},
	}

	l := lexer.New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Kind != tt.kind {
			t.Fatalf("tests[%d] kind wrong. expected=%s, got=%s (lexeme=%q)", i, tt.kind, tok.Kind, tok.Lexeme)
		}
		if tok.Lexeme != tt.lit {
			t.Fatalf("tests[%d] lexeme wrong. expected=%q, got=%q", i, tt.lit, tok.Lexeme)
		}
	}
}

func TestInvalidEscape(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var s | string = "bad\q";
}
`
	l := lexer.New(input)
	for tok := l.NextToken(); tok.Kind != token.EOF; tok = l.NextToken() {
	}
	if len(l.Errors()) == 0 {
		t.Fatalf("expected lexer error for invalid escape, got none")
	}
}

func TestSingleQuotedStrings(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var a | string = 'abc';
    var b | string = 'a\nb';
    var c | string = 'quote: \'';
    var d | string = 'backslash: \\';
}
`
	l := lexer.New(input)

	var strings []string
	for tok := l.NextToken(); tok.Kind != token.EOF; tok = l.NextToken() {
		if tok.Kind == token.String {
			strings = append(strings, tok.Lexeme)
		}
	}

	if len(l.Errors()) > 0 {
		t.Fatalf("unexpected lexer errors: %v", l.Errors())
	}
	if len(strings) != 4 {
		t.Fatalf("expected 4 string literals, got %d", len(strings))
	}
	if strings[0] != "abc" {
		t.Fatalf("expected %q, got %q", "abc", strings[0])
	}
	if strings[1] != "a\nb" {
		t.Fatalf("expected %q, got %q", "a\nb", strings[1])
	}
	if strings[2] != "quote: '" {
		t.Fatalf("expected %q, got %q", "quote: '", strings[2])
	}
	if strings[3] != "backslash: \\" {
		t.Fatalf("expected %q, got %q", "backslash: \\", strings[3])
	}
}

func TestSingleQuotedInvalidEscape(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var s | string = 'bad\q';
}
`
	l := lexer.New(input)
	for tok := l.NextToken(); tok.Kind != token.EOF; tok = l.NextToken() {
	}
	if len(l.Errors()) == 0 {
		t.Fatalf("expected lexer error for invalid escape, got none")
	}
}

func TestSingleQuotedUnterminated(t *testing.T) {
	input := `pckg main;

fun main() | void {
    var s | string = 'unterminated;
}
`
	l := lexer.New(input)
	for tok := l.NextToken(); tok.Kind != token.EOF; tok = l.NextToken() {
	}
	if len(l.Errors()) == 0 {
		t.Fatalf("expected lexer error for unterminated string, got none")
	}
}
