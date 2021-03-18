package jsonlogic

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

const (
	// null
	nullOp = ""

	//  var
	varOp         = "var"
	missingOp     = "missing"
	missingSomeOp = "missing_some"

	// Logic
	ifOp            = "if"
	equalOp         = "==" // TODO - coercion
	equalThreeOp    = "==="
	notEqualOp      = "!="
	notEqualThreeOp = "!=="
	negateOp        = "!"
	doubleNegateOp  = "!!"
	orOp            = "or"
	andOp           = "and"
	ternaryOp       = "?:"

	// Numeric
	greaterOp   = ">"  // TODO - non float
	greaterEqOp = ">=" // TODO
	lessOp      = "<"  // TODO - non float
	lessEqOp    = "<=" // TODO
	maxOp       = "max"
	minOp       = "min"

	plusOp     = "+"
	minusOp    = "-"
	multiplyOp = "*"
	divideOp   = "/"
	moduloOp   = "%"

	// Array operations
	mapOp    = "map"
	reduceOp = "reduce"
	filterOp = "filter"
	allOp    = "all"
	noneOp   = "none"
	someOp   = "some"
	mergeOp  = "merge"

	// String operations
	inOp     = "in"
	catOp    = "cat"
	substrOp = "substr"
)

func identityf(data interface{}) interface{} {
	return data
}

func nullf(data interface{}) interface{} {
	return nil
}

func emptySlice(interface{}) interface{} {
	return []interface{}{}
}

func truef(data interface{}) interface{} {
	return true
}

func falsef(data interface{}) interface{} {
	return false
}

// OpsSet operation names to a function that can build an instance of that
// operation.
type OpsSet map[string]func(args Arguments, ops OpsSet) (ClauseFunc, error)

func BuildArgFunc(arg Argument, ops OpsSet) (ClauseFunc, error) {
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

	defaultArg := nullf

	switch {
	case len(args) == 0:
		return identityf, nil
	case len(args) >= 2:
		if defaultArg, err = BuildArgFunc(args[1], ops); err != nil {
			return nil, err
		}
		fallthrough
	case len(args) >= 1:
		if indexArg, err = BuildArgFunc(args[0], ops); err != nil {
			return nil, err
		}
	}

	return func(data interface{}) interface{} {
		indexVal := indexArg(data)
		defaultVal := defaultArg(data)

		// if the index is an empty string, we don't care about
		// the type and return the entire thing.
		indexstr, ok := indexVal.(string)
		if ok && indexstr == "" {
			return data
		}

		// otherise, we assume this is an indexable type.
		switch data := data.(type) {
		case map[string]interface{}:
			v := DottedRef(data, indexVal)
			if v != nil {
				return v
			}
		case []interface{}:
			v := DottedRef(data, indexVal)
			if v != nil {
				return v
			}
		}
		return defaultVal
	}, nil
}

func buildMissingOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	switch {
	case len(args) == 0:
		return emptySlice, nil
	}

	var termArgs []ClauseFunc
	for _, a := range args {
		termArg, err := BuildArgFunc(a, ops)
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
				resp = append(resp, lval)
			}
		}
		return resp
	}, nil
}

func buildMissingSomeOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) <= 1 {
		return emptySlice, nil
	}

	requiredArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	termsArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		resp := []interface{}{}
		required := requiredArg(data)
		requiredfloat, ok := required.(float64)
		if !ok {
			return resp
		}

		terms := termsArg(data)
		termsslice, ok := terms.([]interface{})
		if !ok {
			return resp
		}

		found := float64(0)
		for _, ta := range termsslice {
			v := DottedRef(data, ta)
			if v != nil {
				found++
				continue
			}
			resp = append(resp, ta)
		}
		if found >= requiredfloat {
			return []interface{}{}
		}

		return resp
	}, nil
}

func buildIfOp3(args Arguments, ops OpsSet) (ClauseFunc, error) {
	var err error

	termArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	lArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	rArg := nullf
	if len(args) == 3 {
		if rArg, err = BuildArgFunc(args[2], ops); err != nil {
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
		termArg, err := BuildArgFunc(a, ops)
		if err != nil {
			return nil, err
		}
		termArgs = append(termArgs, termArg)
	}

	return func(data interface{}) interface{} {
		last := 0
		for i := 0; i < len(termArgs)/2; i++ {
			lval := termArgs[i*2](data)
			if IsTrue(lval) {
				rval := termArgs[i*2+1](data)
				return rval
			}
			last += 2
		}

		// got here, if there is a final term, it should
		// be return
		if last != len(termArgs) {
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
		return BuildArgFunc(args[0], ops)
	case len(args) <= 3:
		return buildIfOp3(args, ops)
	default:
		return buildIfOpMulti(args, ops)
	}
}

func buildTernaryOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	var err error

	termArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	lArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	rArg := nullf
	if len(args) == 3 {
		if rArg, err = BuildArgFunc(args[2], ops); err != nil {
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

func buildAndOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) == 0 {
		return nullf, nil
	}

	var termArgs []ClauseFunc
	for _, ta := range args {
		termArg, err := BuildArgFunc(ta, ops)
		if err != nil {
			return nil, err
		}
		termArgs = append(termArgs, termArg)
	}

	return func(data interface{}) interface{} {
		var lastArg interface{}
		for _, t := range termArgs {
			lastArg = t(data)
			if !IsTrue(lastArg) {
				return lastArg
			}
		}
		return lastArg
	}, nil
}

func buildOrOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) == 0 {
		return nullf, nil
	}

	var termArgs []ClauseFunc
	for _, ta := range args {
		termArg, err := BuildArgFunc(ta, ops)
		if err != nil {
			return nil, err
		}
		termArgs = append(termArgs, termArg)
	}

	return func(data interface{}) interface{} {
		var lastArg interface{}
		for _, t := range termArgs {
			lastArg = t(data)
			if IsTrue(lastArg) {
				return lastArg
			}
		}
		return lastArg
	}, nil
}

func buildEqualOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	switch {
	case len(args) == 0:
		return truef, nil
	case len(args) == 1:
		return falsef, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lVal := lArg(data)
		rVal := rArg(data)

		return IsSoftEqual(lVal, rVal)
	}, nil
}

func buildNotEqualOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	eqf, err := buildEqualOp(args, ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		if eqres, ok := eqf(data).(bool); ok {
			return !eqres
		}
		return false
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

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lVal := toNumber(lArg(data))
		rVal := toNumber(rArg(data))

		return lVal > rVal
	}, nil
}

func buildGreaterEqualOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
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

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lVal := toNumber(lArg(data))
		rVal := toNumber(rArg(data))

		return lVal >= rVal
	}, nil
}

func buildBetweenExOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	mArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := BuildArgFunc(args[2], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lVal := toNumber(lArg(data))
		mVal := toNumber(mArg(data))
		rVal := toNumber(rArg(data))

		return lVal < mVal && mVal < rVal
	}, nil
}

func buildBetweenIncOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	mArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := BuildArgFunc(args[2], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lVal := toNumber(lArg(data))
		mVal := toNumber(mArg(data))
		rVal := toNumber(rArg(data))

		return lVal <= mVal && mVal <= rVal
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

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lVal := toNumber(lArg(data))
		rVal := toNumber(rArg(data))

		return lVal < rVal
	}, nil
}

func buildLessEqualOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) < 2 {
		return func(data interface{}) interface{} {
			return false
		}, nil
	}
	if len(args) >= 3 {
		return buildBetweenIncOp(args, ops)
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lVal := toNumber(lArg(data))
		rVal := toNumber(rArg(data))

		return lVal <= rVal
	}, nil
}

func buildMaxOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	switch {
	case len(args) == 0:
		return nullf, nil
	}

	var termArgs []ClauseFunc
	for _, a := range args {
		termArg, err := BuildArgFunc(a, ops)
		if err != nil {
			return nil, err
		}
		termArgs = append(termArgs, termArg)
	}

	return func(data interface{}) interface{} {
		resp := math.Inf(-1)
		for _, ta := range termArgs {
			item := toNumber(ta(data))
			if math.IsNaN(item) {
				return item
			}
			if item > resp {
				resp = item
			}
		}
		return resp
	}, nil
}

func buildMinOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	switch {
	case len(args) == 0:
		return nullf, nil
	}

	var termArgs []ClauseFunc
	for _, a := range args {
		termArg, err := BuildArgFunc(a, ops)
		if err != nil {
			return nil, err
		}
		termArgs = append(termArgs, termArg)
	}

	return func(data interface{}) interface{} {
		resp := math.Inf(1)
		for _, ta := range termArgs {
			item := toNumber(ta(data))
			if math.IsNaN(item) {
				return item
			}
			if item < resp {
				resp = item
			}
		}
		return resp
	}, nil
}

func buildEqualThreeOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	switch {
	case len(args) == 0:
		return truef, nil
	case len(args) == 1:
		return falsef, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lVal := lArg(data)
		rVal := rArg(data)

		return IsEqual(lVal, rVal)
	}, nil
}

func buildNotEqualThreeOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	eqf, err := buildEqualThreeOp(args, ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		if eqres, ok := eqf(data).(bool); ok {
			return !eqres
		}
		return false
	}, nil
}

func buildNegateOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) == 0 {
		return truef, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		return !IsTrue(lArg(data))
	}, nil
}

func buildDoubleNegateOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) == 0 {
		return falsef, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		return IsTrue(lArg(data))
	}, nil
}

func buildPlusOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	var termArgs []ClauseFunc
	for _, a := range args {
		termArg, err := BuildArgFunc(a, ops)
		if err != nil {
			return nil, err
		}
		termArgs = append(termArgs, termArg)
	}

	return func(data interface{}) interface{} {
		resp := 0.0
		for _, ta := range termArgs {
			item := toNumber(ta(data))
			if math.IsNaN(item) {
				return item
			}
			resp += item
		}
		return resp
	}, nil
}

func buildUnaryMinusOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	arg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		item := toNumber(arg(data))
		if math.IsNaN(item) {
			return item
		}
		return -1.0 * item
	}, nil
}

func buildMinusOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) == 0 {
		return nullf, nil
	}

	if len(args) == 1 {
		return buildUnaryMinusOp(args, ops)
	}

	var termArgs []ClauseFunc
	for _, a := range args {
		termArg, err := BuildArgFunc(a, ops)
		if err != nil {
			return nil, err
		}
		termArgs = append(termArgs, termArg)
	}

	return func(data interface{}) interface{} {
		resp := toNumber(termArgs[0](data))
		if math.IsNaN(resp) {
			return resp
		}

		for _, ta := range termArgs[1:] {
			item := toNumber(ta(data))
			if math.IsNaN(item) {
				return resp
			}

			resp -= item
		}
		return resp
	}, nil
}

func buildMultiplyOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) == 0 {
		return nullf, nil
	}

	var termArgs []ClauseFunc
	for _, a := range args {
		termArg, err := BuildArgFunc(a, ops)
		if err != nil {
			return nil, err
		}
		termArgs = append(termArgs, termArg)
	}

	return func(data interface{}) interface{} {
		resp := 1.0
		for _, ta := range termArgs {
			item := toNumber(ta(data))
			if math.IsNaN(item) {
				return item
			}
			resp *= item
		}
		return resp
	}, nil
}

func buildDivideOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) < 2 {
		return nullf, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lVal := toNumber(lArg(data))
		rVal := toNumber(rArg(data))

		return lVal / rVal
	}, nil
}

func buildModuloOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) < 2 {
		return nullf, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lVal := toNumber(lArg(data))
		rVal := toNumber(rArg(data))

		return math.Mod(lVal, rVal)

		panic("modulo disjoint types not implemented")
	}, nil
}

func buildMergeOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	switch {
	case len(args) == 0:
		return emptySlice, nil
	}

	var termArgs []ClauseFunc
	for _, a := range args {
		termArg, err := BuildArgFunc(a, ops)
		if err != nil {
			return nil, err
		}
		termArgs = append(termArgs, termArg)
	}

	return func(data interface{}) interface{} {
		resp := []interface{}{}
		for _, ta := range termArgs {
			item := ta(data)
			sliceitem, ok := item.([]interface{})
			if !ok {
				sliceitem = []interface{}{item}
			}
			resp = append(resp, sliceitem...)
		}
		return resp
	}, nil
}

func buildInOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) <= 1 {
		return falsef, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}
	rArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		res := false
		lval := lArg(data)
		rval := rArg(data)

		switch rval := rval.(type) {
		case string:
			lstr := fmt.Sprintf("%v", lval)
			if strings.Contains(rval, lstr) {
				return true
			}
			return false
		case []interface{}:
			for _, r := range rval {
				if IsDeepEqual(lval, r) {
					return true
				}
			}
			return false
		case map[string]interface{}:
			for k := range rval {
				if IsDeepEqual(lval, k) {
					return true
				}
			}
			return false
		default:
		}

		return res
	}, nil
}

func buildCatOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	var termArgs []ClauseFunc
	for _, a := range args {
		termArg, err := BuildArgFunc(a, ops)
		if err != nil {
			return nil, err
		}
		termArgs = append(termArgs, termArg)
	}

	return func(data interface{}) interface{} {
		resp := ""
		for _, ta := range termArgs {
			resp += fmt.Sprintf("%v", ta(data))
		}
		return resp
	}, nil
}

func buildSubstrOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	var err error
	if len(args) == 0 {
		return func(interface{}) interface{} {
			return "undefined"
		}, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	offsetArg := nullf
	if len(args) >= 2 {
		offsetArg, err = BuildArgFunc(args[1], ops)
		if err != nil {
			return nil, err
		}
	}

	lengthArg := nullf
	if len(args) >= 3 {
		lengthArg, err = BuildArgFunc(args[2], ops)
		if err != nil {
			return nil, err
		}
	}

	return func(data interface{}) interface{} {
		lVal := lArg(data)
		offsetVal := offsetArg(data)
		lengthVal := lengthArg(data)

		var base string
		var ok bool
		if base, ok = lVal.(string); !ok {
			base = fmt.Sprintf("%v", lVal)
		}

		baseLen := utf8.RuneCountInString(base)
		if baseLen == 0 {
			return base
		}

		offset := 0.0
		offset, _ = offsetVal.(float64)
		offsetint := int(offset)

		length := 0.0
		length, _ = lengthVal.(float64)
		lengthint := int(length)

		start := 0
		end := baseLen

		switch {
		case offsetint > 0:
			if offsetint > len(base) {
				offsetint = len(base)
			}
			start = offsetint
		case offsetint < 0:
			if offsetint < (-1 * len(base)) {
				offsetint = -1 * len(base)
			}

			start = len(base) + offsetint
		}

		switch {
		case lengthint > 0:
			if start+lengthint > baseLen {
				lengthint = baseLen - start
			}
			end = start + lengthint
		case lengthint < 0:
			remaining := baseLen - start
			if lengthint*-1 > remaining {
				lengthint = remaining * -1
			}
			end += lengthint
		}

		resp := ""
		i := 0
		for _, c := range base {
			if i < start {
				i++
				continue
			}
			if i >= end {
				break
			}

			resp += string(c)
			i++
		}

		return resp
	}, nil
}

func buildMapOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) < 2 {
		return nullf, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	rArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lval := lArg(data)
		lslice, ok := lval.([]interface{})
		if !ok {
			return []interface{}{}
		}

		resp := make([]interface{}, len(lslice))

		for i, subd := range lslice {
			resp[i] = rArg(subd)
		}

		return resp
	}, nil
}

func buildFilterOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) < 2 {
		return nullf, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	rArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lval := lArg(data)
		lslice, ok := lval.([]interface{})
		if !ok {
			return []interface{}{}
		}

		resp := []interface{}{}

		for _, subd := range lslice {
			if IsTrue(rArg(subd)) {
				resp = append(resp, subd)
			}
		}

		return resp
	}, nil
}

func buildReduceOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) < 3 {
		return nullf, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	fArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	initialArg, err := BuildArgFunc(args[2], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lval := lArg(data)
		lslice, ok := lval.([]interface{})
		if !ok {
			return []interface{}{}
		}

		var acc = initialArg(data)

		for _, subd := range lslice {
			acc = fArg(map[string]interface{}{
				"current":     subd,
				"accumulator": acc,
			})
		}

		return acc
	}, nil
}

func buildAllOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) < 2 {
		return nullf, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	fArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lval := lArg(data)
		lslice, ok := lval.([]interface{})
		if !ok {
			return []interface{}{}
		}
		if len(lslice) == 0 {
			return false
		}

		for _, subd := range lslice {
			if !IsTrue(fArg(subd)) {
				return false
			}
		}

		return true
	}, nil
}

func buildSomeOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) < 2 {
		return nullf, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	fArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lval := lArg(data)
		lslice, ok := lval.([]interface{})
		if !ok {
			return []interface{}{}
		}
		if len(lslice) == 0 {
			return false
		}

		for _, subd := range lslice {
			if IsTrue(fArg(subd)) {
				return true
			}
		}

		return false
	}, nil
}

func buildNoneOp(args Arguments, ops OpsSet) (ClauseFunc, error) {
	if len(args) < 2 {
		return nullf, nil
	}

	lArg, err := BuildArgFunc(args[0], ops)
	if err != nil {
		return nil, err
	}

	fArg, err := BuildArgFunc(args[1], ops)
	if err != nil {
		return nil, err
	}

	return func(data interface{}) interface{} {
		lval := lArg(data)
		lslice, ok := lval.([]interface{})
		if !ok {
			return []interface{}{}
		}
		if len(lslice) == 0 {
			return true
		}

		for _, subd := range lslice {
			if IsTrue(fArg(subd)) {
				return false
			}
		}

		return true
	}, nil
}

// Compile compiles a given clause using the operation constructors in this
// OpsSet
func (ops OpsSet) Compile(c *Clause) (ClauseFunc, error) {
	bf, ok := ops[c.Operator.Name]
	if !ok {
		return nil, fmt.Errorf("unrecognized operation %s", c.Operator.Name)
	}
	return bf(c.Arguments, ops)
}

// DefaultOps is the default set of operations as specified on the jsonlogic
// site.
var DefaultOps = OpsSet{
	nullOp:          buildNullOp,
	varOp:           buildVarOp,
	missingOp:       buildMissingOp,
	missingSomeOp:   buildMissingSomeOp,
	ifOp:            buildIfOp,
	ternaryOp:       buildTernaryOp,
	andOp:           buildAndOp,
	orOp:            buildOrOp,
	equalOp:         buildEqualOp,
	equalThreeOp:    buildEqualThreeOp,
	notEqualOp:      buildNotEqualOp,
	notEqualThreeOp: buildNotEqualThreeOp,
	negateOp:        buildNegateOp,
	doubleNegateOp:  buildDoubleNegateOp,
	lessOp:          buildLessOp,
	lessEqOp:        buildLessEqualOp,
	greaterOp:       buildGreaterOp,
	greaterEqOp:     buildGreaterEqualOp,
	minOp:           buildMinOp,
	maxOp:           buildMaxOp,
	plusOp:          buildPlusOp,
	minusOp:         buildMinusOp,
	multiplyOp:      buildMultiplyOp,
	divideOp:        buildDivideOp,
	moduloOp:        buildModuloOp,
	inOp:            buildInOp,
	catOp:           buildCatOp,
	substrOp:        buildSubstrOp,

	mapOp:    buildMapOp,
	filterOp: buildFilterOp,
	reduceOp: buildReduceOp,
	allOp:    buildAllOp,
	someOp:   buildSomeOp,
	noneOp:   buildNoneOp,

	mergeOp: buildMergeOp,
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
