package jsonlogic

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
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

func toNumber(i interface{}) float64 {
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
		return math.NaN()
	case nil:
		return 0.0
	default:
		return math.NaN()
	}
}

func toString(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
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
			strs[i] = toString(v[i])
		}
		return strings.Join(strs, ",")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toBool(i interface{}) bool {
	return IsTrue(i)
}

// IsEqual is an exact equality check.
func IsEqual(l, r interface{}) bool {
	_, lisfloat := l.(float64)
	_, lisstr := l.(string)
	_, lisbool := l.(bool)
	lslice, lisslice := l.([]interface{})
	lmap, lismap := l.(map[string]interface{})

	_, risfloat := r.(float64)
	_, risstr := r.(string)
	_, risbool := r.(bool)
	rslice, risslice := r.([]interface{})
	rmap, rismap := r.(map[string]interface{})

	switch {
	case l == nil && r == nil:
		return true
	case l == nil || r == nil:
		return false
	case lismap && rismap:
		// compare two maps, must be the same object
		lptr := reflect.ValueOf(lmap)
		rptr := reflect.ValueOf(rmap)
		return lptr == rptr
	case lisslice && risslice:
		lhdr := (*reflect.SliceHeader)(unsafe.Pointer(&lslice))
		rhdr := (*reflect.SliceHeader)(unsafe.Pointer(&rslice))
		return *lhdr == *rhdr
	case
		lisbool && risbool,
		lisfloat && risfloat,
		lisstr && risstr:
		return l == r
	default:
		return false
	}
}

// IsSoftEqual is an equality check that will
// coerce values according to JavaScript rules.
func IsSoftEqual(l, r interface{}) bool {
	_, lisfloat := l.(float64)
	_, lisstr := l.(string)
	_, lisbool := l.(bool)
	lslice, lisslice := l.([]interface{})
	lmap, lismap := l.(map[string]interface{})

	_, risfloat := r.(float64)
	_, risstr := r.(string)
	_, risbool := r.(bool)
	rslice, risslice := r.([]interface{})
	rmap, rismap := r.(map[string]interface{})

	switch {
	case l == nil && r == nil:
		return true
	case l == nil || r == nil:
		return false
	case lismap && rismap:
		// compare two maps, must be the same object
		lptr := reflect.ValueOf(lmap)
		rptr := reflect.ValueOf(rmap)
		return lptr == rptr
	case lisslice && risslice:
		lhdr := (*reflect.SliceHeader)(unsafe.Pointer(&lslice))
		rhdr := (*reflect.SliceHeader)(unsafe.Pointer(&rslice))
		return *lhdr == *rhdr
	case lisslice || risslice:
		if lisslice {
			return IsSoftEqual(toString(l), r)
		}
		return IsSoftEqual(l, toString(r))
	case lismap || rismap:
		return false
	case
		lisbool && risbool,
		lisfloat && risfloat,
		lisstr && risstr:
		return l == r
	default:
		return toNumber(l) == toNumber(r)
	}
}
