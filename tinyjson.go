// tinyjson is a minimalistic JSON tokenizer/parser for producing small tinygo
// binaries. It assumes a valid JSON input already available as a []byte.
package tinyjson

import (
	"strconv"
	"strings"
	"unsafe"
)

// Token represents a valid JSON token, nil when EOF, non-empty otherwise
type Token []byte

// Kind is generally the first character of a token, 0 for EOF, '9' for numbers.
type Kind byte

const (
	EOF         Kind = 0
	StartObject Kind = '{'
	EndObject   Kind = '}'
	StartArray  Kind = '['
	EndArray    Kind = ']'
	String      Kind = '"'
	Number      Kind = '9'
	True        Kind = 't'
	False       Kind = 'f'
	Null        Kind = 'n'
	Colon       Kind = ':'
	Comma       Kind = ','
)

var kindByByte = [256]Kind{
	'{': StartObject,
	'}': EndObject,
	'[': StartArray,
	']': EndArray,
	'"': String,
	'-': Number,
	'.': Number,
	'0': Number,
	'1': Number,
	'2': Number,
	'3': Number,
	'4': Number,
	'5': Number,
	'6': Number,
	'7': Number,
	'8': Number,
	'9': Number,
	't': True,
	'f': False,
	'n': Null,
	':': Colon,
	',': Comma,
}

var (
	trueToken  = Token("true")
	falseToken = Token("false")
	nullToken  = Token("null")
)

// Raw returns the source JSON string of the token.
func (t Token) Raw() string {
	if t == nil {
		return ""
	}
	return unsafe.String(&t[0], len(t))
}

// Kind returns the kind of the token, generally its first character, 0 for EOF, '9' for numbers.
func (t Token) Kind() Kind {
	if t == nil {
		return EOF
	}
	return kindByByte[t[0]]
}

// Scalar returns a Go scalar value corresponding to this token, panics if impossible.
func (t Token) Scalar() any {
	switch t.Kind() {
	case EOF, Null:
		return nil
	case Number:
		return t.Float()
	case String:
		return unquoteString(t)
	case True:
		return true
	case False:
		return false
	default:
		panic("unexpected JSON: " + t.Raw())
	}
}

// Str returns an unquoted Go string value corresponding to this token, with
// escape seqeuences handled. Returns an empty string for null or EOF, and
// the original JSON strings for false, true and numbers. Panics otherwise.
func (t Token) Str() string {
	switch t.Kind() {
	case EOF, Null:
		return ""
	case String:
		return unquoteString(t)
	case True, False, Number:
		return t.Raw()
	default:
		panic("unexpected JSON: " + t.Raw())
	}
}

// Int returns an int value corresponding to this token, panics if impossible.
func (t Token) Int() int {
	if t.Kind() == Number {
		if v, err := strconv.ParseInt(t.Raw(), 10, 0); err == nil {
			return int(v)
		}
	}
	panic("unexpected JSON: " + t.Raw())
}

// Int returns an int64 value corresponding to this token, panics if impossible.
func (t Token) Int64() int64 {
	if t.Kind() == Number {
		if v, err := strconv.ParseInt(t.Raw(), 10, 0); err == nil {
			return v
		}
	}
	panic("unexpected JSON: " + t.Raw())
}

// Int returns an uint64 value corresponding to this token, panics if impossible.
func (t Token) Uint64() uint64 {
	if t.Kind() == Number {
		if v, err := strconv.ParseUint(t.Raw(), 10, 0); err == nil {
			return v
		}
	}
	panic("unexpected JSON: " + t.Raw())
}

// Int returns a float64 value corresponding to this token, panics if impossible.
func (t Token) Float() float64 {
	if t.Kind() == Number {
		if v, err := strconv.ParseFloat(t.Raw(), 64); err == nil {
			return v
		}
	}
	panic("unexpected JSON: " + t.Raw())
}

// Int returns true or false value corresponding to this token, panics if impossible.
func (t Token) Bool() bool {
	switch t.Kind() {
	case True:
		return true
	case False:
		return false
	default:
		panic("unexpected JSON: " + t.Raw())
	}
}

func peekNextTokenKind(data []byte) (kind Kind, remainder []byte) {
	start := 0
	n := len(data)
	for {
		if start == n {
			return EOF, nil
		}
		if !isWhitespace(data[start]) {
			break
		}
		start++
	}

	return kindByByte[data[start]], data[start:]
}

func nextToken(data []byte) (token Token, remainder []byte) {
	start := 0
	n := len(data)
	for {
		if start == n {
			return nil, nil
		}
		if !isWhitespace(data[start]) {
			break
		}
		start++
	}

	c := data[start]
	switch c {
	case '"':
		return scanString(data[start:])
	case 't':
		return trueToken, data[start+4:]
	case 'f':
		return falseToken, data[start+5:]
	case 'n':
		return nullToken, data[start+4:]
	default:
		k := kindByByte[c]
		if k == Number {
			return scanNumber(data[start:])
		} else if k != 0 {
			return Token(data[start : start+1]), data[start+1:]
		} else {
			panic("invalid JSON")
		}
	}
}

func scanString(data []byte) (Token, []byte) {
	n := len(data)
	for i := 1; i < n; i++ {
		switch data[i] {
		case '"':
			return Token(data[:i+1]), data[i+1:]
		case '\\':
			i++
		}
	}
	panic("invalid JSON")
}

func scanNumber(data []byte) (Token, []byte) {
	for i, c := range data {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', 'e', 'E', '+', '-':
			continue
		default:
			return Token(data[:i]), data[i:]
		}
	}
	return Token(data), nil
}

func unquoteString(s []byte) string {
	n := len(s)
	s = s[1 : n-1]
	n -= 2
	if !hasEscape(s) {
		return unsafe.String(&s[0], len(s))
	}
	var buf strings.Builder
	buf.Grow(len(s))
	for i := 0; i < n; i++ {
		c := s[i]
		if c != '\\' {
			buf.WriteByte(c)
		} else {
			i++
			if i == n {
				panic("invalid JSON")
			}
			c = s[i]
			switch c {
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'u':
				if i+4 >= n {
					panic("invalid JSON")
				}
				u, err := strconv.ParseUint(unsafe.String(&s[i+1], 4), 16, 32)
				if err != nil {
					panic("invalid JSON")
				}
				buf.WriteRune(rune(u))
				i += 4
			default:
				buf.WriteByte(c)
			}
		}
	}
	return buf.String()
}

func hasEscape(s []byte) bool {
	for _, c := range s {
		if c == '\\' {
			return true
		}
	}
	return false
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t'
}

// Raw is a []byte encoding of an unparsed portion of JSON document.
type Raw []byte

// Next returns the next token in the JSON data.
func (raw *Raw) Next() Token {
	token, remainder := nextToken(*raw)
	*raw = Raw(remainder)
	return token
}

// Peek returns what Next().Kind() would return without advancing past the next
// token. (Peek does advance past leading whitespace to run in amortized O(1),
// assuming all tokens will be eventually scanned or skipped over.)
func (raw *Raw) Peek() Kind {
	kind, remainder := peekNextTokenKind(*raw)
	*raw = Raw(remainder)
	return kind
}

// StartObject ensures the next token is an open curly brace and returns
// the first object key. Returns nil for an empty object.
//
// Idiomatic usage:
//
//	for key := raw.StartObject(); key != nil; key = raw.ContinueObject() {
//		switch key.Str() { ... }
//	}
func (raw *Raw) StartObject() Token {
	if t := raw.Next(); t.Kind() != StartObject {
		panic("unexpected JSON: " + t.Raw())
	}
	return raw.ContinueObject()
}

// ContinueObject returns the next object key, skipping over a comma if any.
// Returns nil if no more keys are present.
func (raw *Raw) ContinueObject() Token {
again:
	t := raw.Next()
	switch t.Kind() {
	case Comma:
		goto again
	case String:
		colon := raw.Next()
		if colon.Kind() != Colon {
			panic("invalid JSON")
		}
		return t
	case EndObject:
		return nil
	default:
		// log.Printf("t = >>>%s<<<, raw = >>>%s<<<", t, *raw)
		panic("invalid JSON")
	}
}

// StartArray ensures the next token is an open square bracket. Follow up with
// a call to ContinueArray to iterate over the array elements:
//
//	for raw.StartArray(); raw.ContinueArray(); {
//		// process the next value here via .Next(), .Skip(), .Str(), .Int(), etc.
//	}
func (raw *Raw) StartArray() {
	if t := raw.Next(); t.Kind() != StartArray {
		panic("unexpected JSON: " + t.Raw())
	}
}
func (raw *Raw) ContinueArray() bool {
again:
	switch raw.Peek() {
	case Comma:
		raw.Next()
		goto again
	case EndArray:
		raw.Next()
		return false
	case EOF:
		panic("invalid JSON")
	default:
		return true
	}
}

// Null skips 'null' token and returns true if the next token is null,
// returns false without advancing the parser otherwise.
func (raw *Raw) Null() bool {
	if raw.Peek() == Null {
		raw.Next()
		return true
	}
	return false
}

// Str returns .Next().Str(), i.e. parses and returns the next string value or panics if impossible; see [Token.Str] for details.
func (raw *Raw) Str() string { return raw.Next().Str() }

// Int returns .Next().Int(), i.e. parses and returns the next JSON number as an int value or panics if impossible; see [Token.Int].
func (raw *Raw) Int() int { return raw.Next().Int() }

// Int64 returns .Next().Int64(), i.e. parses and returns the next JSON number as an int64 value or panics if impossible; see [Token.Int64].
func (raw *Raw) Int64() int64 { return raw.Next().Int64() }

// Uint64 returns .Next().Uint64(), i.e. parses and returns the next JSON number as an uint64 value or panics if impossible; see [Token.Uint64].
func (raw *Raw) Uint64() uint64 { return raw.Next().Uint64() }

// Float returns .Next().Float(), i.e. parses and returns the next JSON number as a float64 value or panics if impossible; see [Token.Float].
func (raw *Raw) Float() float64 { return raw.Next().Float() }

// Bool returns .Next().Bool(), i.e. parses and returns the next JSON boolean value, or panics if impossible; see [Token.Bool].
func (raw *Raw) Bool() bool { return raw.Next().Bool() }

// Value returns the next JSON value; arrays are returned as []any, objects as map[string]any.
func (raw *Raw) Value() any {
	t := raw.Next()
	switch t.Kind() {
	case EOF:
		return nil
	case StartObject:
		result := make(map[string]any)
		for key := raw.ContinueObject(); key != nil; key = raw.ContinueObject() {
			result[key.Str()] = raw.Value()
		}
		return result
	case StartArray:
		var result []any
		for raw.ContinueArray() {
			result = append(result, raw.Value())
		}
		return result
	case String, Number, True, False, Null:
		return t.Scalar()
	default:
		panic("invalid JSON")
	}
}

// Skip advances past the next JSON value (including skipping over objects and arrays).
func (raw *Raw) Skip() {
	t := raw.Next()
	switch t.Kind() {
	case StartObject:
		for key := raw.ContinueObject(); key != nil; key = raw.ContinueObject() {
			raw.Skip()
		}
	case StartArray:
		for raw.ContinueArray() {
			raw.Skip()
		}
	case String, Number, True, False, Null:
		break
	default:
		panic("invalid JSON")
	}
}

// EnsureEOF panics if more JSON data is found.
func (raw *Raw) EnsureEOF() {
	if raw.Peek() != EOF {
		panic("invalid JSON")
	}
}
