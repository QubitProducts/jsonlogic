package jsonlogic

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClauseEval(t *testing.T) {
	type test struct {
		name      string
		rule      string
		marshalTo string
		data      interface{}
		expect    interface{}
		expErr    string
	}

	tests := []test{
		{
			name:   "invalid-json",
			rule:   `{`,
			expect: true,
			expErr: "unexpected end of JSON input",
		},
		{
			name:   "invalid-clause",
			rule:   `"hello"`,
			expect: true,
			expErr: "jsonlogic parse: json: cannot unmarshal string into Go value of type map[string]jsonlogic.Arguments",
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
			name:      "simple",
			rule:      `{ "==" : [1, 1] }`,
			marshalTo: `{"==":[1,1]}`,
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
			data:      map[string]float64{"a": 1},
			expect:    1,
		},
		{
			name:      "data-driven-sugar",
			rule:      `{ "var" : "a" }`,
			marshalTo: `{"var":["a"]}`,
			data:      map[string]float64{"a": 1},
			expect:    1,
		},
		{
			name:      "data-driven-array",
			rule:      `{ "var" : 1 }`,
			marshalTo: `{"var":[1]}`,
			data:      []string{"apple", "banana", "carrot"},
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
				"pie":  map[string]string{"filling": "apple"},
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
				if st.expErr != "" {
					assert.EqualError(t, err, st.expErr)
					return
				} else {
					assert.NoError(t, err)
				}

				marshalTo, err := json.Marshal(c)
				assert.NoError(t, err)
				assert.Equal(t, st.marshalTo, string(marshalTo))

				/*
					v, err := c.Eval(st.data)
					assert.EqualError(t, err, st.expErr)
					if err != nil {
						return
					}

					assert.Equal(t, v, st.expect)
				*/
			})
		})
	}
}
