package checkers

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

type Config struct {
	RequireStringKey bool
	NoPrintfLike     bool
}

type CallContext struct {
	Expr      *ast.CallExpr
	Func      *types.Func
	Signature *types.Signature
}

type Checker interface {
	ExtractLoggingKeyAndValues(pass *analysis.Pass, call *CallContext) []ast.Expr
	CheckLoggingKey(pass *analysis.Pass, keyValuesArgs []ast.Expr)
	CheckPrintfLikeSpecifier(pass *analysis.Pass, messageArgs []ast.Expr)
}

func ExecuteChecker(c Checker, pass *analysis.Pass, call *CallContext, cfg Config) {
	keyValuesArgs := c.ExtractLoggingKeyAndValues(pass, call)

	if len(keyValuesArgs)%2 != 0 {
		firstArg := keyValuesArgs[0]
		lastArg := keyValuesArgs[len(keyValuesArgs)-1]
		pass.Report(analysis.Diagnostic{
			Pos:      firstArg.Pos(),
			End:      lastArg.End(),
			Category: DiagnosticCategory,
			Message:  "odd number of arguments passed as key-value pairs for logging",
		})
	}

	if cfg.RequireStringKey {
		c.CheckLoggingKey(pass, keyValuesArgs)
	}

	if cfg.NoPrintfLike {
		c.CheckPrintfLikeSpecifier(pass, call.Expr.Args)
	}
}
