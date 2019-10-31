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
	*c = Clause{
		Arguments: []Argument{{
			Value: raw,
		}},
	}
	return nil
}

// MarshalJSON marshals a jsonlogic Clause into
// pure JSON.
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

// ClauseFunc takes input data, returns a result which
// could be any valid json type. jsonlogic seems to
// prefer returning null to returning any specific errors.
type ClauseFunc func(data interface{}) interface{}

// Compile builds a ClauseFunc that will execute
// the provided rule against the data.
func (c *Clause) Compile() (ClauseFunc, error) {
	panic("not implemented")
}
