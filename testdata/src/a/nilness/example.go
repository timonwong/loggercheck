package nilness

import (
	"errors"
	stdlog "log"
	"os"
	"runtime"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
)

// Time mirrors metav1.Time: a struct whose String method is promoted from an
// embedded value, so a nil *Time panics when formatted.
type Time struct {
	time.Time
}

// NamespacedName mirrors types.NamespacedName (value-receiver String).
type NamespacedName struct {
	Namespace string
	Name      string
}

func (n NamespacedName) String() string { return n.Namespace + "/" + n.Name }

type CertificateStatus struct {
	NotAfter    *Time
	RenewalTime *Time
}

type Certificate struct {
	Status CertificateStatus
}

type holder struct {
	Name NamespacedName
}

var errTest = errors.New("test")

// cert-manager true positive 1: pointer loaded from a struct field.
func certManagerFieldLoad(log logr.Logger, crt *Certificate, reset bool) {
	if reset {
		crt.Status.RenewalTime = nil
	}
	log.V(1).Info("updating status fields", "notAfter", crt.Status.NotAfter, "renewalTime", crt.Status.RenewalTime) // want `logging value may panic when nil because its element type implements fmt.Stringer` `logging value may panic when nil because its element type implements fmt.Stringer`
	klog.InfoS("updating status fields", "notAfter", crt.Status.NotAfter) // want `logging value may panic when nil because its element type implements fmt.Stringer`
}

func owningCertForSecret(name string) *NamespacedName {
	if name == "" {
		return nil
	}
	return &NamespacedName{Name: name}
}

// cert-manager true positive 2: reachable both when owner is nil and when it is not.
func certManagerOwner(log logr.Logger, secretName string, certName NamespacedName) {
	owner := owningCertForSecret(secretName)
	if owner == nil || *owner != certName {
		log.V(1).Info("refusing to issue", "annotationValue", owner, "expected", certName) // want `logging value may panic when nil because its element type implements fmt.Stringer`
		return
	}
}

// cert-manager false positive: pointer constructed immediately before logging.
func certManagerConstructed(log logr.Logger, ns, name string) {
	key := &NamespacedName{Namespace: ns, Name: name}
	log.Info("constructed", "key", key)
	klog.InfoS("constructed", "key", key)
}

func compositeLiteralInline(log logr.Logger) {
	log.Info("inline", "key", &NamespacedName{Name: "x"})
}

func newBuiltin(log logr.Logger) {
	t := new(Time)
	log.Info("new", "time", t)
}

func addressOfLocal(log logr.Logger) {
	var t Time
	log.Info("address of local", "time", &t)
}

func addressOfField(log logr.Logger, h *holder) {
	log.Info("address of field", "name", &h.Name)
}

func parameter(log logr.Logger, p *Time) {
	log.Info("parameter", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
}

func guardedNonNil(log logr.Logger, p *Time) {
	if p != nil {
		log.Info("guarded", "time", p)
	}
}

func earlyReturnOnNil(log logr.Logger, p *Time) {
	if p == nil {
		return
	}
	log.Info("early return", "time", p)
}

func guardedNil(log logr.Logger, p *Time) {
	if p == nil {
		log.Info("nil branch", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
	}
}

func nilLiteral(log logr.Logger) {
	var p *Time
	log.Info("zero value", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
}

func panicOnNil(log logr.Logger, p *Time) {
	if p == nil {
		panic("nil")
	}
	log.Info("after panic guard", "time", p)
}

func exitOnNil(log logr.Logger, p *Time) {
	if p == nil {
		os.Exit(1)
	}
	log.Info("after exit guard", "time", p)
}

func goexitOnNil(log logr.Logger, p *Time) {
	if p == nil {
		runtime.Goexit()
	}
	log.Info("after goexit guard", "time", p)
}

func klogFatalOnNil(p *Time) {
	if p == nil {
		klog.Fatal("missing time")
	}
	klog.InfoS("after klog.Fatal guard", "time", p)
}

func klogFatalfOnNil(p *Time) {
	if p == nil {
		klog.Fatalf("missing %s", "time")
	}
	klog.InfoS("after klog.Fatalf guard", "time", p)
}

func klogExitOnNil(p *Time) {
	if p == nil {
		klog.Exit("missing time")
	}
	klog.InfoS("after klog.Exit guard", "time", p)
}

func klogFlushAndExitOnNil(p *Time) {
	if p == nil {
		klog.FlushAndExit(time.Second, 1)
	}
	klog.InfoS("after klog.FlushAndExit guard", "time", p)
}

func stdlogFatalOnNil(log logr.Logger, p *Time) {
	if p == nil {
		stdlog.Fatal("missing time")
	}
	log.Info("after log.Fatal guard", "time", p)
}

func stdlogPanicfOnNil(log logr.Logger, p *Time) {
	if p == nil {
		stdlog.Panicf("missing %s", "time")
	}
	log.Info("after log.Panicf guard", "time", p)
}

func stdlogLoggerFatalOnNil(log logr.Logger, logger *stdlog.Logger, p *Time) {
	if p == nil {
		logger.Fatalln("missing time")
	}
	log.Info("after (*log.Logger).Fatalln guard", "time", p)
}

func die() { os.Exit(1) }

// Only calls to a fixed set of library functions end the nil branch; a local
// helper that never returns is not analyzed (remaining false positive).
func localNoReturnHelperOnNil(log logr.Logger, p *Time) {
	if p == nil {
		die()
	}
	log.Info("after local helper guard", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
}

func nonExitingCallOnNil(p *Time) {
	if p == nil {
		klog.Error("missing time")
	}
	klog.InfoS("after klog.Error", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
}

// Each read of a struct field is a separate SSA load, so a nil check on one
// read says nothing about the next (remaining false positive).
func guardedFieldLoad(log logr.Logger, crt *Certificate) {
	if crt.Status.NotAfter != nil {
		log.Info("guarded field", "notAfter", crt.Status.NotAfter) // want `logging value may panic when nil because its element type implements fmt.Stringer`
	}
	if notAfter := crt.Status.NotAfter; notAfter != nil {
		log.Info("guarded copy", "notAfter", notAfter)
	}
}

func shortCircuitGuard(log logr.Logger, p *Time, skip bool) {
	if p == nil || skip {
		return
	}
	log.Info("after short-circuit guard", "time", p)
}

func lookup(name string) (*NamespacedName, error) {
	if name == "" {
		return nil, errTest
	}
	return &NamespacedName{Name: name}, nil
}

// Non-nilness implied by a sibling error result needs interprocedural
// knowledge (remaining false positive).
func errGuard(log logr.Logger, name string) {
	n, err := lookup(name)
	if err != nil {
		return
	}
	log.Info("found", "name", n) // want `logging value may panic when nil because its element type implements fmt.Stringer`
}

func derefBeforeLog(log logr.Logger, p *NamespacedName) {
	_ = p.Name
	log.Info("after field read", "name", p)
}

func loadBeforeLog(log logr.Logger, p *NamespacedName) {
	v := *p
	_ = v
	log.Info("after load", "name", p)
}

func derefNotDominating(log logr.Logger, p *NamespacedName, cond bool) {
	if cond {
		_ = p.Name
	}
	log.Info("deref does not dominate", "name", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
}

func derefAfterLog(log logr.Logger, p *NamespacedName) string {
	log.Info("deref after log", "name", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
	return p.Name
}

func phiOfAllocs(log logr.Logger, cond bool) {
	var p *NamespacedName
	if cond {
		p = &NamespacedName{Name: "a"}
	} else {
		p = &NamespacedName{Name: "b"}
	}
	log.Info("phi", "name", p)
}

func phiWithNilEdge(log logr.Logger, cond bool) {
	var p *NamespacedName
	if cond {
		p = &NamespacedName{Name: "a"}
	}
	log.Info("phi with nil edge", "name", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
}

func phiDefaulted(log logr.Logger, p *NamespacedName) {
	if p == nil {
		p = &NamespacedName{Name: "default"}
	}
	log.Info("defaulted", "name", p)
}

func phiLoop(log logr.Logger, xs []*NamespacedName) {
	p := &NamespacedName{Name: "first"}
	for _, x := range xs {
		log.Info("loop", "name", p)
		if x != nil {
			p = x
		}
	}
}

func phiLoopUnguarded(log logr.Logger, xs []*NamespacedName) {
	p := &NamespacedName{Name: "first"}
	for _, x := range xs {
		log.Info("loop", "name", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
		p = x
	}
}

func funcLiteral(log logr.Logger) {
	func() {
		p := &NamespacedName{Name: "x"}
		log.Info("closure", "name", p)
	}()

	f := func(p *NamespacedName) {
		log.Info("closure parameter", "name", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
	}
	f(nil)
}

// A variable captured by a closure lives in a heap cell, so each use is a
// separate load and is not provably non-nil (remaining false positive).
func capturedVariable(log logr.Logger) func() string {
	p := &NamespacedName{Name: "x"}
	log.Info("captured", "name", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
	return func() string { return p.Name }
}

func verbosityAndVariants(log logr.Logger, p *Time) {
	log.V(1).Info("v", "time", new(Time))
	log.V(1).Info("v", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`

	l2 := log.WithValues("time", &Time{})
	l2 = l2.WithValues("time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`

	l2.Error(errTest, "error", "time", &Time{})
	l2.Error(errTest, "error", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
}

func klogVariants(p *Time) {
	klog.InfoS("klog", "time", &Time{})
	klog.InfoS("klog", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
	klog.V(2).InfoS("klog", "time", new(Time))
	klog.V(2).InfoS("klog", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
	klog.ErrorS(errTest, "klog", "time", &Time{})
	klog.ErrorS(errTest, "klog", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
	if p != nil {
		klog.V(2).InfoS("klog", "time", p)
	}
}

func deferredAndGo(log logr.Logger, p *Time) {
	defer log.Info("deferred", "time", &Time{})
	defer log.Info("deferred", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
	go log.Info("go", "time", new(Time))
}

func mixedPositions(log logr.Logger, p *Time) {
	q := &Time{}
	log.Info("mixed", "a", p, "b", q, "c", p, "d", q) // want `logging value may panic when nil because its element type implements fmt.Stringer` `logging value may panic when nil because its element type implements fmt.Stringer`
}

func generic[T any](log logr.Logger, p *Time, v T) {
	log.Info("generic", "time", &Time{}, "v", v)
	log.Info("generic", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
}

var _ = generic[int]

var globalTime *Time

// Package-level initializers are built into the synthetic package init function.
var _ = func() bool {
	klog.InfoS("init", "time", &Time{})
	klog.InfoS("init", "time", globalTime) // want `logging value may panic when nil because its element type implements fmt.Stringer`
	return true
}()

var initLogger = logr.Discard().WithValues("time", &Time{})

func _(log logr.Logger) {
	log.Info("blank function", "key", &NamespacedName{})
}
