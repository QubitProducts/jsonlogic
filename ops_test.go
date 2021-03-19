package jsonlogic_test

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/tcolgate/jsonlogic"
)

func init() {
}

// ExampleOpsSet demonstrates extending the set of available operations.
//
// Operations are collected together in an OpSet that can be asked to compile a
// given clause. Each operator (op) supported by a given OpSet must implement
// a BuildArgFunc that compiles that operation and returns the resulting
// ClauseFunc that will execute the operation at evaluation time. The build
// function is usually devided into two stages.
//
// At Compile time we check we have the correct number of arguments, and deal
// with errors and boundary conditions. We then call the BuildArgFunc for each
// argument that we would like to use at Evaluation time. The compile time stage
// then returns a closure over the arg functions we collect as a ClausFunc, that
// does the Exectution time evaluation.
//
// At Execution time, we call the claus functions we collected in the closure to
// retrieve the result of the argument for this particular call, and return
// and answer as neccessary.
//
// The following example implements a regular expression match operation.
func ExampleOpsSet() {
	buildMatchOp := func(args jsonlogic.Arguments, ops jsonlogic.OpsSet) (jsonlogic.ClauseFunc, error) {
		// We want two args, 1 for the data we are going to match, and one
		// for the regex itself.
		if len(args) < 2 {
			return func(ctx context.Context, data interface{}) interface{} {
				return false
			}, nil
		}

		// We build a function for the data the user
		// gives us to match against.
		lArg, err := jsonlogic.BuildArgFunc(args[0], ops)
		if err != nil {
			return nil, err
		}

		// We build a function for the second argument that
		// we hope will become a string to can compile to
		// a regexp.
		rArg, err := jsonlogic.BuildArgFunc(args[1], ops)
		if err != nil {
			return nil, err
		}

		return func(ctx context.Context, data interface{}) interface{} {
			// We evaluate the first argument, using
			// the data the user provided.
			lval := lArg(ctx, data)
			lstr, ok := lval.(string)
			if !ok {
				// We only match against strings, everything else
				// is false.
				return false
			}

			// We evaluate the second argument, using
			// the data the user provided.
			rval := rArg(ctx, data)
			rstr, ok := rval.(string)
			if !ok {
				// We can only build regexp out of strings.
				return false
			}

			// we compile oour string regexp (this could be cached).
			rx, err := regexp.Compile(rstr)
			if err != nil {
				return false // JsonLogic never errors, bad things return false
			}

			// Finally we call out string match and return the answer
			// we shoud on return string, float64, bool, []map[string]interface{},
			// []interface{} (where the interfaces only consist of the same set of
			// types.
			return rx.MatchString(lstr)
		}, nil
	}

	// Add our function to an OpSet
	ops := jsonlogic.DefaultOps
	ops["match"] = buildMatchOp

	cls := jsonlogic.Clause{}
	_ = json.Unmarshal([]byte(`{"match": [{"var":""},"this"]}`), &cls)

	cf, _ := ops.Compile(&cls)

	var tests = []interface{}{
		`this matches`,
		`that doesn't`,
		float64(1),
	}
	ctx := context.Background()
	for _, t := range tests {
		fmt.Printf("match(%#v) = %v\n", t, jsonlogic.IsTrue(cf(ctx, t)))
	}
	// Output:
	// match("this matches") = true
	// match("that doesn't") = false
	// match(1) = false
}
