package jsonlogic

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClauseEval(t *testing.T) {
	type test struct {
		name   string
		rule   string
		data   interface{}
		expect interface{}
		expErr assert.ErrorAssertionFunc
	}

	tests := []test{
		{
			name:   "invalid",
			rule:   `x`,
			expect: true,
			expErr: assert.Error,
		},
		{
			name:   "simple",
			rule:   `{ "==" : [1, 1] }`,
			expect: true,
			expErr: assert.NoError,
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
			data:   map[string]float64{"a": 1},
			expect: 1,
		},
		{
			name:   "data-driven-sugar",
			rule:   `{ "var" : "a" }`,
			data:   map[string]float64{"a": 1},
			expect: 1,
		},
		{
			name:   "data-driven-array",
			rule:   `{ "var" : 1 }`,
			data:   []string{"apple", "banana", "carrot"},
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
				"pie":  map[string]string{"filling": "apple"},
			},
			expect: true,
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
			data: 20,
		},
	}

	for _, st := range tests {
		t.Run(st.name, func(t *testing.T) {
			var c Clause
			err := json.Unmarshal([]byte(st.rule), &c)
			assert.EqualError(t, err, st.expErr)
			if err != nil {
				return
			}

			v, err := c.Eval(st.data)
			assert.EqualError(t, err, st.expErr)
			if err != nil {
				return
			}

			assert.Equal(t, v, st.expect)
		})
	}
}
