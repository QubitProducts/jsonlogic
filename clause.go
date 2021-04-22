package jsonlogic

import (
	"encoding/json"
	"fmt"
)

// Operator represents a jsonlogic Operator.
type Operator struct {
	Name string
}

// Argument represents any valid argument to a jsonlogic
// operator.
type Argument struct {
	Clause *Clause
	Value  interface{}
}

// MarshalJSON implements json.Marshaler. It enforces
// rending of clause arguments as an array (even if there was
// just one non array argument in the original clause when
// unmarshaled).
func (a Argument) MarshalJSON() ([]byte, error) {
	switch {
	case a.Clause != nil && a.Value != nil:
		return nil, fmt.Errorf("an argument should only have a clause, or a value, not both")
	case a.Clause != nil:
		return json.Marshal(a.Clause)
	default:
		return json.Marshal(a.Value)
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Argument) UnmarshalJSON(bs []byte) error {
	c := Clause{}
	clauseErr := json.Unmarshal(bs, &c)
	if clauseErr == nil {
		*a = Argument{
			Clause: &c,
			Value:  nil,
		}
		return nil
	}
	var v interface{}
	vErr := json.Unmarshal(bs, &v)
	if vErr == nil {
		*a = Argument{
			Value: v,
		}
		return nil
	}

	return fmt.Errorf("could not parse argument, %w", vErr)
}

// Arguments represents the list of arguments to a jsonlogic
// Clause.
type Arguments []Argument

// UnmarshalJSON implements json.Unmarshaler.
func (args *Arguments) UnmarshalJSON(bs []byte) error {
	slice := []Argument{}
	sliceErr := json.Unmarshal(bs, &slice)
	if sliceErr == nil {
		*args = slice
		return nil
	}
	arg := Argument{}
	if oneErr := json.Unmarshal(bs, &arg); oneErr == nil {
		*args = []Argument{arg}
		return nil
	}
	return fmt.Errorf("could not parse arguments")
}

// Clause represents a JsonLogic clause.
type Clause struct {
	Operator  Operator
	Arguments Arguments
}

// UnmarshalJSON parses JSON data as a JsonLogic
// Clause.
func (c *Clause) UnmarshalJSON(bs []byte) error {
	clause := map[string]Arguments{}
	err := json.Unmarshal(bs, &clause)
	if err == nil && len(clause) == 1 {
		for k, v := range clause {
			*c = Clause{
				Operator: Operator{
					Name: k,
				},
				Arguments: v,
			}
			return nil
		}
	}

	var raw interface{}
	err = json.Unmarshal(bs, &raw)
	if err != nil {
		return err
	}
	// this is a bit subtle, we want to differentiate instances of the empty
	// slice, this forces a new slice header.
	if rawslice, ok := raw.([]interface{}); ok && len(rawslice) == 0 {
		raw = make([]interface{}, 0, 1)
	}
	*c = Clause{
		Arguments: []Argument{{
			Value: raw,
		}},
	}
	return nil
}

// MarshalJSON implements json.Marshaler. It enforces
// rending of clause arguments as an array (even if there was
// just one non array argument in the original clause when
// unmarshaled).
func (c Clause) MarshalJSON() ([]byte, error) {
	switch c.Operator.Name {
	case "":
		return json.Marshal(c.Arguments[0].Value)
	default:
		return json.Marshal(map[string]Arguments{
			c.Operator.Name: c.Arguments,
		})
	}
}
