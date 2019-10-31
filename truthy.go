package jsonlogic

import "fmt"

// IsTrue implements the truthy/falsy semantics of jsonlogic
// as documented here: https://jsonlogic.com/truthy.html
// in addition, an emptry struct is considerd true, as
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
