package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"avenir/internal/token"
)

type Lexer struct {
	input []rune

	pos int

	ch   rune
	line int
	col  int

	pending         []token.Token
	inString        bool
	stringDelimiter rune
	inInterp        bool
	interpDepth     int
	stringHasInterp bool
	stringStartPos  token.Position
	errors          []string
}

func New(input string) *Lexer {
	l := &Lexer{
		input: []rune(input),
		line:  1,
		col:   0,
	}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	if len(l.pending) > 0 {
		tok := l.pending[0]
		l.pending = l.pending[1:]
		return tok
	}

	if l.inInterp {
		return l.nextInterpToken()
	}

	if l.inString {
		return l.nextStringToken()
	}

	l.skipWhitespaceAndComments()

	pos := token.Position{
		Line:   l.line,
		Column: l.col,
	}

	ch := l.ch

	// EOF
	if ch == 0 {
		return token.Token{
			Kind:   token.EOF,
			Lexeme: "",
			Pos:    pos,
		}
	}

	// Numbers
	if isDigit(ch) {
		lit := l.readNumber()
		// Check if it's a float (contains '.' or 'e'/'E')
		kind := token.Int
		for _, r := range lit {
			if r == '.' || r == 'e' || r == 'E' {
				kind = token.Float
				break
			}
		}
		return token.Token{
			Kind:   kind,
			Lexeme: lit,
			Pos:    pos,
		}
	}

	// Bytes literals: b"..."
	if ch == 'b' && l.peekChar() == '"' {
		l.readChar() // consume 'b'
		l.readChar() // consume opening quote
		lit, ok := l.readSimpleString('"')
		if !ok {
			return token.Token{Kind: token.Illegal, Lexeme: "", Pos: pos}
		}
		return token.Token{
			Kind:   token.Bytes,
			Lexeme: lit,
			Pos:    pos,
		}
	}

	// Identifiers / keywords
	if isLetter(ch) {
		lit := l.readIdentifier()
		kind := token.LookupIdent(lit)
		return token.Token{
			Kind:   kind,
			Lexeme: lit,
			Pos:    pos,
		}
	}

	// Strings
	if ch == '"' || ch == '\'' {
		l.inString = true
		l.stringDelimiter = ch
		l.stringHasInterp = false
		l.stringStartPos = pos
		l.readChar() // consume opening quote
		return l.nextStringToken()
	}

	// Single- and two-character tokens
	var kind token.Kind
	var lexeme string

	switch ch {
	case ';':
		kind = token.Semicolon
		lexeme = ";"
	case ',':
		kind = token.Comma
		lexeme = ","
	case '.':
		kind = token.Dot
		lexeme = "."
	case ':':
		kind = token.Colon
		lexeme = ":"
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			kind = token.OrOr
			lexeme = "||"
		} else {
			kind = token.Pipe
			lexeme = "|"
		}
	case '(':
		kind = token.LParen
		lexeme = "("
	case ')':
		kind = token.RParen
		lexeme = ")"
	case '{':
		kind = token.LBrace
		lexeme = "{"
	case '}':
		kind = token.RBrace
		lexeme = "}"
	case '[':
		kind = token.LBracket
		lexeme = "["
	case ']':
		kind = token.RBracket
		lexeme = "]"
	case '?':
		kind = token.Question
		lexeme = "?"
	case '+':
		kind = token.Plus
		lexeme = "+"
	case '-':
		kind = token.Minus
		lexeme = "-"
	case '*':
		kind = token.Star
		lexeme = "*"
	case '/':
		kind = token.Slash
		lexeme = "/"
	case '%':
		kind = token.Percent
		lexeme = "%"
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			kind = token.NotEq
			lexeme = "!="
		} else {
			kind = token.Bang
			lexeme = "!"
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			kind = token.AndAnd
			lexeme = "&&"
		} else {
			kind = token.Illegal
			lexeme = "&"
		}
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			kind = token.Eq
			lexeme = "=="
		} else {
			kind = token.Assign
			lexeme = "="
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			kind = token.LtEq
			lexeme = "<="
		} else {
			kind = token.Lt
			lexeme = "<"
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			kind = token.GtEq
			lexeme = ">="
		} else {
			kind = token.Gt
			lexeme = ">"
		}
	default:
		kind = token.Illegal
		lexeme = string(ch)
	}

	l.readChar()

	return token.Token{
		Kind:   kind,
		Lexeme: lexeme,
		Pos:    pos,
	}
}

// Helpers

func (l *Lexer) readChar() {
	if l.pos >= len(l.input) {
		l.ch = 0
		return
	}

	l.ch = l.input[l.pos]
	l.pos++

	if l.ch == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
}

func (l *Lexer) peekChar() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		// Skip whitespace
		for unicode.IsSpace(l.ch) {
			l.readChar()
		}

		// Comments
		if l.ch == '/' {
			switch l.peekChar() {
			case '/':
				// Single-line comment
				l.readChar() // '/'
				l.readChar() // second '/'
				for l.ch != '\n' && l.ch != 0 {
					l.readChar()
				}
				continue
			case '*':
				// Multi-line comment
				l.readChar() // '/'
				l.readChar() // '*'
				for {
					if l.ch == 0 {
						// EOF inside comment
						return
					}
					if l.ch == '*' && l.peekChar() == '/' {
						l.readChar() // '*'
						l.readChar() // '/'
						break
					}
					l.readChar()
				}
				continue
			}
		}

		break
	}
}

func (l *Lexer) readIdentifier() string {
	start := l.pos - 1 // current rune is already in l.ch
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return string(l.input[start : l.pos-1])
}

func (l *Lexer) readNumber() string {
	start := l.pos - 1
	for isDigit(l.ch) {
		l.readChar()
	}
	// Check for decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	// Check for exponent
	if l.ch == 'e' || l.ch == 'E' {
		l.readChar() // consume 'e' or 'E'
		if l.ch == '+' || l.ch == '-' {
			l.readChar() // consume sign
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return string(l.input[start : l.pos-1])
}

func (l *Lexer) nextStringToken() token.Token {
	startPos := l.stringStartPos
	var sb []rune

	for {
		if l.ch == 0 || l.ch == '\n' {
			l.errorf(startPos, "unterminated string literal")
			l.inString = false
			return token.Token{Kind: token.Illegal, Lexeme: "", Pos: startPos}
		}
		if l.ch == l.stringDelimiter {
			l.readChar() // consume closing quote
			l.inString = false
			if l.stringHasInterp {
				if len(sb) > 0 {
					l.pending = append(l.pending, token.Token{
						Kind:   token.StringEnd,
						Lexeme: "",
						Pos:    startPos,
					})
					return token.Token{
						Kind:   token.StringPart,
						Lexeme: string(sb),
						Pos:    startPos,
					}
				}
				return token.Token{
					Kind:   token.StringEnd,
					Lexeme: "",
					Pos:    startPos,
				}
			}
			return token.Token{
				Kind:   token.String,
				Lexeme: string(sb),
				Pos:    startPos,
			}
		}
		if l.ch == '$' && l.peekChar() == '{' {
			l.stringHasInterp = true
			l.readChar() // consume '$'
			l.readChar() // consume '{'
			l.inInterp = true
			l.interpDepth = 1
			if len(sb) > 0 {
				l.pending = append(l.pending, token.Token{
					Kind:   token.InterpStart,
					Lexeme: "${",
					Pos:    startPos,
				})
				return token.Token{
					Kind:   token.StringPart,
					Lexeme: string(sb),
					Pos:    startPos,
				}
			}
			return token.Token{
				Kind:   token.InterpStart,
				Lexeme: "${",
				Pos:    startPos,
			}
		}
		if l.ch == '\\' {
			escPos := token.Position{Line: l.line, Column: l.col}
			l.readChar()
			r, ok := l.readEscape(escPos)
			if !ok {
				return token.Token{Kind: token.Illegal, Lexeme: "", Pos: escPos}
			}
			sb = append(sb, r)
			l.readChar()
			continue
		}
		sb = append(sb, l.ch)
		l.readChar()
	}
}

func (l *Lexer) nextInterpToken() token.Token {
	l.skipWhitespaceAndComments()

	pos := token.Position{
		Line:   l.line,
		Column: l.col,
	}
	ch := l.ch
	if ch == 0 {
		l.errorf(pos, "unterminated interpolation")
		l.inInterp = false
		return token.Token{Kind: token.Illegal, Lexeme: "", Pos: pos}
	}
	if ch == '{' {
		l.interpDepth++
		l.readChar()
		return token.Token{Kind: token.LBrace, Lexeme: "{", Pos: pos}
	}
	if ch == '}' {
		if l.interpDepth == 1 {
			l.readChar()
			l.inInterp = false
			return token.Token{Kind: token.InterpEnd, Lexeme: "}", Pos: pos}
		}
		l.interpDepth--
		l.readChar()
		return token.Token{Kind: token.RBrace, Lexeme: "}", Pos: pos}
	}

	// Strings inside interpolation (no nested interpolation)
	if ch == '"' || ch == '\'' {
		delimiter := ch
		l.readChar() // consume opening quote
		lit, ok := l.readSimpleString(delimiter)
		if !ok {
			return token.Token{Kind: token.Illegal, Lexeme: "", Pos: pos}
		}
		return token.Token{Kind: token.String, Lexeme: lit, Pos: pos}
	}
	if ch == 'b' && l.peekChar() == '"' {
		l.readChar() // consume 'b'
		l.readChar() // consume opening quote
		lit, ok := l.readSimpleString('"')
		if !ok {
			return token.Token{Kind: token.Illegal, Lexeme: "", Pos: pos}
		}
		return token.Token{Kind: token.Bytes, Lexeme: lit, Pos: pos}
	}

	// Fallback to normal lexing for other tokens
	return l.nextTokenNormal(pos)
}

func (l *Lexer) nextTokenNormal(pos token.Position) token.Token {
	ch := l.ch

	// Identifiers / keywords
	if isLetter(ch) {
		lit := l.readIdentifier()
		kind := token.LookupIdent(lit)
		return token.Token{
			Kind:   kind,
			Lexeme: lit,
			Pos:    pos,
		}
	}

	// Numbers
	if isDigit(ch) {
		lit := l.readNumber()
		kind := token.Int
		for _, r := range lit {
			if r == '.' || r == 'e' || r == 'E' {
				kind = token.Float
			break
		}
		}
		return token.Token{
			Kind:   kind,
			Lexeme: lit,
			Pos:    pos,
		}
	}

	// Single- and two-character tokens
	var kind token.Kind
	var lexeme string

	switch ch {
	case ';':
		kind = token.Semicolon
		lexeme = ";"
	case ',':
		kind = token.Comma
		lexeme = ","
	case '.':
		kind = token.Dot
		lexeme = "."
	case ':':
		kind = token.Colon
		lexeme = ":"
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			kind = token.OrOr
			lexeme = "||"
		} else {
			kind = token.Pipe
			lexeme = "|"
		}
	case '(':
		kind = token.LParen
		lexeme = "("
	case ')':
		kind = token.RParen
		lexeme = ")"
	case '{':
		kind = token.LBrace
		lexeme = "{"
	case '}':
		kind = token.RBrace
		lexeme = "}"
	case '[':
		kind = token.LBracket
		lexeme = "["
	case ']':
		kind = token.RBracket
		lexeme = "]"
	case '?':
		kind = token.Question
		lexeme = "?"
	case '+':
		kind = token.Plus
		lexeme = "+"
	case '-':
		kind = token.Minus
		lexeme = "-"
	case '*':
		kind = token.Star
		lexeme = "*"
	case '/':
		kind = token.Slash
		lexeme = "/"
	case '%':
		kind = token.Percent
		lexeme = "%"
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			kind = token.NotEq
			lexeme = "!="
		} else {
			kind = token.Bang
			lexeme = "!"
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			kind = token.AndAnd
			lexeme = "&&"
		} else {
			kind = token.Illegal
			lexeme = "&"
		}
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			kind = token.Eq
			lexeme = "=="
		} else {
			kind = token.Assign
			lexeme = "="
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			kind = token.LtEq
			lexeme = "<="
		} else {
			kind = token.Lt
			lexeme = "<"
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			kind = token.GtEq
			lexeme = ">="
		} else {
			kind = token.Gt
			lexeme = ">"
		}
	default:
		kind = token.Illegal
		lexeme = string(ch)
	}

	l.readChar()

	return token.Token{
		Kind:   kind,
		Lexeme: lexeme,
		Pos:    pos,
	}
}

func (l *Lexer) readSimpleString(delimiter rune) (string, bool) {
	startPos := token.Position{Line: l.line, Column: l.col}
	var sb []rune
	for {
		if l.ch == 0 || l.ch == '\n' {
			l.errorf(startPos, "unterminated string literal")
			return "", false
		}
		if l.ch == delimiter {
			l.readChar()
			return string(sb), true
		}
		if l.ch == '\\' {
			escPos := token.Position{Line: l.line, Column: l.col}
			l.readChar()
			r, ok := l.readEscape(escPos)
			if !ok {
				return "", false
			}
			sb = append(sb, r)
			l.readChar()
			continue
		}
		sb = append(sb, l.ch)
		l.readChar()
	}
}

func (l *Lexer) readEscape(pos token.Position) (rune, bool) {
	switch l.ch {
	case '\\':
		return '\\', true
	case '"':
		return '"', true
	case '\'':
		return '\'', true
	case 'n':
		return '\n', true
	case 't':
		return '\t', true
	case 'r':
		return '\r', true
	case '0':
		return 0, true
	case 'u':
		return l.readHexEscape(pos, 4)
	case 'x':
		return l.readHexEscape(pos, 2)
	default:
		l.errorf(pos, "invalid escape sequence")
		return 0, false
	}
}

func (l *Lexer) readHexEscape(pos token.Position, count int) (rune, bool) {
	var val rune
	for i := 0; i < count; i++ {
		l.readChar()
		ch := l.ch
		if ch == 0 {
			l.errorf(pos, "unterminated escape sequence")
			return 0, false
		}
		v, ok := hexValue(ch)
		if !ok {
			l.errorf(pos, "invalid hex escape")
			return 0, false
		}
		val = val*16 + v
	}
	return val, true
}

func hexValue(ch rune) (rune, bool) {
	switch {
	case ch >= '0' && ch <= '9':
		return ch - '0', true
	case ch >= 'a' && ch <= 'f':
		return ch - 'a' + 10, true
	case ch >= 'A' && ch <= 'F':
		return ch - 'A' + 10, true
	default:
		return 0, false
	}
}

func (l *Lexer) errorf(pos token.Position, msg string) {
	l.errors = append(l.errors, formatError(pos, msg))
}

func formatError(pos token.Position, msg string) string {
	return fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, msg)
}

func (l *Lexer) Errors() []string {
	return l.errors
}

func isLetter(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	if ch > utf8.RuneSelf {
		return false
	}
	return ch >= '0' && ch <= '9'
}
