// Package jsonlogic is a Go implementation of the jsonlogic as described.
// ar http://jsonlogic.com.  All example content for all operations, as given on
// the web site work as expected.
//
// var data lookup assumes the parsed in data scructure is akin to that
// provided  by a raw json.Unmarshal. That is only the following types are
// supported:
//   primitives of type string, float64 or bool
//   map[string]interface{} // where interface{} is a compatible type
//   []interface{} // where interface{} is a compatible type
//
// We cannot currently query Go native structs, or maps/slices of other native Go
// types.
//
// Incompatibilities may exist in support for javascript type coercion. Any
// incompatibilties found should be reported as bugs.
package jsonlogic
