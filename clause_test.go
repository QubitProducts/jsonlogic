package jsonlogic

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClauseEval(t *testing.T) {
	type test struct {
		name       string
		rule       string
		marshalTo  string
		data       interface{}
		expect     interface{}
		parseErr   string
		compileErr string
	}

	tests := []test{
		{
			name:     "invalid-json",
			rule:     `{`,
			expect:   true,
			parseErr: "unexpected end of JSON input",
		},
		{ // invalid clause, but valid JSON, should return the thing passed in.
			name:      "invalid-clause",
			rule:      `"hello"`,
			marshalTo: `"hello"`,
			expect:    "hello",
		},
		{
			name:      "invalid-clause2",
			rule:      `{"var":["a"],"hola":"thing"}`,
			marshalTo: `{"hola":"thing","var":["a"]}`,
			expect: map[string]interface{}{
				"hola": "thing",
				"var":  []interface{}{"a"},
			},
		},
		{
			name:      "always",
			rule:      `true`,
			marshalTo: `true`,
			expect:    true,
		},
		{
			name:      "never",
			rule:      `false`,
			marshalTo: `false`,
			expect:    false,
		},
		{
			name:       "unknown-op",
			rule:       `{ "XXX" : [1, 1] }`,
			marshalTo:  `{"XXX":[1,1]}`,
			compileErr: "unrecognized operation XXX",
		},
		{
			name:      "simple",
			rule:      `{ "==" : [1, 1] }`,
			marshalTo: `{"==":[1,1]}`,
			expect:    true,
		},
		{
			name:      "simple-uncoerced",
			rule:      `{ "===" : [1, 1] }`,
			marshalTo: `{"===":[1,1]}`,
			expect:    true,
		},
		{
			name: "compound",
			rule: `{"and" : [
								 { ">" : [3,1] },
								 { "<" : [1,3] }
							 ] }`,
			marshalTo: `{"and":[{"\u003e":[3,1]},{"\u003c":[1,3]}]}`,
			expect:    true,
		},
		{
			name:      "data-driven",
			rule:      `{ "var" : ["a"] }`,
			marshalTo: `{"var":["a"]}`,
			data:      map[string]interface{}{"a": float64(1)},
			expect:    float64(1),
		},
		{
			name:      "data-driven-sugar",
			rule:      `{ "var" : "a" }`,
			marshalTo: `{"var":["a"]}`,
			data:      map[string]interface{}{"a": float64(1)},
			expect:    float64(1),
		},
		{
			name:      "data-driven-array",
			rule:      `{ "var" : 1 }`,
			marshalTo: `{"var":[1]}`,
			data:      []interface{}{"apple", "banana", "carrot"},
			expect:    "banana",
		},
		{
			name: "data-driven-mixed",
			rule: `{ "and" : [
									{"<" : [ { "var" : "temp" }, 110 ]},
									{"==" : [ { "var" : "pie.filling" }, "apple" ] }
								] }`,
			marshalTo: `{"and":[{"\u003c":[{"var":["temp"]},110]},{"==":[{"var":["pie.filling"]},"apple"]}]}`,
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
			marshalTo: `{"if":[{"==":[{"%":[{"var":["i"]},15]},0]},"fizzbuzz",{"==":[{"%":[{"var":["i"]},3]},0]},"fizz",{"==":[{"%":[{"var":["i"]},5]},0]},"buzz",{"var":["i"]}]}`,
			data:      20,
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
				} else {
					assert.NoErrorf(t, err, "unmarshal error")
				}

				marshalTo, err := json.Marshal(c)
				assert.NoErrorf(t, err, "marshal error")
				assert.Equalf(t, st.marshalTo, string(marshalTo), "re-marshaled clause")

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
