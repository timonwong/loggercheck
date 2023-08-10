package checkers

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

type Slog struct {
	General
}

func (z Slog) FilterKeyAndValues(pass *analysis.Pass, keyAndValues []ast.Expr) []ast.Expr {
	// Check the argument count
	filtered := make([]ast.Expr, 0, len(keyAndValues))
	for _, arg := range keyAndValues {
		// Skip any zapcore.Field we found
		switch arg := arg.(type) {
		case *ast.CallExpr, *ast.Ident:
			typ := pass.TypesInfo.TypeOf(arg)
			switch typ := typ.(type) {
			case *types.Named:
				obj := typ.Obj()
				// check slog.Group() constructed group slog.Attr
				if obj != nil && obj.Name() == "Attr" {
					// since we also check `slog.Group` so it is OK skip here
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

var _ Checker = (*Slog)(nil)
