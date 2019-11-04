package jsonlogic

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClauseMarshal(t *testing.T) {
	type test struct {
		name      string
		rule      string
		marshalTo string
		parseErr  string
	}

	tests := []test{
		{
			name:     "invalid-json",
			rule:     `{`,
			parseErr: "unexpected end of JSON input",
		},
		{ // invalid clause, but valid JSON, should return the thing passed in.
			name:      "invalid-clause",
			rule:      `"hello"`,
			marshalTo: `"hello"`,
		},
		{
			name:      "invalid-clause2",
			rule:      `{"var":["a"],"hola":"thing"}`,
			marshalTo: `{"hola":"thing","var":["a"]}`,
		},
		{
			name:      "always",
			rule:      `true`,
			marshalTo: `true`,
		},
		{
			name:      "never",
			rule:      `false`,
			marshalTo: `false`,
		},
		{
			name:      "simple",
			rule:      `{ "==" : [1, 1] }`,
			marshalTo: `{"==":[1,1]}`,
		},
		{
			name: "compound",
			rule: `{"and" : [
								 { ">" : [3,1] },
								 { "<" : [1,3] }
							 ] }`,
			marshalTo: `{"and":[{"\u003e":[3,1]},{"\u003c":[1,3]}]}`,
		},
		{
			name:      "data-driven",
			rule:      `{ "var" : ["a"] }`,
			marshalTo: `{"var":["a"]}`,
		},
		{
			name:      "data-driven-sugar",
			rule:      `{ "var" : "a" }`,
			marshalTo: `{"var":["a"]}`,
		},
		{
			name:      "data-driven-sugar",
			rule:      `{ "var" : "a" }`,
			marshalTo: `{"var":["a"]}`,
		},
	}

	for _, st := range tests {
		t.Run(st.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				var c Clause
				err := json.Unmarshal([]byte(st.rule), &c)
				if st.parseErr != "" {
					assert.EqualErrorf(t, err, st.parseErr, "unmarshal error")
					return
				}
				assert.NoErrorf(t, err, "unmarshal error")

				marshalTo, err := json.Marshal(c)
				assert.NoErrorf(t, err, "marshal error")
				assert.Equalf(t, st.marshalTo, string(marshalTo), "re-marshaled clause")
			})
		})
	}
}

func TestClauseEval(t *testing.T) {
	type test struct {
		name       string
		rule       string
		data       interface{}
		expect     interface{}
		compileErr string
	}

	tests := []test{
		{ // invalid clause, but valid JSON, should return the thing passed in.
			name:   "invalid-clause",
			rule:   `"hello"`,
			expect: "hello",
		},
		{
			name: "invalid-clause2",
			rule: `{"var":["a"],"hola":"thing"}`,
			expect: map[string]interface{}{
				"hola": "thing",
				"var":  []interface{}{"a"},
			},
		},
		{
			name:   "always",
			rule:   `true`,
			expect: true,
		},
		{
			name:   "never",
			rule:   `false`,
			expect: false,
		},
		{
			name:       "unknown-op",
			rule:       `{ "XXX" : [1, 1] }`,
			compileErr: "unrecognized operation XXX",
		},
		{
			name:   "simple",
			rule:   `{ "==" : [1, 1] }`,
			expect: true,
		},
		{
			name:   "equal-no-arg",
			rule:   `{ "==" : [] }`,
			expect: true,
		},
		{
			name:   "equal-one-arg",
			rule:   `{ "==" : [1] }`,
			expect: false,
		},
		{
			name:   "equal-coerced",
			rule:   `{ "==" : [1, "1"] }`,
			expect: true,
		},
		{
			name:   "equal-coerced-bool",
			rule:   `{ "==" : [0, false] }`,
			expect: true,
		},
		{
			name:   "equal-strict",
			rule:   `{"===" : [1, 1]}`,
			expect: true,
		},
		{
			name:   "equal-strict-false",
			rule:   `{"===" : [1, "1"]}`,
			expect: false,
		},
		{
			name:   "notequal-coerced-bool",
			rule:   `{ "!=" : [1, 2] }`,
			expect: true,
		},
		{
			name:   "notequal-coerced-bool-false",
			rule:   `{ "!=" : [1, "1"] }`,
			expect: false,
		},
		{
			name:   "notequal-strict",
			rule:   `{ "!==" : [1, 2] }`,
			expect: true,
		},
		{
			name:   "notequal-strict",
			rule:   `{ "!==" : [1, "1"] }`,
			expect: true,
		},
		{
			name:   "notequal-strict-uncoerced",
			rule:   `{ "!==" : [1, "1"] }`,
			expect: true,
		},
		{
			name:   "notequal-strict-false",
			rule:   `{ "!==" : [1, 1] }`,
			expect: false,
		},
		{
			name:   "negate-false",
			rule:   `{ "!" : [false]}`,
			expect: true,
		},
		{
			name:   "negate-true",
			rule:   `{ "!" : [true]}`,
			expect: false,
		},
		{
			name:   "negate-noargs",
			rule:   `{ "!" : []}`,
			expect: true,
		},
		{
			name:   "negate-one-arg",
			rule:   `{ "!" : false}`,
			expect: true,
		},
		{
			name:   "double-negate",
			rule:   `{"!!": [ [] ] }`,
			expect: false,
		},
		{
			name:   "double-negate-true",
			rule:   `{"!!": [ "0" ] }`,
			expect: true,
		},
		{
			name:   "double-negate-true",
			rule:   `{"!!": [ "0" ] }`,
			expect: true,
		},
		{
			name:   "or",
			rule:   `{"or": [ true,false ] }`,
			expect: true,
		},
		{
			name:   "or",
			rule:   `{"or": [ false,true ] }`,
			expect: true,
		},
		{
			name:   "or-last-arg",
			rule:   `{"or": [ false,"a" ] }`,
			expect: "a",
		},
		{
			name:   "or-multi-last-arg",
			rule:   `{"or": [ false,0,"a" ] }`,
			expect: "a",
		},
		{
			name:   "or-multi-last-false",
			rule:   `{"or": [ false,0,[] ] }`,
			expect: []interface{}{},
		},
		{
			name:   "and",
			rule:   `{"and": [ true,true ] }`,
			expect: true,
		},
		{
			name:   "and-false",
			rule:   `{"and": [ true,false ] }`,
			expect: false,
		},
		{
			name:   "and-multi",
			rule:   `{"and":[true,"a",3]}`,
			expect: 3.0,
		},
		{
			name:   "and-multi-false",
			rule:   `{"and":[true,"",3]}`,
			expect: "",
		},
		{
			name:   "greater",
			rule:   `{">":[2,1]}`,
			expect: true,
		},
		{
			name:   "greater-or-equal",
			rule:   `{">=":[1,1]}`,
			expect: true,
		},
		{
			name:   "less",
			rule:   `{"<":[1,2]}`,
			expect: true,
		},
		{
			name:   "less-or-equal",
			rule:   `{"<=": [1,1]}`,
			expect: true,
		},
		{
			name:   "greater-false",
			rule:   `{">":[2,2]}`,
			expect: false,
		},
		{
			name:   "greater-or-equal-false",
			rule:   `{">=":[1,2]}`,
			expect: false,
		},
		{
			name:   "less-false",
			rule:   `{"<":[1,1]}`,
			expect: false,
		},
		{
			name:   "less-or-equal-false",
			rule:   `{"<=": [1,0]}`,
			expect: false,
		},
		{
			name:   "between",
			rule:   `{"<":[1,2,3]}`,
			expect: true,
		},
		{
			name:   "between-low",
			rule:   `{"<":[1,1,3]}`,
			expect: false,
		},
		{
			name:   "betwee-high",
			rule:   `{"<":[1,4,3]}`,
			expect: false,
		},
		{
			name:   "between-inc",
			rule:   `{"<=":[1,2,3]}`,
			expect: true,
		},
		{
			name:   "between-inc-low-equal",
			rule:   `{"<=":[1,1,3]}`,
			expect: true,
		},
		{
			name:   "betwee-inc-high",
			rule:   `{"<=":[1,4,3]}`,
			expect: false,
		},
		{
			name:   "less-with-data",
			rule:   `{ "<": [0, {"var":"temp"}, 100]}`,
			data:   map[string]interface{}{"temp": 37.0},
			expect: true,
		},
		{
			name:   "min",
			rule:   `{"min":[1,2,3]}`,
			expect: 1.0,
		},
		{
			name:   "max",
			rule:   `{"max":[1,2,3]}`,
			expect: 3.0,
		},
		{
			name:   "max-nil",
			rule:   `{"max":[]}`,
			expect: nil,
		},
		{
			name:   "plus",
			rule:   `{"+":[4,2]}`,
			expect: 6.0,
		},
		{
			name:   "plus-single",
			rule:   `{"+":[4]}`,
			expect: 4.0,
		},
		{
			name:   "plus-none",
			rule:   `{"+":[]}`,
			expect: 0.0,
		},
		{
			name:   "minus",
			rule:   `{"-":[4,2]}`,
			expect: 2.0,
		},
		{
			name:   "minus-none",
			rule:   `{"-":[]}`,
			expect: nil,
		},
		{
			name:   "multiply",
			rule:   `{"*":[4,2]}`,
			expect: 8.0,
		},
		{
			name:   "multiply-one",
			rule:   `{"*":[4]}`,
			expect: 4.0,
		},
		{
			name:   "multiply-one",
			rule:   `{"*":[]}`,
			expect: nil, // this one actually errors on jsonlogic
		},
		{
			name:   "divide",
			rule:   `{"/":[4,2]}`,
			expect: 2.0,
		},
		{
			name:   "divide-one",
			rule:   `{"/":[4]}`,
			expect: nil,
		},
		{
			name:   "divide-multi", // jsonlogic seems to ignore multiple args here
			rule:   `{"/":[4,2,2]}`,
			expect: 2.0,
		},
		{
			name:   "plus-multi",
			rule:   `{"+":[2,2,2,2,2]}`,
			expect: 10.0,
		},
		{
			name:   "multiply-multi",
			rule:   `{"*":[2,2,2,2,2]}`,
			expect: 32.0,
		},
		{
			name:   "unary-minus",
			rule:   `{"-":[2]}`,
			expect: -2.0,
		},
		{
			name:   "unary-minus-flip",
			rule:   `{"-":[-2]}`,
			expect: 2.0,
		},
		{
			name:   "unary-plus-string",
			rule:   `{"+":["3.14"]}`,
			expect: 3.14,
		},
		{
			name: "compound",
			rule: `{"and" : [
								 { ">" : [3,1] },
								 { "<" : [1,3] }
							 ] }`,
			expect: true,
		},
		{
			name:   "data-driven",
			rule:   `{ "var" : ["a"] }`,
			data:   map[string]interface{}{"a": float64(1)},
			expect: float64(1),
		},
		{
			name:   "data-driven-default",
			rule:   `{ "var" : ["b",26] }`,
			data:   map[string]interface{}{"a": float64(1)},
			expect: float64(26),
		},
		{
			name:   "data-driven-sugar",
			rule:   `{ "var" : "a" }`,
			data:   map[string]interface{}{"a": float64(1)},
			expect: float64(1),
		},
		{
			name:   "data-driven-empty-key",
			rule:   `{ "var" : "" }`,
			data:   map[string]interface{}{"a": float64(1)},
			expect: map[string]interface{}{"a": float64(1)},
		},
		{
			name:   "data-driven-empty-array-key",
			rule:   `{ "var" : [""] }`,
			data:   2.0,
			expect: 2.0,
		},
		{
			name:   "data-driven-array",
			rule:   `{ "var" : 1 }`,
			data:   []interface{}{"apple", "banana", "carrot"},
			expect: "banana",
		},
		{
			name: "data-driven-mixed",
			rule: `{ "and" : [
									{"<" : [ { "var" : "temp" }, 110 ]},
									{"==" : [ { "var" : "pie.filling" }, "apple" ] }
								] }`,
			data: map[string]interface{}{
				"temp": 100.0,
				"pie":  map[string]interface{}{"filling": "apple"},
			},
			expect: true,
		},
		{
			name: "fizzbuzz",
			rule: `{ "if": [
								{"==": [ { "%": [ { "var": "i" }, 15 ] }, 0]},
								"fizzbuzz",

								{"==": [ { "%": [ { "var": "i" }, 3 ] }, 0]},
								"fizz",

								{"==": [ { "%": [ { "var": "i" }, 5 ] }, 0]},
								"buzz",

								{ "var": "i" }
							 ]}`,
			expect: "buzz",
			data:   map[string]interface{}{"i": float64(20)},
		},
		{
			name:   "missing",
			rule:   `{"missing":["a", "b"]}`,
			data:   map[string]interface{}{"a": "apple", "c": "carrot"},
			expect: []interface{}{"b"},
		},
		{
			name:   "missing-none-missing",
			rule:   `{"missing":["a", "b"]}`,
			data:   map[string]interface{}{"a": "apple", "b": "carrot"},
			expect: []interface{}{},
		},
		{
			name: "missing-with-if",
			rule: `{"if":[
							{"missing":["a", "b"]},
							"Not enough fruit",
							"OK to proceed"
						 ]}`,
			data:   map[string]interface{}{"a": "apple", "b": "carrot"},
			expect: "OK to proceed",
		},
		{
			name:   "missing_some-ok",
			rule:   `{"missing_some":[1, ["a", "b", "c"]]}`,
			data:   map[string]interface{}{"a": "apple"},
			expect: []interface{}{},
		},
		{
			name:   "missing_some-toomany",
			rule:   `{"missing_some":[2, ["a", "b", "c"]]}`,
			data:   map[string]interface{}{"a": "apple"},
			expect: []interface{}{"b", "c"},
		},
		{
			name: "missing_some-complex",
			rule: `{"if" :[
							{"merge": [
								{"missing":["first_name", "last_name"]},
						    {"missing_some":[1, ["cell_phone", "home_phone"] ]}
							]},
							"We require first name, last name, and one phone number.",
							"OK to proceed"
						 ]}`,
			data:   map[string]interface{}{"first_name": "Bruce", "last_name": "Wayne"},
			expect: "We require first name, last name, and one phone number.",
		},
		{
			name:   "if-true",
			rule:   `{"if" : [ true, "yes", "no" ]}`,
			data:   nil,
			expect: "yes",
		},
		{
			name:   "if-false",
			rule:   `{"if" : [ false, "yes", "no" ]}`,
			data:   nil,
			expect: "no",
		},
		{
			name: "if-multi",
			rule: `{"if" : [
							{"<": [{"var":"temp"}, 0] }, "freezing",
							{"<": [{"var":"temp"}, 100] }, "liquid",
							"gas"
						]}`,
			data:   map[string]interface{}{"temp": 55.0},
			expect: "liquid",
		},
		{
			name: "map",
			rule: `{"map":[
							{"var":"integers"},
							{"*":[{"var":""},2]}
						 ]}`,
			data: map[string]interface{}{
				"integers": []interface{}{1.0, 2.0, 3.0, 4.0, 5.0},
			},
			expect: []interface{}{2.0, 4.0, 6.0, 8.0, 10.0},
		},
		{
			name: "filter",
			rule: `{"filter":[
							{"var":"integers"},
							{"%":[{"var":""},2]}
						 ]}`,
			data: map[string]interface{}{
				"integers": []interface{}{1.0, 2.0, 3.0, 4.0, 5.0},
			},
			expect: []interface{}{1.0, 3.0, 5.0},
		},
		{
			name: "reduce",
			rule: `{"reduce":[
							{"var":"integers"},
					    {"+":[{"var":"current"}, {"var":"accumulator"}]},
							0
						 ]}`,
			data: map[string]interface{}{
				"integers": []interface{}{1.0, 2.0, 3.0, 4.0, 5.0},
			},
			expect: 15.0,
		},
		{
			name:   "all",
			rule:   `{"some" : [ [-1,0,1], {">":[{"var":""}, 0]} ]}`,
			expect: true,
		},
		{
			name:   "all-fail",
			rule:   `{"all" : [ [-1,0,1], {"==":[{"var":""}, 0]} ]}`,
			expect: false,
		},
		{
			name:   "some",
			rule:   `{"some" : [ [-1,0,1], {">":[{"var":""}, 0]} ]}`,
			expect: true,
		},
		{
			name:   "some-fail",
			rule:   `{"some" : [ [-1,0,1], {">":[{"var":""}, 2]} ]}`,
			expect: false,
		},
		{
			name:   "none",
			rule:   `{"none" : [ [-3,-2,-1], {">":[{"var":""}, 0]} ]}`,
			expect: true,
		},
		{
			name:   "none-fail",
			rule:   `{"none" : [ [-3,-2,0], {"==":[{"var":""}, 0]} ]}`,
			expect: false,
		},
		{
			name:   "all-empty",
			rule:   `{"all" : [ [], {">":[{"var":""}, 0]} ]}`,
			expect: false,
		},
		{
			name:   "some-empty",
			rule:   `{"some" : [ [], {">":[{"var":""}, 0]} ]}`,
			expect: false,
		},
		{
			name:   "none-empty",
			rule:   `{"none" : [ [], {">":[{"var":""}, 0]} ]}`,
			expect: true,
		},
		{
			name:   "merge-empty",
			rule:   `{"merge" : [ ]}`,
			data:   nil,
			expect: []interface{}{},
		},
		{
			name:   "merge-regulat",
			rule:   `{"merge":[ [1, 2], [3,4]]}`,
			data:   nil,
			expect: []interface{}{1.0, 2.0, 3.0, 4.0},
		},
		{
			name:   "merge-coerce",
			rule:   `{"merge":[ 1, 2, [3,4]]}`,
			data:   nil,
			expect: []interface{}{1.0, 2.0, 3.0, 4.0},
		},
		{
			name:   "merge-coerce",
			rule:   `{"merge":[ 1, 2, [3,[4,5]]]}`,
			data:   nil,
			expect: []interface{}{1.0, 2.0, 3.0, []interface{}{4.0, 5.0}},
		},
		{
			name:   "in-array",
			rule:   `{"in":[true,"true"]}`,
			data:   nil,
			expect: true,
		},
		{
			name:   "in-string",
			rule:   `{"in":["Spring", "Springfield"]}`,
			data:   nil,
			expect: true,
		},
		{
			name:   "in-string-coerced",
			rule:   `{"in":[1, "hello1"]}`,
			data:   nil,
			expect: true,
		},
		{
			name:   "in-string-coerced-false",
			rule:   `{"in":[2, "hello1"]}`,
			data:   nil,
			expect: false,
		},
		{
			name:   "cat",
			rule:   `{"cat": ["I love", " pie"]}`,
			data:   nil,
			expect: "I love pie",
		},
		{
			name:   "cat-coerved",
			rule:   `{"cat": ["I love", " pie",false,1]}`,
			data:   nil,
			expect: "I love piefalse1",
		},
		{
			name:   "substr-nil",
			rule:   `{"substr": []}`,
			data:   nil,
			expect: "undefined",
		},
		{
			name:   "substr",
			rule:   `{"substr": ["jsonlogic", 4]}`,
			data:   nil,
			expect: "logic",
		},
		{
			name:   "substr",
			rule:   `{"substr": ["jsonlogic", 20]}`,
			data:   nil,
			expect: "",
		},
		{
			name:   "substr-from-end",
			rule:   `{"substr": ["jsonlogic", -5]}`,
			data:   nil,
			expect: "logic",
		},
		{
			name:   "substr-from-end-toolong",
			rule:   `{"substr": ["jsonlogic", -20]}`,
			data:   nil,
			expect: "jsonlogic",
		},
		{
			name:   "substr-from-end-shorter",
			rule:   `{"substr": ["jsonlogic", -2]}`,
			data:   nil,
			expect: "ic",
		},
		{
			name:   "substr-limit",
			rule:   `{"substr": ["jsonlogic", 1,3]}`,
			data:   nil,
			expect: "son",
		},
		{
			name:   "substr-limit-from-end",
			rule:   `{"substr": ["jsonlogic", 4,-2]}`,
			data:   nil,
			expect: "log",
		},
	}

	for _, st := range tests {
		t.Run(st.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				var c Clause
				err := json.Unmarshal([]byte(st.rule), &c)
				assert.NoErrorf(t, err, "unmarshal error")

				cf, err := Compile(&c)
				if st.compileErr != "" {
					assert.EqualErrorf(t, err, st.compileErr, "compile error")
					if err != nil {
						return
					}
				} else {
					assert.NoErrorf(t, err, "compile error")
				}

				v := cf(st.data)
				assert.Equalf(t, st.expect, v, "response data")
			})
		})
	}
}

func BenchmarkFizzBuzz(b *testing.B) {
	b.ReportAllocs()
	fizzbuzz := `{ "if": [
								{"==": [ { "%": [ { "var": "i" }, 15 ] }, 0]},
								"fizzbuzz",

								{"==": [ { "%": [ { "var": "i" }, 3 ] }, 0]},
								"fizz",

								{"==": [ { "%": [ { "var": "i" }, 5 ] }, 0]},
								"buzz",

								{ "var": "i" }
							 ]}`
	data := map[string]interface{}{"i": float64(20)}

	var c Clause
	err := json.Unmarshal([]byte(fizzbuzz), &c)
	if err != nil {
		b.Fatalf("unmarshal failed, %v", err)
	}

	cf, err := Compile(&c)
	if err != nil {
		b.Fatalf("compile failed, %v", err)
	}
	b.ResetTimer()
	for i := b.N; i >= 0; i-- {
		cf(data)
	}
}
