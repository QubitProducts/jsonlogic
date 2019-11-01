package jsonlogic

import (
	"strconv"
	"strings"
)

func deref(data interface{}, ref []string) interface{} {
	if len(ref) == 0 {
		return nil
	}
	switch data := data.(type) {
	case map[string]interface{}:
		val, ok := data[ref[0]]
		if !ok {
			return nil
		}
		if len(ref) == 1 {
			return val
		}
		return deref(val, ref[1:])
	}
	return nil
}

// DottedRef attempts to resolve a dotted reference into a
// Go type. Only map[string]interface{} is supported for now.
func DottedRef(data interface{}, ref interface{}) interface{} {
	var refStr string
	switch ref := ref.(type) {
	case string:
		refStr = ref
	case float64:
		intref := int(ref)
		if ref != float64(intref) || intref < 0 {
			return nil
		}
		refStr = strconv.Itoa(intref)
	default:
		return nil
	}

	return deref(data, strings.Split(refStr, "."))
}
