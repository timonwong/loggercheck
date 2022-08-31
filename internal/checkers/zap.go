package checkers

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

type Zap struct {
	General
}

func (z Zap) ExtractLoggingKeyAndValues(pass *analysis.Pass, call *CallContext) []ast.Expr {
	args := call.Expr.Args
	params := call.Signature.Params()

	nparams := params.Len() // variadic => nonzero
	startIndex := nparams - 1
	nargs := len(args)

	// Check the argument count
	keyValuesArgs := make([]ast.Expr, 0, nargs-startIndex)
	for i := startIndex; i < nargs; i++ {
		arg := args[i]
		switch arg := arg.(type) {
		case *ast.CallExpr, *ast.Ident:
			typ := pass.TypesInfo.TypeOf(arg)
			switch typ := typ.(type) {
			case *types.Named:
				obj := typ.Obj()
				// This is a strongly-typed field. Consume it and move on.
				// Actually it's go.uber.org/zap/zapcore.Field, however for simplicity
				// we don't check the import path
				if obj != nil && obj.Name() == "Field" {
					continue
				}
			default:
				// pass
			}
		}
		keyValuesArgs = append(keyValuesArgs, arg)
	}
	return keyValuesArgs
}

var _ Checker = (*Zap)(nil)
