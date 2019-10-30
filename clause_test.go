package jsonlogic

import "testing"

func TestClauseEval(t *testing.T) {
	type test struct {
		name   string
		rule   string
		data   interface{}
		expect interface{}
		expErr string
	}

	tests := []test{
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
		},
	}

	for _, st := range tests {
		t.Run(st.name, func(t *testing.T) {
		})
	}
}
