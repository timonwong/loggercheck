package checkers

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"

	"github.com/timonwong/loggercheck/internal/bytebufferpool"
	"github.com/timonwong/loggercheck/internal/stringutil"
)

// getStringValueFromArg returns true if the argument is string literal or string constant.
func getStringValueFromArg(pass *analysis.Pass, arg ast.Expr) (value string, ok bool) {
	switch arg := arg.(type) {
	case *ast.BasicLit: // literals, must be string
		if arg.Kind == token.STRING {
			return arg.Value, true
		}
	case *ast.Ident: // identifiers, we require constant string key
		if arg.Obj != nil && arg.Obj.Kind == ast.Con {
			typeAndValue := pass.TypesInfo.Types[arg]
			if typ, ok := typeAndValue.Type.(*types.Basic); ok {
				if typ.Kind() == types.String {
					return typeAndValue.Value.ExactString(), true
				}
			}
		}
	}

	return "", false
}

func checkLoggingKey(pass *analysis.Pass, keyValuesArgs []ast.Expr) {
	for i := 0; i < len(keyValuesArgs); i += 2 {
		arg := keyValuesArgs[i]
		if value, ok := getStringValueFromArg(pass, arg); ok {
			if stringutil.IsASCII(value) {
				continue
			}

			pass.Report(analysis.Diagnostic{
				Pos:      arg.Pos(),
				End:      arg.End(),
				Category: "logging",
				Message: fmt.Sprintf(
					"logging keys are expected to be alphanumeric strings, please remove any non-latin characters from %s",
					value),
			})
		} else {
			pass.Report(analysis.Diagnostic{
				Pos:      arg.Pos(),
				End:      arg.End(),
				Category: "logging",
				Message: fmt.Sprintf(
					"logging keys are expected to be inlined constant strings, please replace %q provided with string",
					renderNodeEllipsis(pass.Fset, arg)),
			})
		}
	}
}

func renderNodeEllipsis(fset *token.FileSet, v interface{}) string {
	const maxLen = 20

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	_ = printer.Fprint(buf, fset, v)
	s := buf.String()
	if utf8.RuneCountInString(s) > maxLen {
		// Copied from go/constant/value.go
		i := 0
		for n := 0; n < maxLen-3; n++ {
			_, size := utf8.DecodeRuneInString(s[i:])
			i += size
		}
		s = s[:i] + "..."
	}
	return s
}
