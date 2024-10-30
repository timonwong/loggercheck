package checkers

import (
	"go/ast"
	"go/types"

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
			switch typ := typ.(type) {
			case *types.Alias, *types.Named:
				var obj *types.TypeName
				if cTyp, ok := typ.(*types.Alias); ok {
					obj = cTyp.Obj()
				} else {
					obj = typ.(*types.Named).Obj()
				}
				if obj != nil && obj.Name() == objName {
					continue
				}
			default:
				// pass
			}
		}

		filtered = append(filtered, arg)
	}

	return filtered
}
