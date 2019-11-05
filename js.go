package jsonlogic

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// IsTrue implements the truthy/falsy semantics of jsonlogic
// as documented here: https://jsonlogic.com/truthy.html
// in addition, an emptry struct is considered true, as
// per jsonlogic JavaScript handling of "{}"
func IsTrue(i interface{}) bool {
	if i == nil {
		return false
	}
	switch v := i.(type) {
	case float64:
		return v != 0
	case map[string]interface{}:
		return true
	case []interface{}:
		return len(v) != 0
	case string:
		return v != ""
	case bool:
		return v
	default:
		return true
	}
}

func toNumber(i interface{}) interface{} {
	switch v := i.(type) {
	case string:
		tstr := strings.TrimSpace(v)
		if tstr == "" {
			return 0.0
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return math.NaN()
		}
		return f
	case float64:
		return v
	case bool:
		if v {
			return 1.0
		}
		return 0.0
	case []interface{}:
		if len(v) == 0 {
			return 0.0
		}
		if len(v) == 1 {
			return toNumber(v[0])
		}
		return false
	case nil:
		return 0.0
	default:
		return math.NaN()
	}
}

func toString(i interface{}) interface{} {
	switch v := i.(type) {
	case string:
		return i
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return "null"
	case []interface{}:
		var strs = make([]string, len(v))
		for i := range v {
			strs[i] = toString(v[i]).(string)
		}
		return strings.Join(strs, ",")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toBool(i interface{}) (interface{}, bool) {
	return IsTrue(i), true
}

// IsEqual is an exact equality check.
func IsEqual(l, r interface{}) bool {
	_, lisfloat := l.(float64)
	_, lisstr := l.(string)
	_, lisbool := l.(bool)

	_, risfloat := r.(float64)
	_, risstr := r.(string)
	_, risbool := r.(bool)

	switch {
	case l == nil && r == nil:
		return true
	case
		lisfloat && risfloat,
		lisbool && risbool,
		lisstr && risstr:
		return l == r
	default:
		return reflect.DeepEqual(l, r)
	}
}

// IsSoftEqual is an equality check that will
// coerce values according to JavaScript rules.
func IsSoftEqual(l, r interface{}) bool {
	_, lisfloat := l.(float64)
	_, lisstr := l.(string)
	_, lisbool := l.(bool)

	_, risfloat := r.(float64)
	_, risstr := r.(string)
	_, risbool := r.(bool)

	switch {
	case l == nil && r == nil:
		return true
	case
		lisbool && risbool,
		lisfloat && risfloat,
		lisstr && risstr:
		return l == r
	case lisstr || risstr:
		return fmt.Sprintf("%v", l) == fmt.Sprintf("%v", r)
	case lisbool || risbool:
		return IsTrue(l) == IsTrue(r)
	default:
		return reflect.DeepEqual(l, r)
	}
}
