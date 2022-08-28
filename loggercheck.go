package loggercheck

import (
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

const Doc = `Checks key valur pairs for common logger libraries (logr,klog,zap).`

func NewAnalyzer(opts ...Option) *analysis.Analyzer {
	l := &loggercheck{}
	for _, o := range opts {
		o(l)
	}

	l.cfg.init(l)
	a := &analysis.Analyzer{
		Name:     "loggercheck",
		Doc:      Doc,
		Run:      l.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	checkerKeys := strings.Join(staticPatternGroups.Names(), ",")
	a.Flags.Init("loggercheck", flag.ExitOnError)
	a.Flags.Var(&l.patternFile, "patternfile", "path to a file contains a list of patterns")
	a.Flags.Var(&l.disable, "disable", fmt.Sprintf("comma-separated list of disabled logger checker (%s)", checkerKeys))
	return a
}

type loggercheck struct {
	disable     loggerCheckersFlag // flag -disable
	patternFile patternFileFlag    // flag -patternfile

	cfg *Config // used for external integration, for example golangci-lint
}

func (l *loggercheck) isCheckerDisabled(name string) bool {
	return l.disable.Has(name)
}

func (l *loggercheck) isValidLoggerFunc(fn *types.Func) bool {
	pkg := fn.Pkg()
	if pkg == nil {
		return false
	}

	for i := range staticPatternGroups {
		pg := &staticPatternGroups[i]
		if l.isCheckerDisabled(pg.Name) {
			// Skip ignored logger checker.
			continue
		}

		if pg.Match(fn, pkg) {
			return true
		}
	}

	customPatternGroups := l.patternFile.patternGroups
	for i := range customPatternGroups {
		pg := &customPatternGroups[i]
		if pg.Match(fn, pkg) {
			return true
		}
	}

	return false
}

func (l *loggercheck) checkLoggerArguments(pass *analysis.Pass, call *ast.CallExpr) {
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

func (l *loggercheck) run(pass *analysis.Pass) (interface{}, error) {
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
