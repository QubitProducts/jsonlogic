package jsonlogic

import (
	"fmt"
	"math"
	"reflect"
)

const (
	// null
	nullOp = ""

	//  var
	varOp         = "var"          // TODO - full path
	missingOp     = "missing"      // TODO
	missingSomeOp = "missing_some" // TODO

	// Logic
	ifOp            = "if"
	equalOp         = "==" // TODO - coercion
	equalThreeOp    = "==="
	notEqualOp      = "!="  // TODO
	notEqualThreeOp = "!==" // TODO
	notOp           = "!"   // TODO
	notTwoOp        = "!!"  // TODO
	orOp            = "or"  // TODO
	andOp           = "and"

	// Numeric
	greaterOp   = ">"   // TODO - non float
	greaterEqOp = ">="  // TODO
	lessOp      = "<"   // TODO - non float
	lessEqOpOp  = "<="  // TODO
	maxOp       = "max" // TODO
	minOp       = "min" // TODO

	// Array operations
	plusOp     = "+" // TODO
	minusOp    = "-" // TODO
	multiplyOp = "*" // TODO
	divideOp   = "/" // TODO
	moduloOp   = "%"

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
			if ok && len(index) == 0 {
				return data
			}
			v := DottedRef(data, indexVal)
			if v != nil {
				return v
			}
			return defaultVal

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

func nullf(data interface{}) interface{} {
	return nil
}

func buildMissingOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	switch {
	case len(args) == 0:
		return func(data interface{}) interface{} {
			return data
		}, nil
	}

	var termArgs []ClauseFunc
	for _, a := range args {
		termArg, err := buildArgFunc(a, ops)
		if err != nil {
			return nil, err
		}
		termArgs = append(termArgs, termArg)
	}

	return func(data interface{}) interface{} {
		resp := []interface{}{}
		for _, ta := range termArgs {
			lval := ta(data)
			v := DottedRef(data, lval)
			if v == nil {
				resp = append(resp, v)
			}
		}
		return resp
	}, nil
}

func buildIfOp3(args Arguments, ops OpsSet) (ClauseFunc, error) {
	var err error

	termArg, err := buildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	lArg, err := buildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	rArg := nullf
	if len(args) == 3 {
		if rArg, err = buildArgFunc(args[1], ops); err != nil {
			return nil, err
		}
	}

	return func(data interface{}) interface{} {
		termVal := termArg(data)
		lVal := lArg(data)
		rVal := rArg(data)
		if IsTrue(termVal) {
			return lVal
		}
		return rVal
	}, nil
}

func buildIfOpMulti(args Arguments, ops OpsSet) (ClauseFunc, error) {
	var termArgs []ClauseFunc
	for _, a := range args {
		termArg, err := buildArgFunc(a, ops)
		if err != nil {
			return nil, err
		}
		termArgs = append(termArgs, termArg)
	}

	return func(data interface{}) interface{} {
		last := 0
		for i := 0; i <= len(termArgs)/2; i++ {
			lval := termArgs[i*2](data)
			if IsTrue(lval) {
				rval := termArgs[i*2+1](data)
				return rval
			}
			last += 2
		}
		// got here, if there is a final term, it should
		// be return
		if last == len(termArgs) {
			return termArgs[len(termArgs)-1](data)
		}
		return nil
	}, nil
}

func buildIfOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	switch {
	case len(args) == 0:
		return nullf, nil
	case len(args) == 1:
		return func(data interface{}) interface{} {
			return data
		}, nil
	case len(args) <= 3:
		return buildIfOp3(args, ops)
	default:
		return buildIfOpMulti(args, ops)
	}

}

func buildAndOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) == 0 {
		return nullf, nil
	}

	var termArgs []ClauseFunc
	for _, ta := range args {
		termArg, err := buildArgFunc(ta, ops)
		if err != nil {
			return nil, err
		}
		termArgs = append(termArgs, termArg)
	}

	return func(data interface{}) interface{} {
		for _, t := range termArgs {
			v := t(data)
			if !IsTrue(v) {
				return false
			}
		}
		return true
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

func buildGreaterOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	switch {
	case len(args) == 0:
		return func(data interface{}) interface{} {
			return false
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

		lFloat, lisfloat := lVal.(float64)
		rFloat, risfloat := rVal.(float64)
		if lisfloat && risfloat {
			return lFloat > rFloat
		}

		panic("greater than disjoint types not implemented")
	}, nil
}

func buildBetweenExOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	lArg, err := buildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	mArg, err := buildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := buildArgFunc(args[2], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lVal := lArg(data)
		mVal := mArg(data)
		rVal := rArg(data)

		lFloat, lisfloat := lVal.(float64)
		mFloat, misfloat := mVal.(float64)
		rFloat, risfloat := rVal.(float64)
		if lisfloat && misfloat && risfloat {
			return lFloat < mFloat && mFloat < rFloat
		}

		panic("less than disjoint types not implemented")
	}, nil
}

func buildLessOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) < 2 {
		return func(data interface{}) interface{} {
			return false
		}, nil
	}
	if len(args) >= 3 {
		return buildBetweenExOp(args, ops)
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

		lFloat, lisfloat := lVal.(float64)
		rFloat, risfloat := rVal.(float64)
		if lisfloat && risfloat {
			return lFloat < rFloat
		}

		panic("less than disjoint types not implemented")
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

func buildModuloOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) < 2 {
		return func(data interface{}) interface{} {
			return nil
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

		lFloat, lisfloat := lVal.(float64)
		rFloat, risfloat := rVal.(float64)
		if lisfloat && risfloat {
			return math.Mod(lFloat, rFloat)
		}

		panic("modulo disjoint types not implemented")
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
	ifOp:         buildIfOp,
	andOp:        buildAndOp,
	equalOp:      buildEqualOp,
	equalThreeOp: buildEqualThreeOp,
	lessOp:       buildLessOp,
	greaterOp:    buildGreaterOp,
	moduloOp:     buildModuloOp,
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
