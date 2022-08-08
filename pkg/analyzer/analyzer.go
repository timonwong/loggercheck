package analyzer

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "logrlint",
	Doc:      "Check logr arguments.",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

const (
	LogrFullName = "github.com/go-logr/logr.Logger"
)

type CheckFunc func(*analysis.Pass, []ast.Expr)

var MethodsToCheck = map[string]CheckFunc{
	"Error":      checkEvenArguments(2),
	"Info":       checkEvenArguments(1),
	"WithValues": checkEvenArguments(0),
}

func checkEvenArguments(argToSkip int) CheckFunc {
	return func(pass *analysis.Pass, args []ast.Expr) {
		if len(args) <= argToSkip {
			return
		}

		count := len(args) - argToSkip
		if count%2 != 0 {
			lastArg := args[len(args)-1]
			pass.Reportf(lastArg.Pos(), "odd number of arguments passed as key-value pairs for logging")
		}
	}
}

func isLogrInstance(pass *analysis.Pass, expr *ast.SelectorExpr) bool {
	if expr, ok := expr.X.(*ast.SelectorExpr); ok {
		return isLogrInstance(pass, expr)
	}

	objectOf := pass.TypesInfo.ObjectOf(expr.Sel)
	typ := objectOf.Type()
	if typ, ok := typ.(*types.Named); !ok {
		return false
	} else if typ.String() != LogrFullName {
		return false
	}

	return true
}

func run(pass *analysis.Pass) (interface{}, error) {
	callExprs := []ast.Node{&ast.CallExpr{}}
	inp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inp.Nodes(callExprs, func(node ast.Node, push bool) bool {
		callExpr, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		funExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		funName := funExpr.Sel.Name
		checker, ok := MethodsToCheck[funName]
		if !ok {
			return true
		}

		switch expr := funExpr.X.(type) {
		case *ast.CallExpr:
			exprFunExpr, ok := expr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			if !isLogrInstance(pass, exprFunExpr) {
				return true
			}
		case *ast.SelectorExpr:
			if !isLogrInstance(pass, expr) {
				return true
			}
		default:
			return true
		}

		// dot dot dot is hard, just skip...
		if callExpr.Ellipsis.IsValid() {
			return false
		}

		checker(pass, callExpr.Args)
		return false
	})

	return nil, nil
}
