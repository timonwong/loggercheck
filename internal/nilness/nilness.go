// Package nilness answers whether a pointer passed as a variadic logging
// argument is provably non-nil at the call site, using an SSA form of the
// package that is built on first use.
//
// The dominating-condition logic is adapted from
// golang.org/x/tools/go/analysis/passes/nilness (BSD license), queried on
// demand for a single value instead of computed for whole functions.
package nilness

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ssa"
)

// testHook, when set by tests, runs at the start of the SSA build ("build")
// and of each query ("query").
var testHook func(phase string)

// Index maps call expressions to their SSA call instructions.
type Index struct {
	pass  *analysis.Pass
	built bool
	calls map[token.Pos]ssa.CallInstruction // nil when the build failed
}

// NewLazyIndex returns an Index that builds SSA for the package on the first
// query, so packages without candidate arguments pay nothing. It does not use
// buildssa, which would require ctrlflow for every package.
func NewLazyIndex(pass *analysis.Pass) *Index {
	return &Index{pass: pass}
}

// NonNilVariadicArg reports whether call.Args[argIndex], which must belong to
// the variadic part of a call without "...", is provably non-nil.
func (x *Index) NonNilVariadicArg(call *ast.CallExpr, argIndex int) (nonNil bool) {
	if x == nil {
		return false
	}
	if !x.built {
		x.built = true
		x.calls = buildCalls(x.pass)
	}
	if x.calls == nil {
		return false
	}

	defer func() {
		if recover() != nil {
			nonNil = false
		}
	}()
	if testHook != nil {
		testHook("query")
	}

	instr := x.calls[call.Lparen]
	if instr == nil {
		return false // e.g. unreachable code removed by the SSA builder
	}
	v := variadicValue(instr, len(call.Args), argIndex)
	if v == nil {
		return false
	}
	block, idx := instrPoint(instr)
	if block == nil {
		return false
	}
	q := query{visiting: make(map[*ssa.Phi]bool)}
	return q.nonNilAt(v, block, idx, nil)
}

// buildCalls returns nil if the SSA builder panics, which it may do on
// inputs it does not support; the caller then reports as if nothing is known.
func buildCalls(pass *analysis.Pass) (calls map[token.Pos]ssa.CallInstruction) {
	defer func() {
		if recover() != nil {
			calls = nil
		}
	}()
	if testHook != nil {
		testHook("build")
	}

	prog := ssa.NewProgram(pass.Fset, 0)
	prog.SetNoReturn(noReturn)
	for _, p := range pass.Pkg.Imports() {
		prog.CreatePackage(p, nil, nil, true)
	}
	pkg := prog.CreatePackage(pass.Pkg, pass.Files, pass.TypesInfo, false)
	pkg.Build()

	var fns []*ssa.Function
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok {
				if fn := prog.FuncValue(pass.TypesInfo.Defs[fd.Name].(*types.Func)); fn != nil {
					fns = append(fns, fn)
				}
			}
		}
	}
	// Package-level variable initializers (and closures inside them) live in
	// the synthetic init function, which has no declaration.
	if initFn := pkg.Func("init"); initFn != nil {
		fns = append(fns, initFn)
	}
	return indexCalls(fns)
}

func indexCalls(fns []*ssa.Function) map[token.Pos]ssa.CallInstruction {
	calls := make(map[token.Pos]ssa.CallInstruction)
	var add func(fn *ssa.Function)
	add = func(fn *ssa.Function) {
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				ci, ok := instr.(ssa.CallInstruction)
				if !ok {
					continue
				}
				pos := ci.Common().Pos()
				if !pos.IsValid() {
					continue
				}
				if prev, dup := calls[pos]; dup && prev != ci {
					calls[pos] = nil // ambiguous: never trust it
					continue
				}
				calls[pos] = ci
			}
		}
		for _, anon := range fn.AnonFuncs {
			add(anon)
		}
	}
	for _, fn := range fns {
		add(fn)
	}
	return calls
}

// noReturn lets the SSA builder end the block after calls that never return,
// so a nil branch guarded by e.g. klog.Fatal does not flow into the code after
// it. go/ssa only consults it for static callees: calls through an interface
// such as testing.TB are not covered. Promoted methods like (*testing.T).Fatal
// arrive as their declaring method, (*testing.common).Fatal.
func noReturn(fn *types.Func) bool {
	if fn.Pkg() == nil {
		return false
	}
	path := fn.Pkg().Path()
	// GOPATH-mode vendoring prefixes the import path, e.g. "a/vendor/k8s.io/klog/v2".
	if i := strings.LastIndex(path, "/vendor/"); i >= 0 {
		path = path[i+len("/vendor/"):]
	}
	name := fn.Name()
	isMethod := fn.Signature().Recv() != nil
	switch path {
	case "os":
		return !isMethod && name == "Exit"
	case "runtime":
		return !isMethod && name == "Goexit"
	case "log":
		return strings.HasPrefix(name, "Fatal") || strings.HasPrefix(name, "Panic")
	case "testing":
		switch name {
		case "Fatal", "Fatalf", "FailNow", "Skip", "Skipf", "SkipNow":
			return isMethod
		}
	case "k8s.io/klog/v2":
		return !isMethod &&
			(strings.HasPrefix(name, "Fatal") || strings.HasPrefix(name, "Exit") || name == "FlushAndExit")
	}
	return false
}

// variadicValue recovers the operand stored into the implicit varargs array.
// go/ssa lowers f(a, b...) to
//
//	t0 = new [n]any (varargs)
//	t1 = &t0[i]; store MakeInterface(arg_i) -> t1
//	t2 = slice t0[:]
//
// and gives the stores the operand's own position rather than the syntax
// position, so arguments are matched by index, not by position.
func variadicValue(instr ssa.CallInstruction, nSyntaxArgs, argIndex int) ssa.Value {
	common := instr.Common()
	if len(common.Args) == 0 {
		return nil
	}
	sl, ok := common.Args[len(common.Args)-1].(*ssa.Slice)
	if !ok {
		return nil
	}
	arr, ok := sl.X.(*ssa.Alloc)
	if !ok {
		return nil
	}
	ptr, ok := arr.Type().Underlying().(*types.Pointer)
	if !ok {
		return nil
	}
	at, ok := ptr.Elem().Underlying().(*types.Array)
	if !ok {
		return nil
	}
	nFixed := common.Signature().Params().Len() - 1
	if nFixed < 0 || at.Len() != int64(nSyntaxArgs-nFixed) {
		return nil // multi-valued argument expanded into the varargs
	}
	want := int64(argIndex - nFixed)
	if want < 0 || want >= at.Len() {
		return nil
	}
	return storedOperand(arr, want)
}

func storedOperand(arr *ssa.Alloc, index int64) ssa.Value {
	for _, ref := range *arr.Referrers() {
		ia, ok := ref.(*ssa.IndexAddr)
		if !ok || ia.X != arr {
			continue
		}
		c, ok := ia.Index.(*ssa.Const)
		if !ok || c.Int64() != index {
			continue
		}
		for _, ref2 := range *ia.Referrers() {
			st, ok := ref2.(*ssa.Store)
			if !ok || st.Addr != ia {
				continue
			}
			if mi, ok := st.Val.(*ssa.MakeInterface); ok {
				return mi.X
			}
			return nil
		}
	}
	return nil
}

func instrPoint(instr ssa.Instruction) (block *ssa.BasicBlock, idx int) {
	b := instr.Block()
	if b == nil {
		return nil, 0
	}
	for i, in := range b.Instrs {
		if in == instr {
			return b, i
		}
	}
	return nil, 0
}

type query struct {
	visiting map[*ssa.Phi]bool
}

// nonNilAt reports whether v is provably non-nil just before b.Instrs[idx].
// When edgeTo is non-nil, the point is the end of b on the edge b->edgeTo.
func (q *query) nonNilAt(v ssa.Value, b *ssa.BasicBlock, idx int, edgeTo *ssa.BasicBlock) bool {
	switch v := v.(type) {
	case *ssa.Alloc, *ssa.FieldAddr, *ssa.IndexAddr, *ssa.Global,
		*ssa.Function, *ssa.MakeClosure:
		return true
	case *ssa.Const:
		return false
	case *ssa.Phi:
		if q.phi(v) {
			return true
		}
	}

	if conditionFact(v, b, edgeTo) {
		return true
	}
	if dereferencedBefore(v, b, idx) {
		return true
	}
	if ct, ok := v.(*ssa.ChangeType); ok {
		return q.nonNilAt(ct.X, b, idx, edgeTo)
	}
	return false
}

func (q *query) phi(p *ssa.Phi) bool {
	if q.visiting[p] {
		// A phi cycle only copies values that entered through its other
		// edges, so assuming non-nil here is sound once those are checked.
		return true
	}
	q.visiting[p] = true
	defer delete(q.visiting, p)

	preds := p.Block().Preds
	for i, edge := range p.Edges {
		pred := preds[i]
		if !q.nonNilAt(edge, pred, len(pred.Instrs), p.Block()) {
			return false
		}
	}
	return true
}

// conditionFact looks for a "v != nil" / "v == nil" branch that dominates the
// point and proves v non-nil. Like the nilness pass, a branch only
// contributes a fact to a successor whose sole predecessor is the branching
// block.
func conditionFact(v ssa.Value, b, edgeTo *ssa.BasicBlock) bool {
	if edgeTo != nil && branchProvesNonNil(v, b, edgeTo) {
		return true
	}
	for child := b; child != nil; child = child.Idom() {
		if len(child.Preds) != 1 {
			continue
		}
		if branchProvesNonNil(v, child.Preds[0], child) {
			return true
		}
	}
	return false
}

// branchProvesNonNil reports whether taking the edge from->to implies that v
// is non-nil.
func branchProvesNonNil(v ssa.Value, from, to *ssa.BasicBlock) bool {
	if len(from.Instrs) == 0 || len(from.Succs) != 2 || from.Succs[0] == from.Succs[1] {
		return false
	}
	ifInstr, ok := from.Instrs[len(from.Instrs)-1].(*ssa.If)
	if !ok {
		return false
	}
	binop, ok := ifInstr.Cond.(*ssa.BinOp)
	if !ok || (binop.Op != token.EQL && binop.Op != token.NEQ) {
		return false
	}
	isNilCheck := (binop.X == v && isNilConst(binop.Y)) || (binop.Y == v && isNilConst(binop.X))
	if !isNilCheck {
		return false
	}
	// nonNilSucc is the successor taken when v != nil.
	nonNilSucc := from.Succs[1]
	if binop.Op == token.NEQ {
		nonNilSucc = from.Succs[0]
	}
	return to == nonNilSucc
}

func isNilConst(v ssa.Value) bool {
	c, ok := v.(*ssa.Const)
	return ok && c.IsNil()
}

// dereferencedBefore reports whether an instruction that panics on a nil v
// dominates the point, so v cannot be nil once the point is reached.
func dereferencedBefore(v ssa.Value, b *ssa.BasicBlock, idx int) bool {
	refs := v.Referrers()
	if refs == nil {
		return false
	}
	for _, ref := range *refs {
		if !derefs(ref, v) {
			continue
		}
		rb := ref.Block()
		if rb == b {
			for i := 0; i < idx && i < len(b.Instrs); i++ {
				if b.Instrs[i] == ref {
					return true
				}
			}
			continue
		}
		if rb != nil && rb.Dominates(b) {
			return true
		}
	}
	return false
}

func derefs(instr ssa.Instruction, v ssa.Value) bool {
	switch instr := instr.(type) {
	case *ssa.UnOp:
		return instr.Op == token.MUL && instr.X == v
	case *ssa.FieldAddr:
		return instr.X == v
	case *ssa.Store:
		return instr.Addr == v
	case *ssa.IndexAddr:
		_, isPtr := instr.X.Type().Underlying().(*types.Pointer)
		return isPtr && instr.X == v
	}
	return false
}
