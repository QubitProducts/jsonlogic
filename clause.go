package jsonlogic

import (
	"encoding/json"
	"fmt"
	"log"
)

// Operator repsents a jsonlogic Operator.
type Operator struct {
	Name string
}

// Argument represnts any valid argument to a jsonlogic
// operator.
type Argument struct {
	Clause *Clause
	Value  interface{}
}

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

func (arg *Argument) UnmarshalJSON(bs []byte) error {
	c := Clause{}
	clauseErr := json.Unmarshal(bs, &c)
	if clauseErr == nil {
		*arg = Argument{
			Clause: &c,
			Value:  nil,
		}
		return nil
	}
	var v interface{}
	vErr := json.Unmarshal(bs, &v)
	if vErr == nil {
		*arg = Argument{
			Value: v,
		}
		log.Printf("arg %#v", *arg)
		return nil
	}

	return fmt.Errorf("could not parse argument, %w", vErr)
}

type Arguments []Argument

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

// Clause representes a jsonlogic clause.
type Clause struct {
	Operator  Operator
	Arguments Arguments
}

// UnmarshalJSON parses JSON data as a jsonlogic
// Clause.
func (c *Clause) UnmarshalJSON(bs []byte) error {
	raw := map[string]Arguments{}
	maperr := json.Unmarshal(bs, &raw)
	if maperr == nil {
		if len(raw) != 1 {
			return fmt.Errorf("too many keys for a clause, should be 1")
		}
		for k, v := range raw {
			*c = Clause{
				Operator: Operator{
					Name: k,
				},
				Arguments: v,
			}
		}
		return nil
	}

	var truthy bool
	trutherr := json.Unmarshal(bs, &truthy)
	if trutherr == nil {
		opName := "Never"
		if truthy {
			opName = "Always"
		}
		*c = Clause{
			Operator: Operator{
				Name: opName,
			},
		}
		return nil
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
func (c Clause) MarshalJSON() ([]byte, error) {
	switch c.Operator.Name {
	case "Always":
		return json.Marshal(true)
	case "Never":
		return json.Marshal(false)
	default:
		return json.Marshal(map[string]Arguments{
			c.Operator.Name: c.Arguments,
		})
	}
}

/*
// EvalJSON evluates this clause against the supplied
// data. The data is parsed as json, and raw json
// is returned.
func (c *Clause) CpompileJSON(data []byte) ([]byte, error) {
	panic("not implemented")
}

// Eval evaluates this clause against the provided data.
func (c *Clause) Compile(data interface{}) (interface{}, error) {
	panic("not implemented")
}
*/
