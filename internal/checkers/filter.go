package checkers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func filterKeyAndValues(pass *analysis.Pass, keyAndValues []ast.Expr, objName string) []ast.Expr {
	// Check the argument count
	filtered := make([]ast.Expr, 0, len(keyAndValues))
	for _, arg := range keyAndValues {
		// Skip any object type field we found
		switch arg := arg.(type) {
		case *ast.CallExpr, *ast.Ident:
			typ := pass.TypesInfo.TypeOf(arg)

			if typ, ok := typ.(commonAlias); ok {
				obj := typ.Obj()
				if obj != nil && obj.Name() == objName {
					continue
				}
			}
		}

		filtered = append(filtered, arg)
	}

	return filtered
}
