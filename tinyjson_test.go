package tinyjson

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestNext(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{`object`, `{"name":"John Doe","age":30,"city":"New York"}`, `{ "name" : "John Doe" , "age" : 30 , "city" : "New York" }`},
		{`object with whitespace`, "\n{\n\"name\": \"John Doe\",\t\"age\" :30\n\t\t,\t\"city\":\r\n\"New York\"\n\n\n}", `{ "name" : "John Doe" , "age" : 30 , "city" : "New York" }`},
		{`array`, `[1,2,3]`, `[ 1 , 2 , 3 ]`},
		{`boolean`, `true`, `true`},
		{`string`, `"hello"`, `"hello"`},
		{`string with escapes`, `"escaped\":\\\/\b\f\n\r\t\u263A"`, `"escaped\":\\\/\b\f\n\r\t\u263A"`},
		{`float`, `5.78`, `5.78`},
		{`negative number`, `-23`, `-23`},
		{`scientific notation`, `6.022e23`, `6.022e23`},
		{`empty string`, `""`, `""`},
		{`escaped backslash`, `"\\"`, `"\\"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := strings.Join(allTokens(test.input), " ")
			if actual != test.expected {
				t.Errorf("** Tokens(%v) = %s, wanted %s", test.input, actual, test.expected)
			}
		})
	}
}

func TestSkip(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{`true`, `true 42`, "42"},
		{`number`, `3.141 42`, "42"},
		{`string`, `"foo bar" 42`, "42"},
		{`EOF`, `true`, ""},
		{`object`, `{"name":"John Doe","age":30,"city":"New York"} 42`, "42"},
		{`nested structure`, `{"name":"John Doe","items":[1,2,3, {"subkey": 123, "test": []}], "city":{"name": "New York"}} 42`, "42"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw := Raw(test.input)
			raw.Skip()
			actual := raw.Next().Raw()
			if actual != test.expected {
				t.Errorf("** Tokens(%v) = %s, wanted %s", test.input, actual, test.expected)
			}
		})
	}
}

func TestStr(t *testing.T) {
	tests := []struct {
		name     string
		token    Token
		expected string
	}{
		{`string token`, Token(`"hello"`), "hello"},
		{`escape double quote`, Token(`"\""`), `"`},
		{`escape backslash`, Token(`"\\"`), `\`},
		{`escape forward slash`, Token(`"\/"`), "/"},
		{`escape backspace`, Token(`"\b"`), "\b"},
		{`escape form feed`, Token(`"\f"`), "\f"},
		{`escape newline`, Token(`"\n"`), "\n"},
		{`escape carriage return`, Token(`"\r"`), "\r"},
		{`escape tab`, Token(`"\t"`), "\t"},
		{`multiple escapes`, Token(`"\n\t\f"`), "\n\t\f"},
		{`escapes with other characters`, Token(`"foo\nbar\tboz\t\\fubar\ffizboz"`), "foo\nbar\tboz\t\\fubar\ffizboz"},
		{`unicode escape`, Token(`"\u263A"`), "☺"},
		{`true`, Token("true"), "true"},
		{`false`, Token("false"), "false"},
		{`null`, Token("null"), ""},
		{`EOF`, Token(nil), ""},
		{`int`, Token("42"), "42"},
		{`float`, Token("3.141"), "3.141"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.token.Str()
			if actual != test.expected {
				t.Errorf("** Token.String(%s) = %s, wanted %s", test.token, actual, test.expected)
			}
		})
	}
}
func TestInt64(t *testing.T) {
	tests := []struct {
		name     string
		token    Token
		expected int64
	}{
		{`positive number`, Token("123"), 123},
		{`negative number`, Token("-456"), -456},
		{`zero`, Token("0"), 0},
		{`int64 max`, Token("9223372036854775807"), 9223372036854775807},
		{`int64 min`, Token("-9223372036854775808"), -9223372036854775808},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.token.Int64()
			if actual != test.expected {
				t.Errorf("** Token.Int64(%v) = %d, wanted %d", test.token, actual, test.expected)
			}
		})
	}
}

func TestUint64(t *testing.T) {
	tests := []struct {
		name     string
		token    Token
		expected uint64
	}{
		{`zero`, Token("0"), 0},
		{`positive number`, Token("123"), 123},
		{`max uint64`, Token("18446744073709551615"), 18446744073709551615},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.token.Uint64()
			if actual != test.expected {
				t.Errorf("** Token.Uint64(%v) = %d, wanted %d", test.token, actual, test.expected)
			}
		})
	}
}

func TestFloat(t *testing.T) {
	tests := []struct {
		name     string
		token    Token
		expected float64
	}{
		{`positive float`, Token("3.14"), 3.14},
		{`negative float`, Token("-2.718"), -2.718},
		{`zero`, Token("0.0"), 0.0},
		{`scientific notation`, Token("6.022e23"), 6.022e23},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.token.Float()
			if actual != test.expected {
				t.Errorf("** Token.Float64(%v) = %g, wanted %g", test.token, actual, test.expected)
			}
		})
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		name     string
		token    Token
		expected bool
	}{
		{`true`, Token("true"), true},
		{`false`, Token("false"), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.token.Bool()
			if actual != test.expected {
				t.Errorf("** Token.Bool(%v) = %t, wanted %t", test.token, actual, test.expected)
			}
		})
	}
}

func TestValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{`eof`, ``, nil},
		{`null`, `null`, nil},
		{`true`, `true`, true},
		{`false`, `false`, false},
		{`integer`, `123`, 123.0},
		{`float`, `3.14`, 3.14},
		{`scientific notation`, `6.022e23`, 6.022e23},
		{`string`, `"Hello, World!"`, "Hello, World!"},
		{`array`, `[1, "two", 3.0, true, null]`, []any{1.0, "two", 3.0, true, nil}},
		{`object`, `{"name":"John", "age":30, "city":"New York"}`, map[string]any{"name": "John", "age": 30.0, "city": "New York"}},
		{`nested object`, `{"person":{"name":"John", "age":30}, "city":"New York"}`, map[string]any{"person": map[string]any{"name": "John", "age": 30.0}, "city": "New York"}},
		{`nested array`, `[1, [2, 3], 4]`, []any{1.0, []any{2.0, 3.0}, 4.0}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw := Raw(test.input)
			actual := raw.Value()
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("** Raw.Value() = %v, wanted %v", actual, test.expected)
			}
		})
	}
}

func TestNull(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{`eof`, ``, false},
		{`null`, `null`, true},
		{`true`, `true`, false},
		{`false`, `false`, false},
		{`integer`, `123`, false},
		{`float`, `3.14`, false},
		{`string`, `"foo"`, false},
		{`array`, `[1, "two", 3.0, true, null]`, false},
		{`object`, `{"name":"John", "age":30, "city": null}`, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw := Raw(test.input)
			actual := raw.Null()
			if actual != test.expected {
				t.Errorf("** Raw.Null() = %v, wanted %v", actual, test.expected)
			}
		})
	}
}

func TestPanics(t *testing.T) {
	tests := []struct {
		name     string
		f        func()
		expected string
	}{
		{`bare word`, func() { raw(`xxx`).Next() }, "invalid JSON"},
		{`unclosed string`, func() { raw(`"xxx`).Next() }, "invalid JSON"},
		{`unterminated escape`, func() { raw(`"xxx\`).Next() }, "invalid JSON"},
		{`unfinished unicode escape`, func() { raw(`"xxx\u12"`).Str() }, "invalid JSON"},
		{`unfinished unicode escape in unquote`, func() { unquoteString([]byte(`"xxx\"`)) }, "invalid JSON"},
		{`invalid unicode escape`, func() { raw(`"xxx\u123Z"`).Str() }, "invalid JSON"},

		{`array cannot Str`, func() { raw(`[]`).Str() }, "unexpected JSON: ["},
		{`array cannot Int`, func() { raw(`[]`).Int() }, "unexpected JSON: ["},
		{`array cannot Scalar`, func() { raw(`[]`).Next().Scalar() }, "unexpected JSON: ["},

		{`object cannot Str`, func() { raw(`{}`).Str() }, "unexpected JSON: {"},
		{`comma cannot Str`, func() { raw(`,`).Str() }, "unexpected JSON: ,"},
		{`comma cannot Value`, func() { raw(`,`).Value() }, "invalid JSON"},
		{`comma cannot Skip`, func() { raw(`,`).Skip() }, "invalid JSON"},
		{`comma cannot EnsureEOF`, func() { raw(`,`).EnsureEOF() }, "invalid JSON"},

		{`null cannot Int`, func() { raw(`null`).Int() }, "unexpected JSON: null"},
		{`null cannot Float`, func() { raw(`null`).Float() }, "unexpected JSON: null"},
		{`null cannot Bool`, func() { raw(`null`).Bool() }, "unexpected JSON: null"},

		{`string cannot Int`, func() { raw(`"42"`).Int() }, `unexpected JSON: "42"`},
		{`string cannot Int64`, func() { raw(`"42"`).Int64() }, `unexpected JSON: "42"`},
		{`string cannot Uint64`, func() { raw(`"42"`).Uint64() }, `unexpected JSON: "42"`},
		{`string cannot StartObject`, func() { raw(`"42"`).StartObject() }, `unexpected JSON: "42"`},
		{`string cannot StartArray`, func() { raw(`"42"`).StartArray() }, `unexpected JSON: "42"`},

		{`unclosed object`, func() { raw(`{"xxx": 42`).Value() }, "invalid JSON"},
		{`unclosed array`, func() { raw(`["xxx"`).Value() }, "invalid JSON"},
		{`no colon in object`, func() { raw(`{"a" 1}`).Value() }, "invalid JSON"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ensurePanic(t, test.f, test.expected)
		})
	}
}

func BenchmarkDecode(t *testing.B) {
	orig := Raw(`{"title":"one","count":1}`)
	for i := 0; i < t.N; i++ {
		raw := orig
		var bar Bar
		bar.DecodeJSON(&raw)
		raw.EnsureEOF()
	}
}

func raw(data string) *Raw {
	raw := Raw(data)
	return &raw
}

func allTokens(data string) []string {
	var tokens []string
	raw := Raw(data)
	for {
		kind := raw.Peek()
		token := raw.Next()
		if kind != token.Kind() {
			panic("kind mismatch")
		}
		if token == nil {
			break
		}
		tokens = append(tokens, token.Raw())
	}
	return tokens
}

func ensurePanic(t testing.TB, f func(), e string) {
	actual := capturePanic(f)
	if actual == nil {
		t.Helper()
		t.Errorf("** succeeded, wanted to panic with: %v", e)
	} else if a := fmt.Sprint(actual); a != e {
		t.Helper()
		t.Errorf("** paniced with: %v, wanted: %v", a, e)
	}
}

func capturePanic(f func()) (panicValue any) {
	defer func() { panicValue = recover() }()
	f()
	return
}

type Bar struct {
	Title string
	Count int
}

func (bar *Bar) DecodeJSON(raw *Raw) {
	for key := raw.StartObject(); key != nil; key = raw.ContinueObject() {
		switch key.Str() {
		case "title":
			bar.Title = raw.Str()
		case "count":
			bar.Count = raw.Int()
		default:
			panic("invalid key " + key.Str())
		}
	}
}
