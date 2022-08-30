package checkers

import (
	"golang.org/x/tools/go/analysis"
)

type General struct{}

var _ Checker = (*General)(nil)

func (g General) CheckAndReport(pass *analysis.Pass, call CallContext, cfg Config) {
	args := call.Expr.Args
	params := call.Signature.Params()

	nparams := params.Len() // variadic => nonzero
	startIndex := nparams - 1
	nargs := len(args)

	// Check the argument count
	variadicLen := nargs - startIndex
	if variadicLen%2 != 0 {
		firstArg := args[startIndex]
		lastArg := args[nargs-1]
		pass.Report(analysis.Diagnostic{
			Pos:      firstArg.Pos(),
			End:      lastArg.End(),
			Category: "logging",
			Message:  "odd number of arguments passed as key-value pairs for logging",
		})
	}

	// Check the "key" type
	if cfg.RequireStringKey {
		checkLoggingKey(pass, args[startIndex:])
	}
}
