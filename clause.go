package jsonlogic

import (
	"encoding/json"
	"fmt"
)

// Operator repsents a jsonlogic Operator.
type Operator struct {
	Name string
}

// Argument represnts any valid argument to a jsonlogic
// operator.
type Argument struct {
	Clause
	string
	float64
}

type Arguments []Argument

func (args *Arguments) UnmarshalJSON(bs []byte) error {
	panic("not implemented")
}

// Clause representes a jsonlogic clause.
type Clause struct {
	Operator  Operator
	Arguments Arguments
}

// UnmarshalJSON parses JSON data as a jsonlogic
// Clause.
func (c *Clause) UnmarshalJSON(bs []byte) error {
	args := map[string]Arguments{}
	maperr := json.Unmarshal(bs, &args)
	if maperr == nil {
		panic("map not implemented")
	}

	var truthy bool
	trutherr := json.Unmarshal(bs, &truthy)
	if trutherr == nil {
		panic("bool not implemented")
	}

	switch {
	case maperr != nil:
		return fmt.Errorf("jsonlogic parse: %w", maperr)
	case trutherr != nil:
		return fmt.Errorf("jsonlogic parse: %w", trutherr)
	default:
		return fmt.Errorf("jsonlogic parse: could not parse as valid clause")
	}
}

// MarshalJSON marshals a jsonlogic Clause into
// pure JSON.
func (c *Clause) MarshalJSON() ([]byte, error) {
	panic("not implemented")
}

// EvalJSON evluates this clause against the supplied
// data. The data is parsed as json, and raw json
// is returned.
func (c *Clause) EvalJSON(data []byte) ([]byte, error) {
	panic("not implemented")
}

// Eval evaluates this clause against the provided data.
func (c *Clause) Eval(data interface{}) (interface{}, error) {
	panic("not implemented")
}
