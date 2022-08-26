package loggercheck

import (
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

const Doc = `Checks key valur pairs for common logger libraries (logr,klog,zap).`

func NewAnalyzer(opts ...Option) *analysis.Analyzer {
	l := &loggercheck{
		disable: loggerCheckersFlag{
			newStringSet(),
		},
	}

	for _, o := range opts {
		o(l)
	}

	if l.config.cfg != nil {
		l.disable = loggerCheckersFlag{
			newStringSet(l.config.cfg.Disable...),
		}
		for _, ck := range l.config.cfg.CustomCheckers {
			addLogger(ck.Name, ck.PackageImport, ck.Funcs)
		}
	}

	a := &analysis.Analyzer{
		Name:     "loggercheck",
		Doc:      Doc,
		Run:      l.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	initFlags(&a.Flags, l)

	return a
}

func initFlags(fs *flag.FlagSet, l *loggercheck) {
	checkerKeys := strings.Join(loggerCheckersByName.Keys(), ",")
	fs.Init("loggercheck", flag.ExitOnError)
	fs.Var(&l.config, "config", `config file path, use "sample" as filename to get sample config`)
	fs.Var(&l.disable, "disable", fmt.Sprintf("comma-separated list of disabled logger checker (%s)", checkerKeys))
}

type loggercheck struct {
	disable loggerCheckersFlag // flag -disable
	config  configFlag         // flag -cfg
}

func (l *loggercheck) isCheckerDisabled(name string) bool {
	return l.disable.Has(name)
}

func (l *loggercheck) getLoggerFuncs(pkgPath string) stringSet {
	for name, entry := range loggerCheckersByName {
		if l.isCheckerDisabled(name) {
			// Skip ignored logger checker.
			continue
		}

		if entry.packageImport == pkgPath {
			return entry.funcs
		}

		if strings.HasSuffix(pkgPath, "/vendor/"+entry.packageImport) {
			return decorateVendoredFuncs(entry.funcs, pkgPath, entry.packageImport)
		}
	}

	return nil
}

func decorateVendoredFuncs(entryFuncs stringSet, currentPkgImport, packageImport string) stringSet {
	funcs := make(stringSet, len(entryFuncs))
	for fn := range entryFuncs {
		lastDot := strings.LastIndex(fn, ".")
		if lastDot == -1 {
			continue // invalid pattern
		}

		importOrReceiver := fn[:lastDot]
		fnName := fn[lastDot+1:]

		if strings.HasPrefix(importOrReceiver, "(") { // is receiver
			if !strings.HasSuffix(importOrReceiver, ")") {
				continue // invalid pattern
			}

			var pointerIndicator string
			if strings.HasPrefix(importOrReceiver[1:], "*") { // pointer type
				pointerIndicator = "*"
			}

			leftOver := strings.TrimPrefix(importOrReceiver, "("+pointerIndicator+packageImport+".")
			importOrReceiver = fmt.Sprintf("(%s%s.%s", pointerIndicator, currentPkgImport, leftOver)
		} else { // is import
			importOrReceiver = currentPkgImport
		}

		fn = fmt.Sprintf("%s.%s", importOrReceiver, fnName)
		funcs.Insert(fn)
	}
	return funcs
}

func (l *loggercheck) isValidLoggerFunc(fn *types.Func) bool {
	pkg := fn.Pkg()
	if pkg == nil {
		return false
	}

	funcs := l.getLoggerFuncs(pkg.Path())
	return funcs.Has(fn.FullName())
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

var cfgInitOnce sync.Once

func (l *loggercheck) run(pass *analysis.Pass) (interface{}, error) {
	if l.config.cfg != nil {
		cfgInitOnce.Do(func() {
			l.disable = loggerCheckersFlag{
				newStringSet(l.config.cfg.Disable...),
			}
		})
	}

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
