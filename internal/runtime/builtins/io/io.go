package io

// IO is the minimal interface needed by builtin IO functions (e.g. print, input).
type IO interface {
	Println(string)
	ReadLine() (string, error)
}
