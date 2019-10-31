package jsonlogic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTrue(t *testing.T) {
	type test struct {
		name        string
		in          interface{}
		expect      bool
		expectPanic bool
	}

	tests := []test{
		{
			name:   "nil",
			in:     nil,
			expect: false,
		},
		{
			name:   "0",
			in:     float64(0),
			expect: false,
		},
		{
			name:   "1",
			in:     float64(1),
			expect: true,
		},
		{
			name:   "-1",
			in:     float64(-1),
			expect: true,
		},

		{
			name:   "empty-array",
			in:     []interface{}{},
			expect: false,
		},
		{
			name:   "non-empty-array",
			in:     []interface{}{1, "hello"},
			expect: true,
		},
		{
			name:   "empty-string",
			in:     "",
			expect: false,
		},
		{
			name:   "non-empty-string",
			in:     "hello",
			expect: true,
		},
		{
			name:   "zero-string",
			in:     "0",
			expect: true,
		},
		{
			name:   "empty-dict",
			in:     map[string]interface{}{},
			expect: true,
		},
	}

	for _, st := range tests {
		t.Run(st.name, func(t *testing.T) {
			tf := func() {
				result := IsTrue(st.in)
				assert.Equal(t, st.expect, result)
			}
			if !st.expectPanic {
				assert.NotPanics(t, tf)
			} else {
				assert.Panics(t, tf)
			}
		})
	}
}
