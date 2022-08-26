package logrlint

import (
	"flag"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

const Doc = `Checks logr and klog arguments.`

func NewAnalyzer() *analysis.Analyzer {
	l := &logrlint{}
	a := &analysis.Analyzer{
		Name:     "logrlint",
		Doc:      Doc,
		Run:      l.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
	a.Flags.Init("logrlint", flag.ExitOnError)
	a.Flags.BoolVar(&l.disableAll, "disableall", false, "disable all logger checkers")
	a.Flags.Var(&l.disable, "disable", "comma-separated list of disabled logger checker")
	a.Flags.Var(&l.enable, "enable", "comma-separated list of enabled logger checker")
	return a
}

type logrlint struct {
	disableAll bool               // flag -disableall
	disable    loggerCheckersFlag // flag -disable
	enable     loggerCheckersFlag // flag -enable
}

type loggerCheck struct {
	packageImport string
	funcNames     stringSet
}

var loggerCheckersByName = map[string]loggerCheck{
	"logr": {
		packageImport: "github.com/go-logr/logr",
		funcNames:     newStringSet("Error", "Info", "WithValues"),
	},
	"klog": {
		packageImport: "k8s.io/klog/v2",
		funcNames:     newStringSet("InfoS", "InfoSDepth", "ErrorS"),
	},
}

func (l *logrlint) isCheckerDisabled(name string) bool {
	if l.disableAll {
		return !l.enable.Has(name)
	}
	return l.disable.Has(name)
}

func (l *logrlint) getLoggerFuncNames(pkgPath string) stringSet {
	for name, entry := range loggerCheckersByName {
		if l.isCheckerDisabled(name) {
			// Skip ignored logger checker.
			continue
		}

		if entry.packageImport == pkgPath {
			return entry.funcNames
		}

		if strings.HasSuffix(pkgPath, "/vendor/"+entry.packageImport) {
			return entry.funcNames
		}
	}

	return nil
}

func (l *logrlint) isValidLoggerFunc(fn *types.Func) bool {
	pkg := fn.Pkg()
	if pkg == nil {
		return false
	}

	names := l.getLoggerFuncNames(pkg.Path())
	return names.Has(fn.Name())
}

func (l *logrlint) checkLoggerArguments(pass *analysis.Pass, call *ast.CallExpr) {
	fn, _ := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
	if fn == nil {
		return // function pointer is not supported
	}

	sig, ok := fn.Type().(*types.Signature)
	if !ok || !sig.Variadic() {
		return // not variadic
	}

	if !l.isValidLoggerFunc(fn) {
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
			Message:  "odd number of arguments passed as key-value pairs for logging",
		})
	}
}

func (l *logrlint) run(pass *analysis.Pass) (interface{}, error) {
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

		l.checkLoggerArguments(pass, call)
	})

	return nil, nil
}
