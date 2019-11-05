package jsonlogic

import (
	"fmt"
	"reflect"
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
		panic(fmt.Errorf("unhandled type %T", i))
	}
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
	case
		lisbool && risbool,
		lisfloat && risfloat,
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
