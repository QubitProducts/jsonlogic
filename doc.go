// Package jsonlogic is a Go implementation of JsonLogic as described
// at http://jsonlogic.com.  All example content for all operations, as given on
// the web site, work as expected. The test suite provided at
// https://jsonlogic.com/tests.json passes unaltered.
//
// var data lookup assumes the passed-in data structure is akin to that
// provided  by a raw json.Unmarshal. Only the following types are
// supported:
//   primitives of type string, float64 or bool
//   map[string]interface{} // where interface{} is a compatible type
//   []interface{} // where interface{} is a compatible type
//
// We cannot currently query Go native structs, or maps/slices of other native Go
// types.
//
// Incompatibilities may exist in support for JavaScript type coercion. Any
// incompatibilities found should be reported as bugs.
package jsonlogic
