package jsonlogic

import (
	"fmt"
	"reflect"
)

const (
	// null
	nullOp = ""

	//  var
	varOp         = "var"
	missingOp     = "missing"      // TODO
	missingSomeOp = "missing_some" // TODO

	// Logic
	ifOp            = "if" // TODO
	equalOp         = "=="
	equalThreeOp    = "==="
	notEqualOp      = "!="  // TODO
	notEqualThreeOp = "!==" // TODO
	notOp           = "!"   // TODO
	notTwoOp        = "!!"  // TODO
	orOp            = "or"  // TODO
	andOp           = "and" // TODO

	// Numeric
	greater   = ">"   // TODO
	greaterEq = ">="  // TODO
	less      = "<"   // TODO
	lessEqOp  = "<="  // TODO
	maxOp     = "max" // TODO
	minOp     = "min" // TODO

	// Array operations
	plusOp     = "+" // TODO
	minusOp    = "-" // TODO
	multiplyOp = "*" // TODO
	divideOp   = "/" // TODO
	modOp      = "%" // TODO

	// String operations
	inOp     = "in"     // TODO
	catOp    = "cat"    // TODO
	substrOp = "substr" // TODO
)

type OpsSet map[string]func(args Arguments, ops OpsSet) (ClauseFunc, error)

func buildArgFunc(arg Argument, ops OpsSet) (ClauseFunc, error) {
	if arg.Clause == nil {
		return func(interface{}) interface{} {
			return arg.Value
		}, nil
	}
	return ops.Compile(arg.Clause)
}

func buildNullOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	return func(data interface{}) interface{} {
		return args[0].Value
	}, nil
}

func buildVarOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	var err error
	var indexArg ClauseFunc

	defaultArg := func(interface{}) interface{} {
		return nil
	}

	switch {
	case len(args) == 0:
		return func(data interface{}) interface{} {
			return data
		}, nil
	case len(args) >= 2:
		if defaultArg, err = buildArgFunc(args[1], ops); err != nil {
			return nil, err
		}
		fallthrough
	case len(args) >= 1:
		if indexArg, err = buildArgFunc(args[0], ops); err != nil {
			return nil, err
		}
	}

	return func(data interface{}) interface{} {
		indexVal := indexArg(data)
		defaultVal := defaultArg(data)

		switch data := data.(type) {
		case map[string]interface{}:
			index, ok := indexVal.(string)
			if !ok {
				return defaultArg
			}
			val, ok := data[index]
			if !ok {
				return defaultArg
			}
			return val

		case []interface{}:
			index, ok := indexVal.(float64)
			intindex := int(index)

			switch {
			case
				!ok,
				float64(intindex) != index,
				intindex < 0 || intindex >= len(data):

				return defaultVal

			default:
				return data[intindex]
			}
		default:
			return defaultVal
		}
	}, nil
}

func buildEqualOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	switch {
	case len(args) == 0:
		return func(data interface{}) interface{} {
			return true
		}, nil
	case len(args) == 1:
		return func(data interface{}) interface{} {
			return false
		}, nil
	}

	lArg, err := buildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := buildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lVal := lArg(data)
		rVal := rArg(data)

		return fmt.Sprintf("%v", lVal) == fmt.Sprintf("%v", rVal)
	}, nil
}

func buildEqualThreeOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	switch {
	case len(args) == 0:
		return func(data interface{}) interface{} {
			return true
		}, nil
	case len(args) == 1:
		return func(data interface{}) interface{} {
			return false
		}, nil
	}

	lArg, err := buildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := buildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lVal := lArg(data)
		rVal := rArg(data)

		return reflect.DeepEqual(lVal, rVal)
	}, nil
}

func (ops OpsSet) Compile(c *Clause) (ClauseFunc, error) {
	bf, ok := ops[c.Operator.Name]
	if !ok {
		return nil, fmt.Errorf("unrecognized operation %s", c.Operator.Name)
	}
	return bf(c.Arguments, ops)
}

var DefaultOps = OpsSet{
	nullOp:       buildNullOp,
	varOp:        buildVarOp,
	equalOp:      buildEqualOp,
	equalThreeOp: buildEqualThreeOp,
}

// ClauseFunc takes input data, returns a result which
// could be any valid json type. jsonlogic seems to
// prefer returning null to returning any specific errors.
type ClauseFunc func(data interface{}) interface{}

var ops = map[string]func(args Arguments) ClauseFunc{
	nullOp: func(args Arguments) ClauseFunc {
		return func(data interface{}) interface{} {
			return args[0].Value
		}
	},
}

// Compile builds a ClauseFunc that will execute
// the provided rule against the data.
func Compile(c *Clause) (ClauseFunc, error) {
	return DefaultOps.Compile(c)
}
