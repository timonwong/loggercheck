package nilness

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ssa"
)

const src = `package p

type T struct{}

func (T) String() string { return "" }

func log(msg string, kv ...any) {}

func f() { log("m", "k", &T{}) }
`

func loadPass(t *testing.T) (*analysis.Pass, *ast.CallExpr) {
	t.Helper()
	pass, calls := loadSource(t, src)
	return pass, calls[0]
}

// loadSource type-checks a single-file package and returns its calls to log
// and log0 in source order.
func loadSource(t *testing.T, src string) (*analysis.Pass, []*ast.CallExpr) {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", src, 0)
	require.NoError(t, err)

	info := &types.Info{
		Types:      map[ast.Expr]types.TypeAndValue{},
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Implicits:  map[ast.Node]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
		Scopes:     map[ast.Node]*types.Scope{},
		Instances:  map[*ast.Ident]types.Instance{},
	}
	pkg, err := new(types.Config).Check("p", fset, []*ast.File{file}, info)
	require.NoError(t, err)

	var calls []*ast.CallExpr
	ast.Inspect(file, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && (id.Name == "log" || id.Name == "log0") {
				calls = append(calls, c)
			}
		}
		return true
	})
	require.NotEmpty(t, calls)

	return &analysis.Pass{Fset: fset, Files: []*ast.File{file}, Pkg: pkg, TypesInfo: info}, calls
}

func setTestHook(t *testing.T, hook func(phase string)) {
	t.Helper()
	testHook = hook
	t.Cleanup(func() { testHook = nil })
}

func TestNonNilVariadicArg(t *testing.T) {
	pass, call := loadPass(t)
	idx := NewLazyIndex(pass)

	assert.True(t, idx.NonNilVariadicArg(call, 2))
	assert.False(t, idx.NonNilVariadicArg(call, 1), "constant string is not a pointer")
}

func TestNilIndexIsUnknown(t *testing.T) {
	_, call := loadPass(t)

	var idx *Index
	assert.False(t, idx.NonNilVariadicArg(call, 2))
}

func TestUnsupportedArgumentsAreUnknown(t *testing.T) {
	pass, calls := loadSource(t, `package p

type T struct{}

func (T) String() string { return "" }

func log(msg string, kv ...any) {}

func log0(kv ...any) {}

func pair() (*T, *T) { return &T{}, &T{} }

func f(v any, xs []any) {
	log("m", "k", &T{})
	if v != nil {
		log("m", "k", v)
	}
	log("m", xs...)
	log0(pair())
}
`)
	require.Len(t, calls, 4)
	idx := NewLazyIndex(pass)

	assert.False(t, idx.NonNilVariadicArg(calls[0], 0), "fixed parameter")
	assert.False(t, idx.NonNilVariadicArg(calls[0], 3), "index past the arguments")
	assert.False(t, idx.NonNilVariadicArg(calls[1], 2), "interface-typed argument")
	assert.False(t, idx.NonNilVariadicArg(calls[2], 1), "explicit ...")
	assert.False(t, idx.NonNilVariadicArg(calls[3], 0), "multi-valued argument")
	assert.True(t, idx.NonNilVariadicArg(calls[0], 2), "unsupported queries must not disable the index")
}

func TestIndexCallsDropsAmbiguousPositions(t *testing.T) {
	pass, call := loadPass(t)

	// Two builds of one package yield distinct instructions for each call.
	build := func() *ssa.Function {
		prog := ssa.NewProgram(pass.Fset, 0)
		pkg := prog.CreatePackage(pass.Pkg, pass.Files, pass.TypesInfo, false)
		pkg.Build()
		return pkg.Func("f")
	}
	calls := indexCalls([]*ssa.Function{build(), build()})

	require.Contains(t, calls, call.Lparen)
	assert.Nil(t, calls[call.Lparen])
}

func TestBuildPanicIsUnknown(t *testing.T) {
	pass, call := loadPass(t)

	builds := 0
	setTestHook(t, func(phase string) {
		if phase == "build" {
			builds++
			panic("injected build panic")
		}
	})

	idx := NewLazyIndex(pass)
	assert.NotPanics(t, func() {
		assert.False(t, idx.NonNilVariadicArg(call, 2))
		assert.False(t, idx.NonNilVariadicArg(call, 2))
	})
	assert.Equal(t, 1, builds, "a failed build must not be retried")
}

func TestQueryPanicIsUnknown(t *testing.T) {
	pass, call := loadPass(t)

	panicking := true
	setTestHook(t, func(phase string) {
		if phase == "query" && panicking {
			panic("injected query panic")
		}
	})

	idx := NewLazyIndex(pass)
	assert.NotPanics(t, func() {
		assert.False(t, idx.NonNilVariadicArg(call, 2))
	})

	panicking = false
	assert.True(t, idx.NonNilVariadicArg(call, 2), "a query panic must not disable the index")
}
