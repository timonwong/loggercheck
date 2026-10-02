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
)

const src = `package p

type T struct{}

func (T) String() string { return "" }

func log(msg string, kv ...any) {}

func f() { log("m", "k", &T{}) }
`

func loadPass(t *testing.T) (*analysis.Pass, *ast.CallExpr) {
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

	var call *ast.CallExpr
	ast.Inspect(file, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			call = c
		}
		return call == nil
	})
	require.NotNil(t, call)

	return &analysis.Pass{Fset: fset, Files: []*ast.File{file}, Pkg: pkg, TypesInfo: info}, call
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
