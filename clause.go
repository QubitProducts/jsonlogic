package jsonlogic

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

// Clause representes a jsonlogic clause.
type Clause struct {
	Operator  Operator
	Arguments []Argument
}

// UnmarshalJSON parses JSON data as a jsonlogic
// Clause.
func (c *Clause) UnmarshalJSON(bs []byte) error {
	return nil
}

// MarshalJSON marshals a jsonlogic Clause into
// pure JSON.
func (c *Clause) MarshalJSON() ([]byte, error) {
	return nil, nil
}

// EvalJSON evluates this clause against the supplied
// data. The data is parsed as json, and raw json
// is returned.
func (c *Clause) EvalJSON(data []byte) ([]byte, error) {
	return nil, nil
}

// Eval evaluates this clause against the provided data.
func (c *Clause) Eval(data interface{}) (interface{}, error) {
	return nil, nil
}
