package logrlint

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

var Analyzer = newAnalyzer()

func newAnalyzer() *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name:     "logrlint",
		Doc:      "Check logr and klog arguments.",
		Run:      run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
	return a
}

var validLoggerToFuncNames = map[string]stringSet{
	"github.com/go-logr/logr": newStringSet([]string{"Error", "Info", "WithValues"}),
	"k8s.io/klog/v2":          newStringSet([]string{"InfoS", "InfoSDepth", "ErrorS"}),
}

func getLoggerFuncNames(pkgPath, callerPkgPath string) stringSet {
	for loggerPkg, names := range validLoggerToFuncNames {
		if loggerPkg == pkgPath {
			return names
		}

		vendorPath := fmt.Sprintf("%s/vendor/%s", callerPkgPath, loggerPkg)
		if vendorPath == pkgPath {
			return names
		}
	}
	return nil
}

func isValidLoggerFuncName(pass *analysis.Pass, fn *types.Func) bool {
	pkg := fn.Pkg()
	if pkg == nil {
		return false
	}

	names := getLoggerFuncNames(pkg.Path(), pass.Pkg.Name())
	return names.has(fn.Name())
}

func checkLoggerArguments(pass *analysis.Pass, call *ast.CallExpr) {
	fn, _ := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
	if fn == nil {
		return // function pointer is not supported
	}

	sig, ok := fn.Type().(*types.Signature)
	if !ok || !sig.Variadic() {
		return // not variadic
	}

	if !isValidLoggerFuncName(pass, fn) {
		return
	}

	// ellipsis args is hard, just skip
	if call.Ellipsis.IsValid() {
		return
	}

	params := sig.Params()
	nparams := params.Len() // variadic => nonzero
	args := params.At(nparams - 1)
	iface, ok := args.Type().(*types.Slice).Elem().(*types.Interface)
	if !ok || !iface.Empty() {
		return // final (args) param is not ...interface{}
	}

	startIndex := nparams - 1
	nargs := len(call.Args)
	variadicLen := nargs - startIndex
	if variadicLen%2 != 0 {
		firstArg := call.Args[startIndex]
		lastArg := call.Args[nargs-1]
		pass.Report(analysis.Diagnostic{
			Pos:      firstArg.Pos(),
			End:      lastArg.End(),
			Category: "logging",
			Message:  "odd number of arguments passed as key-value pairs for logging"})
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	insp.Preorder(nodeFilter, func(node ast.Node) {
		call := node.(*ast.CallExpr)

		typ := pass.TypesInfo.Types[call.Fun].Type
		if typ == nil {
			// Skip checking functions with unknown type.
			return
		}

		checkLoggerArguments(pass, call)
	})

	return nil, nil
}
