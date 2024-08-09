Minimalistic JSON tokenizer/parser for smaller tinygo WASM binaries
===================================================================

[![Go reference](https://pkg.go.dev/badge/github.com/andreyvit/tinyjson.svg)](https://pkg.go.dev/github.com/andreyvit/tinyjson) ![Zero dependencies](https://img.shields.io/badge/deps-zero-brightgreen) ![Zero magic](https://img.shields.io/badge/magic-none-brightgreen) ![400 LOC](https://img.shields.io/badge/size-400%20LOC-green) ![100% coverage](https://img.shields.io/badge/coverage-100%25-green) [![Go Report Card](https://goreportcard.com/badge/github.com/andreyvit/tinyjson)](https://goreportcard.com/report/github.com/andreyvit/tinyjson)

Use this in place of `encoding/json` or `github.com/json-iterator/tinygo` to avoid adding a lot of unnecessary code into your Web Assembly binaries.

* No reflection
* Zero allocations (unless you use Value to decode `[]any` or `map[string]any`)
* The input is `[]byte`, avoiding all the complexities with partial inputs
* Token type is just `[]byte` slice of the input
* All returned strings are also slices of the input (via `unsafe.String`), except for strings that require processing of escape sequences
* Assume valid JSON on input, panics if not
* Blazing-fast


Usage
-----

Install: `go get github.com/andreyvit/tinyjson@latest`

```go
raw := tinyjson.Raw(`{"name":"test","bars":[{"title":"one","count":1},{"title":"two","count":2}]}`)
var foo Foo
foo.DecodeJSON(&raw)
raw.EnsureEOF()
```

using these example types:

```go
type Foo struct {
	Name string
	Bars []*Bar
}

type Bar struct {
	Title string
	Count int
}

func (foo *Foo) DecodeJSON(raw *tinyjson.Raw) {
	for key := raw.StartObject(); key != nil; key = raw.ContinueObject() {
		switch key.Str() {
		case "name":
			foo.Name = raw.Str()
		case "bars":
			for raw.StartArray(); raw.ContinueArray(); {
				bar := new(Bar)
				bar.DecodeJSON(raw)
				foo.Bars = append(foo.Bars, bar)
			}
		default:
			panic("invalid key " + key.Str())
		}
	}
}

func (bar *Bar) DecodeJSON(raw *tinyjson.Raw) {
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
```


Contributing
------------

“We include what we think Go team would choose to include if this was a standard library package.”

We accept contributions that:

* add better documentation and examples;
* fix bugs;
* decrease the size of the compiled binary.

We recommend [modd](https://github.com/cortesi/modd) (`go install github.com/cortesi/modd/cmd/modd@latest`) for continuous testing during development.

Maintain 100% coverage. It's not often the right choice, but it is for this library.


BSD 2-Clause license
--------------------

Copyright © 2024, Andrey Tarantsov.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
