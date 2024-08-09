package tinyjson_test

import (
	"fmt"

	"github.com/andreyvit/tinyjson"
)

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

func Example() {
	raw := tinyjson.Raw(`{"name":"test","bars":[{"title":"one","count":1},{"title":"two","count":2}]}`)
	var foo Foo
	foo.DecodeJSON(&raw)
	raw.EnsureEOF()

	fmt.Println(foo.Name)
	for _, bar := range foo.Bars {
		fmt.Println(bar.Title, bar.Count)
	}

	// Output: test
	// one 1
	// two 2
}
