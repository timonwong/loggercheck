package checkers

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

type Config struct {
	RequireStringKey bool
}

type CallContext struct {
	Expr      *ast.CallExpr
	Func      *types.Func
	Signature *types.Signature
}

type Checker interface {
	CheckAndReport(pass *analysis.Pass, call CallContext, cfg Config)
}
