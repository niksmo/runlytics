package main

import (
	testifylint "github.com/Antonboom/testifylint/analyzer"
	"github.com/niksmo/runlytics/pkg/osexit"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

type include map[string]struct{}

var includedSimpleChecks = include{
	"S1034": {}, // Use result of type assertion to simplify cases
}

var includedStyleChecks = include{
	"ST1005": {}, // Incorrectly formatted error string
}

var includedQuickFixChecks = include{
	"QF1011": {}, // Omit redundant type from variable declaration
}

func main() {
	checks := []*analysis.Analyzer{
		// Detects if there is only one variable in append.
		appends.Analyzer,

		// Detects useless assignments.
		assign.Analyzer,

		// Checks for common mistakes using the sync/atomic package.
		atomic.Analyzer,

		// Detects common mistakes involving boolean operators.
		bools.Analyzer,

		// Checks for unkeyed composite literals.
		composite.Analyzer,

		// Checks for locks erroneously passed by value.
		copylock.Analyzer,

		// Checks for common mistakes in defer statements.
		defers.Analyzer,

		// Checks that the second argument to errors.As is
		// a pointer to a type implementing error.
		errorsas.Analyzer,

		// Checks for mistakes using HTTP responses.
		httpresponse.Analyzer,

		// Checks for references to enclosing loop
		// variables from within nested functions.
		loopclosure.Analyzer,

		// Checks for failure to call a context cancellation function.
		lostcancel.Analyzer,

		// Checks for useless comparisons against nil.
		nilfunc.Analyzer,

		// Checks consistency of Printf format strings and arguments.
		printf.Analyzer,

		// Checks for shadowed variables.
		shadow.Analyzer,

		// Flags type conversions from integers to strings.
		stringintconv.Analyzer,

		// Checks struct field tags are well formed.
		structtag.Analyzer,

		// Checks for passing non-pointer or non-interface
		// types to unmarshal and decode functions.
		unmarshal.Analyzer,

		// Checks for unreachable code.
		unreachable.Analyzer,

		// Checks for unused results of calls to certain pure functions.
		unusedresult.Analyzer,

		// Checks for unused writes to the elements of a struct or array object.
		unusedwrite.Analyzer,

		// Checks for usage of generic features.
		usesgenerics.Analyzer,

		// Detects simple misuses of sync.WaitGroup.
		waitgroup.Analyzer,

		// Checks testify methods misuses.
		testifylint.New(),

		// Checks whether HTTP response body is closed successfully.
		bodyclose.Analyzer,

		// Detects os.Exit direct call in main function.
		osexit.Analyzer(),
	}

	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	checks = appendIncluded(checks, simple.Analyzers, includedSimpleChecks)
	checks = appendIncluded(checks, stylecheck.Analyzers, includedStyleChecks)
	checks = appendIncluded(checks, quickfix.Analyzers, includedQuickFixChecks)
	multichecker.Main(checks...)
}

func appendIncluded(
	checks []*analysis.Analyzer,
	lintAnalyzers []*lint.Analyzer,
	included include,
) []*analysis.Analyzer {
	for _, v := range lintAnalyzers {
		if _, ok := included[v.Analyzer.Name]; ok {
			checks = append(checks, v.Analyzer)
		}
	}
	return checks
}
