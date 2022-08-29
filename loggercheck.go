package loggercheck

import (
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"os"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/timonwong/loggercheck/internal/rules"
	"github.com/timonwong/loggercheck/internal/sets"
)

const Doc = `Checks key valur pairs for common logger libraries (logr,klog,zap).`

func NewAnalyzer(opts ...Option) *analysis.Analyzer {
	l := &loggercheck{
		rulesetList: append([]rules.Ruleset{}, staticRuleList...), // ensure we make a clone of static rules first
	}
	for _, o := range opts {
		o(l)
	}

	a := &analysis.Analyzer{
		Name:     "loggercheck",
		Doc:      Doc,
		Run:      l.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	a.Flags.Init("loggercheck", flag.ExitOnError)
	a.Flags.StringVar(&l.ruleFile, "rulefile", "", "path to a file contains a list of rules")
	a.Flags.Var(&l.disable, "disable", "comma-separated list of disabled logger checker (klog,logr,zap)")
	return a
}

type loggercheck struct {
	disable  sets.StringSet // flag -disable
	ruleFile string         // flag -rulefile

	rules       []string        // used for external integration, for example golangci-lint
	rulesetList []rules.Ruleset // populate at runtime
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
