package loggercheck

import (
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"os"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/timonwong/loggercheck/internal/bytebufferpool"
	"github.com/timonwong/loggercheck/internal/rules"
	"github.com/timonwong/loggercheck/internal/sets"
)

const Doc = `Checks key valur pairs for common logger libraries (logr,klog,zap).`

func NewAnalyzer(opts ...Option) *analysis.Analyzer {
	l := newLoggerCheck(opts...)
	a := &analysis.Analyzer{
		Name:     "loggercheck",
		Doc:      Doc,
		Flags:    l.fs,
		Run:      l.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
	return a
}

type loggercheck struct {
	disable          sets.StringSet // flag -disable
	ruleFile         string         // flag -rulefile
	requireStringKey bool           // flag -requirestringkey
	fs               flag.FlagSet

	rules       []string        // used for external integration, for example golangci-lint
	rulesetList []rules.Ruleset // populate at runtime
}

func newLoggerCheck(opts ...Option) *loggercheck {
	l := &loggercheck{
		fs:          *flag.NewFlagSet("loggercheck", flag.ExitOnError),
		rulesetList: append([]rules.Ruleset{}, staticRuleList...), // ensure we make a clone of static rules first
	}

	l.fs.StringVar(&l.ruleFile, "rulefile", "", "path to a file contains a list of rules")
	l.fs.Var(&l.disable, "disable", "comma-separated list of disabled logger checker (klog,logr,zap)")
	l.fs.BoolVar(&l.requireStringKey, "requirestringkey", false, "require all logging keys to be inlined literal strings")

	for _, opt := range opts {
		opt(l)
	}

	return l
}

func (l *loggercheck) isCheckerDisabled(name string) bool {
	return l.disable.Has(name)
}

func (l *loggercheck) isValidLoggerFunc(fn *types.Func) bool {
	pkg := fn.Pkg()
	if pkg == nil {
		return false
	}

	for i := range l.rulesetList {
		rs := &l.rulesetList[i]
		if l.isCheckerDisabled(rs.Name) {
			// Skip ignored logger checker.
			continue
		}

		if rs.Match(fn, pkg) {
			return true
		}
	}

	return false
}

// isArgTypeOfString returns true if the argument is string literal or string constant.
func isArgTypeOfString(pass *analysis.Pass, arg ast.Expr) bool {
	switch arg := arg.(type) {
	case *ast.BasicLit: // literals, must be string
		if arg.Kind == token.STRING {
			return true
		}
	case *ast.Ident: // identifiers, we require constant string key
		if arg.Obj != nil && arg.Obj.Kind == ast.Con {
			typ := pass.TypesInfo.Types[arg].Type
			if typ, ok := typ.(*types.Basic); ok {
				if typ.Kind() == types.String {
					return true
				}
			}
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

	// Check the "key" type
	if l.requireStringKey {
		for i := startIndex; i < nargs; i += 2 {
			arg := call.Args[i]
			if isArgTypeOfString(pass, arg) {
				continue
			}

			pass.Report(analysis.Diagnostic{
				Pos:      arg.Pos(),
				End:      arg.End(),
				Category: "logging",
				Message: fmt.Sprintf(
					"logging key are expected to be inlined constant strings, please replace %q provided with string",
					renderNode(pass.Fset, arg)),
			})
		}
	}
}

func renderNode(fset *token.FileSet, v interface{}) string {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	_ = printer.Fprint(buf, fset, v) //nolint:errcheck
	return buf.String()
}

func (l *loggercheck) processConfig() error {
	if l.ruleFile != "" { // flags takes precedence over configs
		f, err := os.Open(l.ruleFile)
		if err != nil {
			return fmt.Errorf("failed to open rule file: %w", err)
		}
		defer f.Close()

		custom, err := rules.ParseRuleFile(f)
		if err != nil {
			return fmt.Errorf("failed to parse rule file: %w", err)
		}
		l.rulesetList = append(l.rulesetList, custom...)
	} else if len(l.rules) > 0 {
		custom, err := rules.ParseRules(l.rules)
		if err != nil {
			return fmt.Errorf("failed to parse rules: %w", err)
		}
		l.rulesetList = append(l.rulesetList, custom...)
	}

	return nil
}

func (l *loggercheck) run(pass *analysis.Pass) (interface{}, error) {
	err := l.processConfig()
	if err != nil {
		return nil, err
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
