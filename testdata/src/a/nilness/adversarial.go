package nilness

import (
	"fmt"
	stdlog "log"
	"os"
	"reflect"
	"runtime"
	"testing"

	userlog "a/nilness/user/log"
	useros "a/nilness/user/os"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
)

func adversarialORGuard(p *NamespacedName, cond bool) {
	if p == nil || cond {
		return
	}
	klog.InfoS("or guard", "p", p)
}

func adversarialReversedNilGuard(p *NamespacedName) {
	if nil != p {
		klog.InfoS("nil on left", "p", p)
	}
}

func adversarialRightPointerComparison(p, q *NamespacedName) {
	if q != p {
		klog.InfoS("right operand is not a nil check", "p", p) // want `logging value may panic when nil`
	}
}

func adversarialDerefThenPointerComparison(p, q *NamespacedName) {
	_ = *p
	if p != q {
		klog.InfoS("deref remains proof after unrelated comparison", "p", p)
	}
}

func adversarialReturningOSFunction(p *NamespacedName) {
	if p == nil {
		_ = os.Getpid()
	}
	klog.InfoS("os function returns", "p", p) // want `logging value may panic when nil`
}

func adversarialReturningRuntimeFunction(p *NamespacedName) {
	if p == nil {
		runtime.Gosched()
	}
	klog.InfoS("runtime function returns", "p", p) // want `logging value may panic when nil`
}

func adversarialDirectExitHook(p *NamespacedName) {
	old := klog.OsExit
	klog.OsExit = func(int) {}
	defer func() { klog.OsExit = old }()
	if p == nil {
		klog.OsExit(1)
	}
	klog.InfoS("function variable returns", "p", p) // want `logging value may panic when nil`
}

func adversarialANDGuard(p *NamespacedName, cond bool) {
	if cond && p != nil {
		klog.InfoS("and guard", "p", p)
	}
	klog.InfoS("after and", "p", p) // want `logging value may panic when nil`
}

func adversarialBooleanEquality(p *NamespacedName) {
	if (p != nil) == false {
		klog.InfoS("false nonnil", "p", p) // want `logging value may panic when nil`
		return
	}
	// Boolean comparison lowering is not recognized (conservative).
	klog.InfoS("true nonnil", "p", p) // want `logging value may panic when nil`
}

func adversarialExpressionSwitch(p *NamespacedName) {
	switch {
	case p == nil:
		klog.InfoS("nil case", "p", p) // want `logging value may panic when nil`
	default:
		klog.InfoS("default", "p", p)
	}
}

func adversarialValueSwitch(p *NamespacedName) {
	switch p {
	case nil:
		return
	}
	klog.InfoS("switch guard", "p", p)
}

func adversarialTypeSwitch(v any) {
	switch p := v.(type) {
	case *NamespacedName:
		klog.InfoS("typed nil possible", "p", p) // want `logging value may panic when nil`
		if p != nil {
			klog.InfoS("guarded assertion", "p", p)
		}
	}
}

func adversarialSelect(p *NamespacedName, ch <-chan bool) {
	select {
	case <-ch:
		if p == nil {
			return
		}
	default:
	}
	klog.InfoS("select bypass", "p", p) // want `logging value may panic when nil`
}

func adversarialGoto(p *NamespacedName, skip bool) {
	if skip {
		goto log
	}
	if p == nil {
		return
	}
log:
	klog.InfoS("goto bypass", "p", p) // want `logging value may panic when nil`
}

func adversarialLabeledBreak(p *NamespacedName, skip bool) {
outer:
	for {
		if skip {
			break outer
		}
		if p == nil {
			return
		}
		break
	}
	klog.InfoS("break bypass", "p", p) // want `logging value may panic when nil`
}

func adversarialLabeledContinue(p *NamespacedName, n int) {
outer:
	for i := 0; i < n; i++ {
		if i == 0 {
			continue outer
		}
		if p == nil {
			return
		}
		klog.InfoS("loop guard", "p", p)
	}
	klog.InfoS("zero iterations", "p", p) // want `logging value may panic when nil`
}

func adversarialShadow(p *NamespacedName, f func() *NamespacedName) {
	if p := f(); p != nil {
		klog.InfoS("inner", "p", p)
	}
	klog.InfoS("outer", "p", p) // want `logging value may panic when nil`
}

func adversarialAssignNil(p *NamespacedName) {
	if p == nil {
		return
	}
	p = nil
	klog.InfoS("reassigned", "p", p) // want `logging value may panic when nil`
}

func adversarialAssignUnknown(p *NamespacedName, f func() *NamespacedName) {
	if p != nil {
		p = f()
		klog.InfoS("new result", "p", p) // want `logging value may panic when nil`
	}
}

func adversarialClosureWrite() {
	p := &NamespacedName{}
	func() { p = nil }()
	klog.InfoS("closure write", "p", p) // want `logging value may panic when nil`
}

func adversarialAliasWrite() {
	p := &NamespacedName{}
	pp := &p
	*pp = nil
	klog.InfoS("alias write", "p", p) // want `logging value may panic when nil`
}

func adversarialDoublePointer(pp **NamespacedName) {
	if *pp != nil {
		*pp = nil
		klog.InfoS("pointee replaced", "p", *pp) // want `logging value may panic when nil`
	}
}

func adversarialGoroutineWrite() {
	p := &NamespacedName{}
	done := make(chan struct{})
	go func() { p = nil; close(done) }()
	<-done
	klog.InfoS("synchronized write", "p", p) // want `logging value may panic when nil`
}

func adversarialPointerComparison(p, q *NamespacedName) {
	if p != q {
		klog.InfoS("different pointers", "p", p) // want `logging value may panic when nil`
	}
}

func adversarialTypedNilInterface(p *NamespacedName) {
	var q *NamespacedName
	var v any = q
	if any(p) != v {
		klog.InfoS("interface comparison", "p", p) // want `logging value may panic when nil`
	}
}

func adversarialBoxedPointer(p *NamespacedName) {
	if any(p) != nil {
		klog.InfoS("boxed typed nil", "p", p) // want `logging value may panic when nil`
	}
}

func adversarialReflection(p *NamespacedName) {
	if !reflect.ValueOf(p).IsNil() {
		klog.InfoS("reflection guard", "p", p) // want `logging value may panic when nil`
	}
}

func adversarialNestedPhi(a, b, c bool) {
	p := &NamespacedName{}
	if a {
		if b {
			p = nil
		} else {
			p = &NamespacedName{}
		}
	} else if c {
		p = &NamespacedName{}
	}
	klog.InfoS("nested nil edge", "p", p) // want `logging value may panic when nil`
}

type adversarialLink struct{ next *adversarialLink }

func (adversarialLink) String() string { return "link" }

func adversarialPhiNilBackedge(head *adversarialLink, n int) {
	for p := head; n > 0; n-- {
		klog.InfoS("possibly nil traversal", "p", p) // want `logging value may panic when nil`
		p = p.next
	}
}

func adversarialPhiSelf(n int) {
	p := &NamespacedName{}
	for i := 0; i < n; i++ {
		if i == 1 {
			p = nil
		}
		klog.InfoS("nil backedge", "p", p) // want `logging value may panic when nil`
	}
}

func adversarialPhiMutual(n int) {
	p, q := &NamespacedName{}, (*NamespacedName)(nil)
	for i := 0; i < n; i++ {
		p, q = q, p
		klog.InfoS("mutual cycle", "p", p, "q", q) // want `logging value may panic when nil` `logging value may panic when nil`
	}
}

func adversarialPhiChangeType(cond bool) {
	var p *Time
	if cond {
		p = &Time{}
	}
	klog.InfoS("converted phi", "p", (*ConvertedTime)(p)) // want `logging value may panic when nil`
}

func adversarialDeferredDeref(p *NamespacedName) {
	defer func() { _ = p.Name }()
	klog.InfoS("before deferred deref", "p", p) // want `logging value may panic when nil`
}

func adversarialUncalledClosure(p *NamespacedName) func() string {
	f := func() string { return p.Name }
	klog.InfoS("closure not run", "p", p) // want `logging value may panic when nil`
	return f
}

func adversarialRecoveredDeref(p *NamespacedName) {
	func() {
		defer func() { _ = recover() }()
		_ = p.Name
	}()
	klog.InfoS("after recovered panic", "p", p) // want `logging value may panic when nil`
}

func adversarialDeferredLogAfterRecover(p *NamespacedName) {
	defer func() {
		_ = recover()
		klog.InfoS("recovered nil", "p", p) // want `logging value may panic when nil`
	}()
	_ = p.Name
}

func adversarialNonDominatingDeref(p *NamespacedName, cond bool) {
	if cond {
		_ = *p
	}
	klog.InfoS("bypass deref", "p", p) // want `logging value may panic when nil`
}

func adversarialOtherDeref(p, q *NamespacedName) {
	_ = *q
	klog.InfoS("other operand", "p", p, "q", q) // want `logging value may panic when nil`
}

func adversarialArrayLen(p *Pair) {
	_ = len(*p)
	klog.InfoS("constant array len", "p", p) // want `logging value may panic when nil`
}

func adversarialNilMap(p *NamespacedName, m map[string]int) {
	_ = m["key"]
	delete(m, "key")
	clear(m)
	klog.InfoS("nil map operations", "p", p) // want `logging value may panic when nil`
}

func adversarialNilSlice(p *NamespacedName, s []int) {
	_ = len(s)
	_ = cap(s)
	s = append(s, 1)
	clear(s)
	klog.InfoS("nil slice operations", "p", p) // want `logging value may panic when nil`
}

func (*NamespacedName) adversarialPointerMethod() {}

func adversarialPointerReceiver(p *NamespacedName) {
	p.adversarialPointerMethod()
	klog.InfoS("pointer receiver accepts nil", "p", p) // want `logging value may panic when nil`
}

func adversarialValueReceiver(p *NamespacedName) {
	_ = p.String()
	klog.InfoS("value receiver dereferences", "p", p)
}

func adversarialMethodExpression(p *NamespacedName) {
	_ = (*NamespacedName).String(p)
	klog.InfoS("method expression dereferences", "p", p) // want `logging value may panic when nil`
}

func adversarialGenericGuard[U any](p *NamespacedName, v U) {
	if p != nil {
		klog.InfoS("generic guard", "p", p, "v", v)
	}
	klog.InfoS("generic unguarded", "p", p) // want `logging value may panic when nil`
}

var _ = adversarialGenericGuard[int]

func adversarialRangeFunc(seq func(func(*NamespacedName) bool)) {
	for p := range seq {
		klog.InfoS("yielded nil possible", "p", p) // want `logging value may panic when nil`
		if p != nil {
			klog.InfoS("yielded guarded", "p", p)
		}
	}
}

func adversarialYieldClosure(p *NamespacedName) {
	seq := func(yield func(*NamespacedName) bool) {
		klog.InfoS("yield closure", "p", p) // want `logging value may panic when nil`
		yield(p)
	}
	for range seq {
	}
}

func adversarialIntegerRange(p *NamespacedName) {
	for i := range 10 {
		_ = max(min(i, 3), 0)
		klog.InfoS("integer range", "p", p) // want `logging value may panic when nil`
	}
}

func adversarialSameLine(p *NamespacedName) {
	adversarialSameLineGenerated(p)
}

func adversarialLineDirective(p *NamespacedName) {
//line generated.go:100
	klog.InfoS("generated safe", "p", &NamespacedName{})
//line generated.go:100
	klog.InfoS("generated unsafe", "p", p) // want `logging value may panic when nil`
//line adversarial.go:397
}

func adversarialGoAndDefer(p *NamespacedName) {
	defer klog.InfoS("deferred argument", "p", p) // want `logging value may panic when nil`
	go klog.InfoS("go argument", "p", p)          // want `logging value may panic when nil`
	if p != nil {
		defer klog.InfoS("deferred safe", "p", p)
		go klog.InfoS("go safe", "p", p)
	}
}

func adversarialChain(log logr.Logger, p *NamespacedName) {
	log.WithValues("safe", &NamespacedName{}).Info("chain", "p", p) // want `logging value may panic when nil`
	log.WithValues("p", p).Info("chain", "safe", &NamespacedName{}) // want `logging value may panic when nil`
}

func adversarialPanicRecover(p *NamespacedName) {
	defer func() {
		_ = recover()
		klog.InfoS("after recovered log panic", "p", p) // want `logging value may panic when nil`
	}()
	if p == nil {
		stdlog.Panic("nil")
	}
	// Capturing p forces a fresh heap-cell load after the guard (conservative).
	klog.InfoS("normal path", "p", p) // want `logging value may panic when nil`
}

func adversarialPanicRecoverNormalPath(p *NamespacedName) {
	defer func() { _ = recover() }()
	if p == nil {
		stdlog.Panic("nil")
	}
	klog.InfoS("recover returns from function", "p", p)
}

func adversarialFunctionValue(p *NamespacedName) {
	f := klog.Fatal
	if p == nil {
		f("nil")
	}
	// Indirect calls do not receive a no-return contract (conservative).
	klog.InfoS("function value guard", "p", p) // want `logging value may panic when nil`
}

func adversarialTestingInterface(tb testing.TB, p *NamespacedName) {
	if p == nil {
		tb.Fatal("nil")
	}
	klog.InfoS("interface guard", "p", p) // want `logging value may panic when nil`
}

type adversarialFataler struct{}

func (adversarialFataler) Fatal(...any) {}

func adversarialUnrelatedFatal(t adversarialFataler, p *NamespacedName) {
	if p == nil {
		t.Fatal("nil")
	}
	klog.InfoS("returning fatal method", "p", p) // want `logging value may panic when nil`
}

func adversarialUserLog(p *NamespacedName) {
	if p == nil {
		userlog.Fatal("nil")
	}
	klog.InfoS("user package log", "p", p) // want `logging value may panic when nil`
}

func adversarialUserOS(p *NamespacedName) {
	if p == nil {
		useros.Exit(1)
	}
	klog.InfoS("user package os", "p", p) // want `logging value may panic when nil`
}

var _ = func() bool {
	var p *NamespacedName
	klog.InfoS("package closure nil", "p", p) // want `logging value may panic when nil`
	return true
}()

func init() {
	var p *NamespacedName
	klog.InfoS("first init nil", "p", p) // want `logging value may panic when nil`
}

func init() {
	p := new(NamespacedName)
	klog.InfoS("second init safe", "p", p)
}

func adversarialNestedClosures(p *NamespacedName) {
	func() { func() { klog.InfoS("nested nil", "p", p) /* want `logging value may panic when nil` */ }() }()
}

var _ fmt.Stringer = (*NamespacedName)(nil)
